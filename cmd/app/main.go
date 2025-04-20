package main

import (
	"ProjectGolang/internal/config"
	"ProjectGolang/pkg/google"
	"ProjectGolang/pkg/log"
	"ProjectGolang/pkg/redis"
	"ProjectGolang/pkg/smtp"
	websocketPkg "ProjectGolang/pkg/websocket"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger := log.NewLogger()
	if err := godotenv.Load(); err != nil {
		logger.Fatalf("Error loading .env file: %v", err)
	}

	fiberApp := config.NewFiber(logger)
	validator := config.NewValidator()
	googleProvider := google.New()
	redisServer := redis.New()
	smtpMailer := smtp.New()
	websocket := websocketPkg.NewAIWebSocketClient()

	server, err := config.NewServer(
		config.WithFiber(fiberApp),
		config.WithLogger(logger),
		config.WithValidator(validator),
		config.WithDatabase(),
		config.WithGoogleProvider(googleProvider),
		config.WithRedisServer(redisServer),
		config.WithSMTPMailer(smtpMailer),
		config.WithWebSocket(websocket),
		config.WithMiddleware(),
		config.WithS3Client(),
		config.WithWhatsappClient(),
		config.WithGeminiClient(),
		config.WithBcryptUtils(),
		config.WithUtils(),
	)
	if err != nil {
		logger.Fatal(err)
	}

	server.RegisterHandler()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	logger.Info("Server started successfully")

	<-sigChan
	logger.Info("Shutting down server...")
}
