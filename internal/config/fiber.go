package config

import (
	"github.com/gofiber/fiber/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

func NewFiber(logger *logrus.Logger) *fiber.App {
	app := fiber.New(
		fiber.Config{
			AppName:           "Sentra Backend",
			BodyLimit:         50 * 1024 * 1024,
			DisableKeepalive:  false,
			StrictRouting:     true,
			CaseSensitive:     true,
			EnablePrintRoutes: true,
			JSONEncoder:       jsoniter.Marshal,
			JSONDecoder:       jsoniter.Unmarshal,
		})

	return app
}
