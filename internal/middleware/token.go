package middleware

import (
	"ProjectGolang/internal/entity"
	jwtPkg "ProjectGolang/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

const (
	AccessTokenSecret = "JWT_ACCESS_TOKEN_SECRET"
)

type tokenMiddleware struct {
}

func newTokenMiddleware() *tokenMiddleware {
	return &tokenMiddleware{}
}

func (m *middleware) NewTokenMiddleware(ctx *fiber.Ctx) error {
	clientIP := ctx.IP()
	authHeader := ctx.Get("Authorization")

	m.log.WithFields(logrus.Fields{
		"path":      ctx.Path(),
		"method":    ctx.Method(),
		"client_ip": clientIP,
		"headers":   ctx.GetReqHeaders(),
	}).Warn("Incoming request")

	if authHeader == "" {
		m.log.WithFields(logrus.Fields{
			"error": "Authorization header is missing",
		}).Warn("Authorization header check")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized, access token invalid or expired",
		})
	}

	headerParts := strings.Split(authHeader, " ")
	m.log.WithFields(logrus.Fields{
		"auth_type": headerParts[0],
		"has_token": len(headerParts) > 1,
	}).Debug("Authorization header check")

	if !strings.HasPrefix(authHeader, "Bearer ") {
		m.log.WithFields(logrus.Fields{
			"error": "Authorization header format is invalid",
		}).Warn("Authorization header check")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized, access token invalid or expired",
		})
	}

	userToken, err := jwtPkg.VerifyTokenHeader(ctx, AccessTokenSecret)
	if err != nil {
		m.log.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Warn("Token verification failed")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized, access token invalid or expired",
		})
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		m.log.WithFields(logrus.Fields{
			"error": "Invalid token claims",
		}).Warn("Token claims check")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized, access token invalid or expired",
		})
	}

	m.log.WithFields(logrus.Fields{
		"claim_keys":   reflect.ValueOf(claims).MapKeys(),
		"exp":          claims["exp"],
		"id_exists":    claims["id"] != nil,
		"email_exists": claims["email"] != nil,
	}).Debug("Token claims")

	if claims["id"] == nil || claims["email"] == nil || claims["username"] == nil {
		m.log.WithFields(logrus.Fields{
			"error": "Token claims are missing required fields",
		}).Warn("Token claims check")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized, access token invalid or expired",
		})
	}

	user := entity.UserLoginData{
		ID:       claims["id"].(string),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}
	ctx.Locals("user", user)

	m.log.Info("Authentication successful")
	return ctx.Next()
}
