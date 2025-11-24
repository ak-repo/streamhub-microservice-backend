package app

import (
	"context"
	"time"

	"github.com/ak-repo/stream-hub/config"
	"github.com/ak-repo/stream-hub/internal/auth_service/domain"
	"github.com/ak-repo/stream-hub/internal/auth_service/port"
	"github.com/ak-repo/stream-hub/pkg/errors"
	"github.com/ak-repo/stream-hub/pkg/helper"
	"github.com/ak-repo/stream-hub/pkg/jwt"
	"github.com/ak-repo/stream-hub/pkg/utils"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	ErrUserExists   = "user already exists"
	ErrInvalidCreds = "invalid credentials"
	ErrUserNotFound = "user not found"
)

type authService struct {
	repo port.UserRepository
	jwt  *jwt.JWTManager
	cfg  *config.Config
}

func NewAuthService(repo port.UserRepository, jwtMgr *jwt.JWTManager, cfg *config.Config) port.AuthService {
	return &authService{repo: repo, jwt: jwtMgr, cfg: cfg}
}

// -------------------- REGISTER --------------------
func (s *authService) Register(ctx context.Context, email, username, password string) (*domain.User, error) {
	existing, _ := s.repo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New(errors.CodeConflict, ErrUserExists, nil)
	}

	hashed, _ := utils.HashPassword(password)

	user := &domain.User{
		Email:        email,
		Username:     username,
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
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

	return user, nil
}

// --------------------------Magic Link --------------------------------
func (s *authService) SendMagicLink(email string) (string, string, error) {
	token, exp, err := s.jwt.GenerateAccessToken("0", email)
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
	message.SetTemplateID("d-9f3316146d5d46e4ba3efdc8c6ba98c6")

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

func (s *authService) VerifyMagicLink(ctx context.Context, token, email string) (*domain.User, error) {

	claims, err := s.jwt.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if claims.Email != email {
		return nil, errors.New(errors.CodeUnauthorized, "token email mismatch", nil)
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New(errors.CodeNotFound, ErrUserNotFound, err)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.New(errors.CodeInternal, "database update error", err)
	}

	user.PasswordHash = ""
	return user, nil
}


