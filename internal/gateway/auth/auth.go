package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUserInactive = errors.New("user is inactive")

type User struct {
	ID           string
	Login        string
	PasswordHash string
	Role         string
	Status       string
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	pool      *pgxpool.Pool
	secret    []byte
	ttl       time.Duration
	issuer    string
}

func NewService(pool *pgxpool.Pool, secret string, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Service{pool: pool, secret: []byte(secret), ttl: ttl, issuer: "grpc-bank-gateway"}
}

func (s *Service) Authenticate(ctx context.Context, login, password string) (string, *User, error) {
	if s == nil || s.pool == nil {
		return "", nil, fmt.Errorf("auth service not configured")
	}
	const q = `SELECT id::text, login, password_hash, role, status FROM users WHERE login = $1`
	var u User
	err := s.pool.QueryRow(ctx, q, login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Role, &u.Status)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}
	if u.Status != "ACTIVE" {
		return "", nil, ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}
	now := time.Now()
	claims := Claims{
		UserID: u.ID,
		Role:   u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", nil, err
	}
	return signed, &u, nil
}

func (s *Service) ParseToken(tokenStr string) (*Claims, error) {
	if s == nil {
		return nil, fmt.Errorf("auth service not configured")
	}
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

type ctxKey string

const userContextKey ctxKey = "auth_user"

func WithUser(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, userContextKey, claims)
}

func UserFromContext(ctx context.Context) *Claims {
	v := ctx.Value(userContextKey)
	if v == nil {
		return nil
	}
	c, _ := v.(*Claims)
	return c
}

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
