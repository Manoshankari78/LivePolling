package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"live-polling-app/backend/models"
	"live-polling-app/backend/repositories"
	"live-polling-app/backend/utils"
)

type AuthService struct {
	users repositories.UserRepository
	jwt   *utils.JWTManager
}

func NewAuthService(users repositories.UserRepository, jwt *utils.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*models.User, string, error) {
	name, email = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(email))
	if err := utils.ValidateName(name); err != nil {
		return nil, "", err
	}
	if err := utils.ValidateEmail(email); err != nil {
		return nil, "", err
	}
	if err := utils.ValidatePassword(password); err != nil {
		return nil, "", err
	}
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, "", ErrConflict
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}
	now := time.Now()
	user := &models.User{Name: name, Email: email, PasswordHash: string(hash), CreatedAt: now, UpdatedAt: now}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, "", err
	}
	token, err := s.jwt.Generate(user.ID.Hex())
	return user, token, err
}
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, "", ErrInvalidCredentials
	}
	token, err := s.jwt.Generate(user.ID.Hex())
	return user, token, err
}
func (s *AuthService) Me(ctx context.Context, id string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrUnauthorized
	}
	return s.users.FindByID(ctx, objectID)
}
