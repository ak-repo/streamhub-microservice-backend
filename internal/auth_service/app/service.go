package app

import (
	"context"
	"fmt"
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

const (
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

// NewAuthService creates a new instance of the Auth Service.
func NewAuthService(
	repo port.UserRepository,
	jwtMgr *jwt.JWTManager,
	cfg *config.Config,
	otpStore *otpredis.OTPStore,
	cloudzcli *authcloudinary.CloudinaryUploader,
) port.AuthService {
	return &authService{
		repo:      repo,
		jwt:       jwtMgr,
		cfg:       cfg,
		otpStore:  otpStore,
		cloudzcli: cloudzcli,
	}
}

// =============================================================================
// AUTHENTICATION (Register, Login, Magic Link)
// =============================================================================

func (s *authService) Register(ctx context.Context, email, username, password string) error {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return errors.New(errors.CodeConflict, ErrUserExists, nil)
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return errors.New(errors.CodeInternal, "failed to hash password", err)
	}

	user := &domain.User{
		ID:            uuid.New().String(),
		Email:         email,
		Username:      username,
		PasswordHash:  hashed,
		Role:          "user",
		AvatarURL:     "https://res.cloudinary.com/dersnukrf/image/upload/v1765461300/avatars/avatars/b304408f6711cd1fa4fa119eacde9a6b.jpg",
		CreatedAt:     time.Now(),
		EmailVerified: false,
		IsBanned:      false,
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
		return nil, errors.New(errors.CodeNotFound, ErrInvalidCreds, nil)
	}

	if !utils.ComparePassword(user.PasswordHash, password) {
		return nil, errors.New(errors.CodeConflict, ErrInvalidCreds, nil)
	}

	if user.IsBanned {
		log.Printf("Login blocked: user %s is banned", user.Username)
		return nil, errors.New(errors.CodeForbidden, "account is suspended", nil)
	}

	return user, nil
}

func (s *authService) SendMagicLink(email string) (string, string, error) {
	token, exp, err := s.jwt.GenerateAccessToken("0", email, "magic-link")
	if err != nil {
		return "", "", errors.New(errors.CodeInternal, "token generation failed", err)
	}

	// Construct URL
	magicLink := fmt.Sprintf("http://%s:%s/verify-link?email=%s&token=%s",
		s.cfg.Services.Front.Host,
		s.cfg.Services.Front.Port,
		email,
		token,
	)

	// Prepare SendGrid Email
	from := mail.NewEmail("StreamHub", "ak001mob@gmail.com")
	to := mail.NewEmail("", email)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "Your Magic Login Link"
	message.SetTemplateID(s.cfg.SendGrid.MagicTemplateID)

	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("magic_link", magicLink)
	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.cfg.SendGrid.APIKey)
	response, err := client.Send(message)

	if err != nil || response.StatusCode >= 300 {
		log.Println("errrr: ", response.StatusCode)
		return "", "", errors.New(errors.CodeInternal, "failed to send verify link email", err)
	}
	log.Println("email: ", email, " status: ", response.StatusCode)

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

	// Update verification status if not already verified
	if !user.EmailVerified {
		user.EmailVerified = true
		if err := s.repo.Update(ctx, user); err != nil {
			return errors.New(errors.CodeInternal, "database update error", err)
		}
	}

	return nil
}

// =============================================================================
// USER LOOKUP & PROFILE
// =============================================================================

func (s *authService) FindUser(ctx context.Context, identifier, method string) (*domain.User, error) {
	var user *domain.User
	var err error

	switch method {
	case "id":
		user, err = s.repo.FindByID(ctx, identifier)
	case "email":
		user, err = s.repo.FindByEmail(ctx, identifier)
	case "username":
		user, err = s.repo.FindByUsername(ctx, identifier)
	default:
		return nil, errors.New(errors.CodeBadRequest, "invalid lookup method", nil)
	}

	if err != nil {
		return nil, errors.New(errors.CodeInternal, "database error", err)
	}
	if user == nil {
		return nil, errors.New(errors.CodeNotFound, fmt.Sprintf("user not found by %s: %s", method, identifier), nil)
	}

	return user, nil
}

func (s *authService) SearchUsers(ctx context.Context, query string) ([]*domain.User, error) {
	// Renamed from FindAllUsers to match the Interface and intent
	users, err := s.repo.SearchUsers(ctx, query)
	if err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to search users", err)
	}
	return users, nil
}

func (s *authService) UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}
	if username != "" && username != user.Username {
		user.Username = username
	}

	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.New(errors.CodeInternal, "failed to update user details", err)
	}

	return user, nil
}

