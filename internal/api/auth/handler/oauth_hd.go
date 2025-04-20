package authHandler

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	"ProjectGolang/pkg/log"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

func (h *AuthHandler) HandleGoogleLogin(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	url, err := h.authService.Auth().LoginGoogle()
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "google_login")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return ctx.Redirect(url.String(), fiber.StatusTemporaryRedirect)
	}
}

func (h *AuthHandler) CallBackFromGoogle(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	state := ctx.FormValue("state")
	if state != os.Getenv("GOOGLE_STATE") {
		h.log.WithFields(log.Fields{
			"request_id": requestID,
			"state":      state,
			"path":       ctx.Path(),
		}).Warn("Invalid state parameter")
		return ctx.Redirect("/", fiber.StatusTemporaryRedirect)
	}

	code := ctx.FormValue("code")

	if code == "" {
		reason := ctx.FormValue("error_reason")
		if reason == "user_denied" {
			h.log.WithFields(log.Fields{
				"request_id": requestID,
				"reason":     reason,
				"path":       ctx.Path(),
			}).Info("User denied access")
			return errHandler.HandleUnauthorized(ctx, requestID, "Access denied by user")
		}

		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("No authorization code provided"), ctx.Path())
	}

	gConfig := h.googleProvider.GetConfig()
	token, err := gConfig.Exchange(c, code)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "exchange_token")
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_user_info")
	}
	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "read_response")
	}

	var userInfo auth.UserGoogle
	if err := json.Unmarshal(response, &userInfo); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "unmarshal_user_info")
	}

	_, err = h.authService.User().GetByEmail(c, userInfo.Email)
	if errors.Is(err, auth.ErrUserWithEmailNotFound) {
		if err := h.authService.User().RegisterUser(c, auth.CreateUserRequest{
			Name:        userInfo.Email,
			PhoneNumber: "",
			Password:    "",
		}); err != nil {
			return errHandler.Handle(ctx, requestID, err, ctx.Path(), "register_google_user")
		}
	} else if err != nil && !errors.Is(err, auth.ErrUserWithEmailNotFound) {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "check_user_email")
	}

	jwtToken, err := h.authService.Auth().Login(c, auth.LoginUserRequest{
		Email: userInfo.Email,
	})
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "login_user")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"token":   jwtToken.AccessToken,
			"expires": jwtToken.ExpiresInMinutes})
	}
}
