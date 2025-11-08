package jwt

import (
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func createJWT(secret []byte, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

func ValidateToken(secret []byte, tokenString string, purpose jwtpurpose.JWTPurpose) (*externalservice.CustomClaims, error) {
	claims := &externalservice.CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errorcode.ErrUnexpectedSigningToken
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	// check if exp
	if !token.Valid {
		return nil, errorcode.ErrInvalidToken
	}

	if claims.Purpose != purpose {
		return nil, errorcode.ErrInvalidJWTPurpose
	}

	return claims, nil
}

// GenerateAcAndRtTokens creates access token and refresh token
func GenerateAcAndRtTokens(cfg *config.JWT, userID uuid.UUID) (string, string, error) {
	accessToken, err := createJWT([]byte(cfg.AccessTokenKey), externalservice.CustomClaims{
		Purpose: jwtpurpose.Access,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	if err != nil {
		return "", "", err
	}

	refreshToken, err := createJWT([]byte(cfg.RefreshTokenKey), externalservice.CustomClaims{
		Purpose: jwtpurpose.Refresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Generate email token
func GenerateEmailToken(secret []byte, expiresIn time.Duration, email string, purpose jwtpurpose.JWTPurpose) (string, error) {
	emailVerifyToken, err := createJWT(
		secret,
		externalservice.CustomClaims{
			Purpose: purpose,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   email,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		},
	)
	if err != nil {
		return "", err
	}
	return emailVerifyToken, nil
}
