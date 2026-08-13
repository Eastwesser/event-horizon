package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Hand-written Validate() until protoc-gen-validate is regenerated into *.pb.validate.go.

func (m *RegisterRequest) Validate() error {
	if m == nil {
		return fmt.Errorf("request is nil")
	}
	if err := validateEmail(m.Email); err != nil {
		return err
	}
	if err := validatePassword(m.Password); err != nil {
		return err
	}
	role := strings.TrimSpace(m.Role)
	if role != "" && role != "user" && role != "author" {
		return fmt.Errorf("role must be empty, user, or author")
	}
	return nil
}

func (m *LoginRequest) Validate() error {
	if m == nil {
		return fmt.Errorf("request is nil")
	}
	if err := validateEmail(m.Email); err != nil {
		return err
	}
	return validatePassword(m.Password)
}

func (m *ValidateTokenRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Token) == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

func (m *RefreshTokenRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.RefreshToken) == "" {
		return fmt.Errorf("refresh_token is required")
	}
	return nil
}

func (m *WhoamiRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.AccessToken) == "" {
		return fmt.Errorf("access_token is required")
	}
	return nil
}

func (m *LogoutRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Token) == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}

func (m *GetUserRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (m *UpdateNicknameRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	n := utf8.RuneCountInString(m.Nickname)
	if n < 1 || n > 32 {
		return fmt.Errorf("nickname must be 1-32 characters")
	}
	return nil
}

func (m *UpdateRoleRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	role := strings.TrimSpace(m.Role)
	if role != "user" && role != "author" && role != "admin" {
		return fmt.Errorf("role must be user, author, or admin")
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if utf8.RuneCountInString(email) > 254 {
		return fmt.Errorf("email too long")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email")
	}
	return nil
}

func validatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < 8 || n > 128 {
		return fmt.Errorf("password must be 8-128 characters")
	}
	return nil
}
