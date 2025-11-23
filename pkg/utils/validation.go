package utils

import "errors"

func ValidateRegister(username, email, password string) error {
	if username == "" || email == "" || password == "" {
		return errors.New("all fields are required")
	}
	return nil
}

func ValidateLogin(email, password string) error {
	if email == "" || password == "" {
		return errors.New("email and password are required")
	}
	return nil
}