func (s *authService) UploadAvatar(ctx context.Context, userID string, fileBytes []byte, filename, contentType string) (string, error) {
	logger.Log.Info("Starting avatar upload for user: " + userID)

	url, err := s.cloudzcli.UploadAvatar(ctx, fileBytes, filename)
	if err != nil {
		return "", errors.New(errors.CodeInternal, "failed to upload image to storage", err)
	}

	if err := s.repo.UpdateAvatar(ctx, userID, url); err != nil {
		return "", errors.New(errors.CodeInternal, "failed to update avatar url in db", err)
	}

	return url, nil
}

// =============================================================================
// PASSWORD MANAGEMENT (Reset & Change)
// =============================================================================

func (s *authService) PasswordReset(ctx context.Context, email string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		// Return Generic Success to prevent Email Enumeration attacks
		return nil
	}

	otp, err := utils.GenerateOTP(6)
	if err != nil {
		return errors.New(errors.CodeInternal, "otp generation failed", err)
	}

	if err := s.otpStore.SaveOTP(ctx, email, otp); err != nil {
		return errors.New(errors.CodeInternal, "failed to store otp", err)
	}

	// Send Email
	from := mail.NewEmail("StreamHub", "ak001mob@gmail.com")
	to := mail.NewEmail("", email)

	message := mail.NewV3Mail()
	message.SetFrom(from)
	message.Subject = "Your Password Reset Code"
	message.SetTemplateID(s.cfg.SendGrid.OTPTemplateID)

	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("otp_code", otp)
	p.SetDynamicTemplateData("username", user.Username)
	p.SetDynamicTemplateData("support_email", "support@streamhub.com")
	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(s.cfg.SendGrid.APIKey)
	resp, err := client.Send(message)
	

	if err != nil || resp.StatusCode >= 300 {
		return errors.New(errors.CodeInternal, "failed to send otp email", err)
	}

	return nil
}

func (s *authService) VerifyPasswordReset(ctx context.Context, otp, password, email string) error {
	cachedOTP, err := s.otpStore.VerifyOTP(ctx, email)
	if err != nil {
		return errors.New(errors.CodeInternal, "otp verification error", err)
	}
	if cachedOTP != otp {
		return errors.New(errors.CodeForbidden, "invalid otp", nil)
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return errors.New(errors.CodeInternal, "failed to hash password", err)
	}

	if err := s.repo.UpdatePassword(ctx, email, hash); err != nil {
		return errors.New(errors.CodeInternal, "failed to save new password", err)
	}

	_ = s.otpStore.DeleteOTP(ctx, email)
	return nil
}

func (s *authService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}

	if !utils.ComparePassword(user.PasswordHash, oldPassword) {
		return errors.New(errors.CodeUnauthorized, "current password incorrect", nil)
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New(errors.CodeInternal, "failed to hash new password", err)
	}

	if err := s.repo.UpdatePassword(ctx, user.Email, hash); err != nil {
		return errors.New(errors.CodeInternal, "database update failed", err)
	}

	return nil
}

// =============================================================================
// ADMIN / GOVERNANCE
// =============================================================================

func (s *authService) ListUsers(ctx context.Context, filter string, limit, offset int32) ([]*domain.User, int32, error) {

	users, total, err := s.repo.ListUsers(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, errors.New(errors.CodeInternal, "failed to list users", err)
	}
	return users, total, nil
}

func (s *authService) BanUser(ctx context.Context, userID, reason string) error {
	if err := s.repo.BanUser(ctx, userID, reason); err != nil {
		return errors.New(errors.CodeInternal, "failed to ban user", err)
	}
	return nil
}

func (s *authService) UnbanUser(ctx context.Context, userID, reason string) error {
	if err := s.repo.UnbanUser(ctx, userID, reason); err != nil {
		return errors.New(errors.CodeInternal, "failed to unban user", err)
	}
	return nil
}

func (s *authService) UpdateRole(ctx context.Context, userID, role string) error {
	if err := s.repo.UpdateRole(ctx, userID, role); err != nil {
		return errors.New(errors.CodeInternal, "failed to update user role", err)
	}
	return nil
}

func (s *authService) BlockUserUpload(ctx context.Context, adminID, userID string, block bool) error {
	// In a real app, verify adminID has permissions here
	if err := s.repo.SetUserUploadBlocked(ctx, userID, block); err != nil {
		return errors.New(errors.CodeInternal, "failed to update user upload status", err)
	}
	return nil
}

func (s *authService) DeleteUser(ctx context.Context, id string) error {

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		return errors.New(errors.CodeNotFound, "failed to delete user", err)
	}
	return nil
}
