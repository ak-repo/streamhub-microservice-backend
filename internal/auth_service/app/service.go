package app

import (
	"context"
	"log"
	"time"

	"github.com/ak-repo/stream-hub/config"
	authcloudinary "github.com/ak-repo/stream-hub/internal/auth_service/adapter/cloudinary"
	otpredis "github.com/ak-repo/stream-hub/internal/auth_service/adapter/redis"
	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/ak-repo/stream-hub/pkg/logger"
	"github.com/ak-repo/stream-hub/pkg/utils"
	"github.com/google/uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	ErrUserExists   = "user already exists"
	ErrInvalidCreds = "invalid credentials"
	ErrUserNotFound = "user not found"
)

type authService struct {
	repo      port.UserRepository
	jwt       *jwt.JWTManager
	cfg       *config.Config
	otpStore  *otpredis.OTPStore
	cloudzcli *authcloudinary.CloudinaryUploader
}

func NewAuthService(repo port.UserRepository, jwtMgr *jwt.JWTManager, cfg *config.Config, otpStore *otpredis.OTPStore, cloudzcli *authcloudinary.CloudinaryUploader) port.AuthService {
	return &authService{repo: repo, jwt: jwtMgr, cfg: cfg, otpStore: otpStore, cloudzcli: cloudzcli}
}

// -------------------- REGISTER --------------------
func (s *authService) Register(ctx context.Context, email, username, password string) error {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return errors.New(errors.CodeConflict, ErrUserExists, nil)
	}

	hashed, _ := utils.HashPassword(password)

	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Username:     username,
		PasswordHash: hashed,
		Role:         "user",
		Avatar_url:   "https://res.cloudinary.com/dersnukrf/image/upload/v1764929207/avatars/avatars/profile.jpg.webp",
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return errors.New(errors.CodeInternal, "database saving failed", err)
	}

	return nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, ErrInvalidCreds, err)
	}
	if user == nil {
		return nil, errors.New(errors.CodeNotFound, ErrInvalidCreds, err)
	}

	if !utils.ComparePassword(user.PasswordHash, password) {
		return nil, errors.New(errors.CodeConflict, ErrInvalidCreds, err)
	}

	log.Println("user: ", user.Username, " is banned: ", user.IsBanned)
	if user.IsBanned {
		return nil, errors.New(errors.CodeForbidden, "email is banned", nil)
	}

	return user, nil
}

// --------------------------Magic Link --------------------------------
func (s *authService) SendMagicLink(email string) (string, string, error) {
	token, exp, err := s.jwt.GenerateAccessToken("0", email, "magic-link")
	if err != nil {
		return "", "", errors.New(errors.CodeInternal, "token generation failed", err)
	}

	magicLink := "http://" + s.cfg.Services.Front.Host + ":" +
		s.cfg.Services.Front.Port +
		"/verify-link?email=" + email +
		"&token=" + token

	// SendGrid
	from := mail.NewEmail("StreamHub", "ak506lap@gmail.com")
	to := mail.NewEmail("", email)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "Your Magic Login Link"
	message.SetTemplateID(s.cfg.SendGrid.MagicTemplate)

	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("magic_link", magicLink)
	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.cfg.SendGrid.Key)
	response, err := client.Send(message)

	if err != nil || response.StatusCode != 202 {
		return "", "", errors.New(errors.CodeInternal, "failed to send verify link", err)
	}

	return magicLink, helper.TimeToString(exp), nil

}

func (s *authService) VerifyMagicLink(ctx context.Context, token, email string) error {

	claims, err := s.jwt.ValidateToken(token)
	if err != nil {
		return errors.New(errors.CodeUnauthorized, "link verification failed", err)
	}

	if claims.Email != email {
		return errors.New(errors.CodeUnauthorized, "token email mismatch", nil)
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}
	user.EmailVerified = true

	if err := s.repo.Update(ctx, user); err != nil {
		return errors.New(errors.CodeInternal, "database update error", err)
	}

	return nil
}

