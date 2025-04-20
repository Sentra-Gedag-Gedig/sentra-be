package google

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ItfGoogle interface {
	GetUserExchangeToken(c *fiber.Ctx, code string) ([]byte, error)
	GetConfig() *oauth2.Config
}

type googleProvider struct {
	config *oauth2.Config
}

func New() ItfGoogle {
	oauthConfgl := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/api/v1/auth/callback-gl",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}

	return &googleProvider{config: oauthConfgl}
}

func (g *googleProvider) GetUserExchangeToken(c *fiber.Ctx, code string) ([]byte, error) {
	token, err := g.config.Exchange(c.Context(), code)
	if err != nil {
		fmt.Printf("Error exchanging token: %v", err)
		return nil, err
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		fmt.Printf("Error getting user info: %v", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing body: %v", err)
		}
	}(resp.Body)

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (g *googleProvider) GetConfig() *oauth2.Config {
	return g.config
}
