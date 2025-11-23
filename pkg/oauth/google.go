package oauth

import (
	"context"
	"errors"

	"google.golang.org/api/idtoken"
)

func VerifyGoogleToken(ctx context.Context, token string) (*OAuthData, error) {
	payload, err := idtoken.Validate(ctx, token, "")
	if err != nil {
		return nil, errors.New("invalid google token")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)

	return &OAuthData{
		Email:    email,
		Name:     name,
		Provider: "google",
	}, nil
}