func (s *authService) FindUser(ctx context.Context, identifier, method string) (*domain.User, error) {

	switch method {
	case "id":
		user, err := s.repo.FindByID(ctx, identifier)
		if err != nil {
			return nil, errors.New(errors.CodeNotFound, "user not found by this ID: "+identifier, err)
		}
		return user, nil
	case "email":
		user, err := s.repo.FindByEmail(ctx, identifier)
		if err != nil {
			return nil, errors.New(errors.CodeNotFound, "user not found by this email: "+identifier, err)
		}
		return user, nil
	case "username":
		user, err := s.repo.FindByUsername(ctx, identifier)
		if err != nil {
			return nil, errors.New(errors.CodeNotFound, "user not found by this email: "+identifier, err)
		}
		return user, nil
	}
	return nil, errors.New(errors.CodeNotFound, "user not found method not allowed", nil)
}

// -----------------------------Password REset--------------

func (s *authService) PasswordReset(ctx context.Context, email string) error {

	// Check if user exists
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}

	// Generate OTP
	otp, err := utils.GenerateOTP(6)
	if err != nil {
		return errors.New(errors.CodeInternal, "otp generation failed", err)
	}

	// Save OTP in Redis (with TTL configured inside otpStore)
	if err := s.otpStore.SaveOTP(ctx, email, otp); err != nil {
		return errors.New(errors.CodeInternal, "failed to store otp", err)
	}

	// -------- Send OTP Email using SendGrid --------

	from := mail.NewEmail("StreamHub", "ak506lap@gmail.com")
	to := mail.NewEmail("", email)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "Your Password Reset Code"
	message.SetTemplateID(s.cfg.SendGrid.OTPTemplate)

	p := mail.NewPersonalization()
	p.AddTos(to)

	// dynamic template fields in SendGrid
	p.SetDynamicTemplateData("otp_code", otp)
	p.SetDynamicTemplateData("username", user.Username)
	p.SetDynamicTemplateData("support_email", "support@streamhub.com")

	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.cfg.SendGrid.Key)
	resp, err := client.Send(message)

	if err != nil || resp.StatusCode != 202 {
		return errors.New(errors.CodeInternal, "failed to send otp email", err)
	}

	return nil
}

func (s *authService) VerifyPasswordReset(ctx context.Context, otp, password, email string) error {

	// redis check for ott
	rOTP, err := s.otpStore.VerifyOTP(ctx, email)
	if err != nil {
		return errors.New(errors.CodeInternal, "failed identify otp", err)
	}
	if rOTP != otp {
		return errors.New(errors.CodeForbidden, "entered otp not matching", nil)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return errors.New(errors.CodeInternal, "failed to hash password", err)
	}
	if err := s.repo.UpdatePassword(ctx, email, hash); err != nil {
		return errors.New(errors.CodeInternal, "failed to save new password", err)
	}

	// delete from redis
	if err := s.otpStore.DeleteOTP(ctx, email); err != nil {
		return errors.New(errors.CodeInternal, "delete from redis failed", err)
	}

	return nil

}

//-----------profile update----------

func (s *authService) UpdateProfile(ctx context.Context, userID, username, email string) (*domain.User, error) {

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}

	if email != "" && email != user.Email {
		u, _ := s.repo.FindByEmail(ctx, email)
		if u != nil && u.ID != userID {
			return nil, errors.New(errors.CodeAlreadyExists, "email already in use", nil)
		}
		user.Email = email
	}

	if username != "" && username != user.Username {
		user.Username = username
	}

	// update
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to update user details", err)
	}

	return user, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID, password, newPassword string) error {

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}

	if !utils.ComparePassword(user.PasswordHash, password) {
		return errors.New(errors.CodeUnauthorized, "enter password not matching", nil)
	}
	hash, _ := utils.HashPassword(newPassword)
	if err := s.repo.UpdatePassword(ctx, user.Email, hash); err != nil {
		return errors.New(errors.CodeInternal, "failed to save new password", err)
	}

	return nil
}

// all users realted services
func (s *authService) FindAllUsers(ctx context.Context, query string) ([]*domain.User, error) {

	return s.repo.FindAll(ctx, query)
}

func (s *authService) UploadAvatar(ctx context.Context, userId string,
	fileBytes []byte,
	filename string,
	contentType string) (string, error) {

	logger.Log.Info("Starting avatar upload for user" + userId)
	url, err := s.cloudzcli.UploadAvatar(ctx, fileBytes, filename)
	if err != nil {
		return "", errors.New(errors.CodeInternal, "failed to upload profile pic", err)
	}
	if err := s.repo.UpdateAvatar(ctx, userId, url); err != nil {
		return "", errors.New(errors.CodeInternal, "failed to upload profile pic", err)
	}

	return url, nil
}
