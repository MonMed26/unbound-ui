package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNoToken            = errors.New("no token provided")
	ErrInvalidToken       = errors.New("invalid token")
)

type Config struct {
	Username      string `yaml:"username"`
	PasswordHash  string `yaml:"password_hash"`
	JWTSecret     string `yaml:"jwt_secret"`
	SessionTTL    string `yaml:"session_ttl"`
	ResetPassword string `yaml:"-" json:"-"` // Not persisted, only via env var
}

type Service struct {
	config *Config
}

func NewService(cfg *Config) *Service {
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = generateSecret()
	}
	// Handle password reset via env var
	if cfg.ResetPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.ResetPassword), bcrypt.DefaultCost)
		if err == nil {
			cfg.PasswordHash = string(hash)
		}
		cfg.ResetPassword = "" // Clear after use
	}
	return &Service{config: cfg}
}

func (s *Service) IsSetupRequired() bool {
	return s.config.Username == "" || s.config.PasswordHash == ""
}

func (s *Service) Setup(username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.config.Username = username
	s.config.PasswordHash = string(hash)
	return nil
}

func (s *Service) Login(username, password string) (string, error) {
	if username != s.config.Username {
		return "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(s.config.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return s.generateToken(username)
}

func (s *Service) ValidateToken(tokenStr string) error {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.config.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}
	return nil
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if err := s.ValidateToken(tokenStr); err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Service) GetConfig() *Config {
	return s.config
}

func (s *Service) generateToken(username string) (string, error) {
	ttl := 24 * time.Hour
	if s.config.SessionTTL != "" {
		if d, err := time.ParseDuration(s.config.SessionTTL); err == nil {
			ttl = d
		}
	}

	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(ttl).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func extractToken(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	return ""
}

func generateSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
