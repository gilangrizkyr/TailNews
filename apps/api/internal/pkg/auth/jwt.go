package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"news-portal/apps/api/internal/domain/user"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

type Claims struct {
	UserID   uint64 `json:"user_id"`
	RoleID   uint64 `json:"role_id"`
	RoleName string `json:"role_name"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

type PreviewClaims struct {
	ArticleID uint64 `json:"article_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived access JWT for authenticated API requests.
func GenerateAccessToken(secret string, u *user.User, roleName string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:   u.ID,
		RoleID:   u.RoleID,
		RoleName: roleName,
		Email:    u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", u.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "news-portal-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken creates a long-lived refresh JWT stored in an HttpOnly cookie.
func GenerateRefreshToken(secret string, userID uint64, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		Issuer:    "news-portal-refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateAccessToken parses and validates a signed access JWT string.
func ValidateAccessToken(secret string, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ValidateRefreshToken parses and returns the user ID from a signed refresh JWT string.
func ValidateRefreshToken(secret string, tokenStr string) (uint64, error) {
	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		return 0, ErrTokenInvalid
	}

	if !token.Valid {
		return 0, ErrTokenInvalid
	}

	var userID uint64
	_, err = fmt.Sscanf(claims.Subject, "%d", &userID)
	if err != nil || userID == 0 {
		return 0, ErrTokenInvalid
	}

	return userID, nil
}

// GeneratePreviewToken creates a signed token specifically for draft article preview.
func GeneratePreviewToken(secret string, articleID uint64, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := PreviewClaims{
		ArticleID: articleID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", articleID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "news-portal-preview",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidatePreviewToken verifies and returns the article ID from a draft preview token.
func ValidatePreviewToken(secret string, tokenStr string) (uint64, error) {
	var claims PreviewClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return 0, ErrTokenInvalid
	}

	return claims.ArticleID, nil
}
