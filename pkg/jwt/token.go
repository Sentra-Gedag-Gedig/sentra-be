package jwtPkg

import (
	"ProjectGolang/internal/entity"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func Sign(Data map[string]interface{}, ExpiredAt time.Duration) (string, int64, error) {
	expiredAt := time.Now().Add(ExpiredAt).Unix()

	JWTSecretKey := os.Getenv("JWT_ACCESS_TOKEN_SECRET")
	if JWTSecretKey == "" {
		return "", 0, fmt.Errorf("JWT_ACCESS_TOKEN_SECRET not set")
	}

	claims := jwt.MapClaims{}
	claims["exp"] = expiredAt
	claims["authorization"] = true

	for i, v := range Data {
		claims[i] = v
	}

	logrus.WithField("claims", claims).Debug("Creating token with claims")

	to := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := to.SignedString([]byte(JWTSecretKey))
	if err != nil {
		logrus.WithError(err).Error("Failed to sign token")
		return "", 0, err
	}

	return accessToken, expiredAt, nil
}

func VerifyTokenHeader(c *fiber.Ctx, secretEnvKey string) (*jwt.Token, error) {
	log := logrus.WithField("func", "VerifyTokenHeader")

	header := c.Get("Authorization")
	if header == "" {
		log.Error("Empty Authorization header")
		return nil, errors.New("empty Authorization header")
	}

	log.WithField("header", header[:10]+"...").Debug("Got Authorization header")

	parts := strings.Split(header, "Bearer ")
	if len(parts) != 2 {
		log.WithField("header_parts", len(parts)).Error("Invalid Authorization format")
		return nil, errors.New("invalid Authorization format")
	}

	accessToken := strings.TrimSpace(parts[1])
	if accessToken == "" {
		log.Error("Empty token after Bearer")
		return nil, errors.New("empty token")
	}

	log.Debug("Token format valid, attempting to parse")

	JWTSecretKey := os.Getenv(secretEnvKey)
	if JWTSecretKey == "" {
		log.Error("JWT_ACCESS_TOKEN_SECRET environment variable not set")
		return nil, errors.New("JWT secret not configured")
	}

	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.WithField("method", token.Header["alg"]).Error("Unexpected signing method")
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(JWTSecretKey), nil
	})

	if err != nil {
		log.WithError(err).Error("Failed to parse JWT token")
		return nil, err
	}

	log.Debug("Token successfully verified")
	return token, nil
}

func GetUserLoginData(c *fiber.Ctx) (entity.UserLoginData, error) {
	userData := c.Locals("user")

	user, ok := userData.(entity.UserLoginData)
	if !ok {
		return entity.UserLoginData{}, fiber.ErrUnauthorized
	}

	return user, nil
}
