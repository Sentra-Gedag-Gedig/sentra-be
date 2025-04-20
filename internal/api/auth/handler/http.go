package authHandler

import (
	authService "ProjectGolang/internal/api/auth/service"
	"ProjectGolang/internal/middleware"
	"ProjectGolang/pkg/google"
	"ProjectGolang/pkg/redis"
	"ProjectGolang/pkg/s3"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	log            *logrus.Logger
	authService    authService.AuthService
	validator      *validator.Validate
	middleware     middleware.Middleware
	googleProvider google.ItfGoogle
	redisServer    redis.IRedis
	s3Client       s3.ItfS3
}

func New(
	log *logrus.Logger,
	as authService.AuthService,
	validate *validator.Validate,
	middleware middleware.Middleware,
	googleProvider google.ItfGoogle,
	redisServer redis.IRedis,
	s3Client s3.ItfS3) *AuthHandler {
	return &AuthHandler{
		log:            log,
		authService:    as,
		validator:      validate,
		middleware:     middleware,
		googleProvider: googleProvider,
		redisServer:    redisServer,
		s3Client:       s3Client,
	}
}

func (h *AuthHandler) Start(srv fiber.Router) {
	auth := srv.Group("/auth")
	auth.Post("/login", h.HandleLogin)
	auth.Post("/login-touch-id", h.LoginTouchID)
	auth.Get("/login-gl", h.HandleGoogleLogin)
	auth.Get("/callback-gl", h.CallBackFromGoogle)
	auth.Patch("/enable-touch-id", h.middleware.NewTokenMiddleware, h.EnableTouchID)

	users := srv.Group("/users")
	users.Post("/", h.HandleRegister)
	users.Post("/profile-photo", h.middleware.NewTokenMiddleware, h.HandleUpdateProfilePhoto)
	users.Post("/face-photo", h.middleware.NewTokenMiddleware, h.HandleUpdateFacePhoto)
	users.Get("/profile-photo", h.middleware.NewTokenMiddleware, h.HandleGetProfilePhoto)
	users.Get("/:id", h.middleware.NewTokenMiddleware, h.HandleGetUserById)
	users.Patch("/", h.middleware.NewTokenMiddleware, h.HandleUpdateUser)
	users.Delete("/:id", h.HandleDeleteUser)

	password := srv.Group("/password")
	password.Patch("/reset-password", h.HandleResetPassword)

	verification := srv.Group("/verification")
	verification.Post("/phone-number-verification", h.HandlePhoneNumberVerification)
	auth.Post("/phone/verify-otp", h.middleware.NewTokenMiddleware, h.HandleVerifyPhoneOTP)

	verification.Patch("/verify-user", h.HandleVerifyOTPandVerifyingUser)
	verification.Post("/verify-otp-update-pin", h.HandleVerifyOTPandPIN)

	verification.Post("/email/send-otp", h.HandleSendEmailOTP)
	verification.Post("/email/verify-otp", h.middleware.NewTokenMiddleware, h.HandleVerifyEmailOTP)
}
