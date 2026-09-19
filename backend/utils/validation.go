package utils

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"
)

func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 2 || len(name) > 80 {
		return errors.New("name must be between 2 and 80 characters")
	}
	return nil
}

func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) > 254 {
		return errors.New("email is too long")
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || !strings.Contains(email, "@") {
		return errors.New("invalid email address")
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return errors.New("password must be between 8 and 72 characters")
	}
	var upper, lower, digit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			upper = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsDigit(r):
			digit = true
		}
	}
	if !upper || !lower || !digit {
		return errors.New("password must contain uppercase, lowercase, and a number")
	}
	return nil
}

func ValidatePoll(question string, options []string) error {
	question = strings.TrimSpace(question)
	if question == "" || len(question) > 240 {
		return errors.New("question is required and must be at most 240 characters")
	}
	if len(options) < 2 || len(options) > 10 {
		return errors.New("poll must have between 2 and 10 options")
	}
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		clean := strings.TrimSpace(option)
		if clean == "" || len(clean) > 120 {
			return errors.New("each option is required and must be at most 120 characters")
		}
		key := strings.ToLower(clean)
		if _, ok := seen[key]; ok {
			return errors.New("duplicate options are not allowed")
		}
		seen[key] = struct{}{}
	}
	return nil
}
