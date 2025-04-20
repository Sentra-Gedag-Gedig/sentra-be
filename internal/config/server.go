package config

import (
	"ProjectGolang/database/postgres"
	authHandler "ProjectGolang/internal/api/auth/handler"
	authRepository "ProjectGolang/internal/api/auth/repository"
	authService "ProjectGolang/internal/api/auth/service"
	budgetHandler "ProjectGolang/internal/api/budget_manager/handler"
	budgetRepository "ProjectGolang/internal/api/budget_manager/repository"
	budgetService "ProjectGolang/internal/api/budget_manager/service"
	detectionHandler "ProjectGolang/internal/api/detection/handler"
	detectionService "ProjectGolang/internal/api/detection/service"
	sentrapayHandler "ProjectGolang/internal/api/sentra_pay/handler"
	sentrapayRepository "ProjectGolang/internal/api/sentra_pay/repository"
	sentrapayService "ProjectGolang/internal/api/sentra_pay/service"
	"ProjectGolang/internal/middleware"
	"ProjectGolang/pkg/bcrypt"
	"ProjectGolang/pkg/doku"
	"ProjectGolang/pkg/gemini"
	"ProjectGolang/pkg/google"
	"ProjectGolang/pkg/redis"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/smtp"
	"ProjectGolang/pkg/utils"
	websocketPkg "ProjectGolang/pkg/websocket"
	"ProjectGolang/pkg/whatsapp"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"os"
)

type ServerOption func(*Server) error

type Server struct {
	engine         *fiber.App
	db             *sqlx.DB
	log            *logrus.Logger
	middleware     middleware.Middleware
	validator      *validator.Validate
	utils          utils.IUtils
	bcryptUtils    bcrypt.IBcrypt
	handlers       []handler
	googleProvider google.ItfGoogle
	redisServer    redis.IRedis
	smtpMailer     smtp.ItfSmtp
	faceWebsocket  websocketPkg.IWebsocket
	whatsappClient whatsapp.IWhatsappSender
	geminiClient   gemini.IGemini
	s3Client       s3.ItfS3
}

type handler interface {
	Start(srv fiber.Router)
}

func NewServer(options ...ServerOption) (*Server, error) {
	server := &Server{}

	for _, option := range options {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if server.engine == nil {
		return nil, fmt.Errorf("fiber app is required")
	}
	if server.log == nil {
		return nil, fmt.Errorf("logger is required")
	}

	return server, nil
}

func WithFiber(fiberApp *fiber.App) ServerOption {
	return func(s *Server) error {
		s.engine = fiberApp
		return nil
	}
}

func WithLogger(logger *logrus.Logger) ServerOption {
	return func(s *Server) error {
		s.log = logger
		return nil
	}
}

func WithValidator(validator *validator.Validate) ServerOption {
	return func(s *Server) error {
		s.validator = validator
		return nil
	}
}

func WithDatabase() ServerOption {
	return func(s *Server) error {
		db, err := postgres.New()
		if err != nil {
			if s.log != nil {
				s.log.Errorf("Failed to connect to database: %v", err)
			}
			return fmt.Errorf("failed to create database connection: %w", err)
		}
		s.db = db
		return nil
	}
}

func WithGoogleProvider(provider google.ItfGoogle) ServerOption {
	return func(s *Server) error {
		s.googleProvider = provider
		return nil
	}
}

func WithRedisServer(redisServer redis.IRedis) ServerOption {
	return func(s *Server) error {
		s.redisServer = redisServer
		return nil
	}
}

func WithSMTPMailer(smtpMailer smtp.ItfSmtp) ServerOption {
	return func(s *Server) error {
		s.smtpMailer = smtpMailer
		return nil
	}
}

func WithWebSocket(webSocket websocketPkg.IWebsocket) ServerOption {
	return func(s *Server) error {
		s.faceWebsocket = webSocket
		return nil
	}
}

func WithMiddleware() ServerOption {
	return func(s *Server) error {
		if s.log == nil {
			return fmt.Errorf("logger must be initialized before middleware")
		}
		s.middleware = middleware.New(s.log)
		return nil
	}
}

func WithS3Client() ServerOption {
	return func(s *Server) error {
		client, err := s3.New()
		if err != nil {
			if s.log != nil {
				s.log.Errorf("Failed to initialize S3 client: %v", err)
			}
			return fmt.Errorf("failed to create S3 client: %w", err)
		}
		s.s3Client = client
		return nil
	}
}

func WithWhatsappClient() ServerOption {
	return func(s *Server) error {
		client, err := whatsapp.New()
		if err != nil {
			if s.log != nil {
				s.log.Errorf("Failed to initialize WhatsApp client: %v", err)
			}
			return fmt.Errorf("failed to create WhatsApp client: %w", err)
		}
		s.whatsappClient = client
		return nil
	}
}

func WithGeminiClient() ServerOption {
	return func(s *Server) error {
		client, err := gemini.NewGeminiClient()
		if err != nil {
			if s.log != nil {
				s.log.Errorf("Failed to create Gemini client: %v", err)
			}
			return fmt.Errorf("failed to create Gemini client: %w", err)
		}
		s.geminiClient = client
		return nil
	}
}

func WithUtils() ServerOption {
	return func(s *Server) error {
		s.utils = utils.New()
		return nil
	}
}

func WithBcryptUtils() ServerOption {
	return func(s *Server) error {
		s.bcryptUtils = bcrypt.New()
		return nil
	}
}

func (s *Server) RegisterHandler() {
	// Auth Domain
	authRepo := authRepository.New(s.db, s.log)
	authServices := authService.New(s.log, authRepo, s.googleProvider, s.smtpMailer, s.redisServer, s.whatsappClient, s.s3Client, s.smtpMailer, s.bcryptUtils, s.utils)
	authHandlers := authHandler.New(s.log, authServices, s.validator, s.middleware, s.googleProvider, s.redisServer, s.s3Client)

	// Detection
	detectionServices := detectionService.NewDetectionService(s.faceWebsocket, s.geminiClient)
	detectionHandlers := detectionHandler.New(s.log, s.validator, s.middleware, detectionServices, s.utils)

	// Budget Manager
	budgetRepo := budgetRepository.New(s.db, s.log)
	budgetServices := budgetService.NewBudgetService(s.log, budgetRepo, s.s3Client, s.utils)
	budgetHandlers := budgetHandler.New(s.log, s.validator, s.middleware, budgetServices)

	// Payment Domain
	dokuClient := doku.NewDokuService(s.log)
	dokuClient.Init()
	dokuRepo := sentrapayRepository.New(s.db, s.log)
	dokuServices := sentrapayService.NewSentraPayService(s.log, dokuRepo, dokuClient, authRepo, s.utils)
	dokuHandlers := sentrapayHandler.New(s.log, s.validator, s.middleware, dokuServices)

	s.setupHealthCheck()
	s.handlers = append(s.handlers, authHandlers, detectionHandlers, budgetHandlers, dokuHandlers)
}

func (s *Server) Run() error {
	router := s.engine.Group("/api/v1")
	s.engine.Use(s.middleware.NewRequestIDMiddleware())

	for _, h := range s.handlers {
		h.Start(router)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	if err := s.engine.Listen(fmt.Sprintf(":%s", port)); err != nil {
		if s.whatsappClient != nil {
			s.whatsappClient.Disconnect()
		}
		return err
	}

	return nil
}

func (s *Server) setupHealthCheck() {
	s.engine.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"message": "Server is Healthy!",
		})
	})
}
