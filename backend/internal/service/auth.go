package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/spider/spider/dao/store"
	"github.com/spider/spider/internal/spidererrors"
	pkgauth "github.com/spider/spider/pkg/auth"
	"github.com/spider/spider/pkg/config"
)

type AuthService struct {
	Users    *store.UserRepo
	Settings *config.Settings
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*store.User, error) {
	user, err := s.Users.GetByEmail(ctx, email)
	if err != nil || !user.IsActive || !pkgauth.VerifyPassword(password, user.HashedPassword) {
		return nil, spidererrors.Authentication("")
	}
	return user, nil
}

func (s *AuthService) IssueToken(user *store.User) (string, error) {
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"email": user.Email,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Duration(s.Settings.JWTExpireMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.Settings.JWTSecret))
}

func (s *AuthService) DecodeToken(tokenStr string) (map[string]string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.Settings.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, spidererrors.Authentication("Invalid or expired token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, spidererrors.Authentication("Invalid token claims")
	}
	return map[string]string{
		"sub":   fmt.Sprint(claims["sub"]),
		"email": fmt.Sprint(claims["email"]),
		"role":  fmt.Sprint(claims["role"]),
	}, nil
}

func (s *AuthService) UserFromToken(ctx context.Context, tokenStr string) (*store.User, error) {
	claims, err := s.DecodeToken(tokenStr)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(claims["sub"])
	if err != nil {
		return nil, spidererrors.Authentication("Invalid token subject")
	}
	user, err := s.Users.GetByID(ctx, userID)
	if err != nil || !user.IsActive {
		return nil, spidererrors.Authentication("User not found")
	}
	return user, nil
}

func (s *AuthService) RequireRoles(user *store.User, allowed ...string) error {
	for _, role := range allowed {
		if user.Role == role {
			return nil
		}
	}
	return spidererrors.Authorization(fmt.Sprintf("Role %s is not permitted", user.Role))
}
