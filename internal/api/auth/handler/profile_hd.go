package authHandler

import (
	"ProjectGolang/internal/api/auth"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"time"
)

func (h *AuthHandler) HandleRegister(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing user creation request")

	var req auth.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	if err := h.authService.User().RegisterUser(c, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "register_user")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusCreated, nil)
	}
}

func (h *AuthHandler) HandleUpdateUser(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	var req auth.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "parse_request_body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	if err := h.authService.User().UpdateUser(c, userData, req); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_user")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleDeleteUser(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	id := ctx.Params("id")

	if err := h.authService.User().DeleteUser(c, id); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "delete_user")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}

func (h *AuthHandler) HandleUpdateProfilePhoto(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing update profile photo request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_form_file")
	}

	result, err := h.authService.User().UpdateProfilePhoto(c, userData.ID, file)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_profile_photo")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}

func (h *AuthHandler) HandleGetUserById(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get user by id request")

	var userId string
	paramId := ctx.Params("id")

	if paramId != "" && paramId != "me" {
		userId = paramId
	} else {
		userData, err := jwtPkg.GetUserLoginData(ctx)
		if err != nil {
			return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
		}
		userId = userData.ID
	}

	repo, err := h.authService.GetRepository().NewClient(false)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_repository")
	}

	user, err := repo.Users.GetByID(c, userId)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_user_by_id")
	}

	profilePhotoURL := user.ProfilePhotoURL
	if profilePhotoURL != "" {
		presignedURL, err := h.s3Client.PresignUrl(profilePhotoURL)
		if err != nil {
			h.log.WithFields(log.Fields{
				"request_id": requestID,
				"error":      err.Error(),
				"path":       ctx.Path(),
			}).Warn("Failed to presign URL for profile photo")
		} else {
			profilePhotoURL = presignedURL
		}
	}

	response := auth.UserResponse{
		ID:                        user.ID,
		Email:                     user.Email,
		Name:                      user.Name,
		NationalIdentityNumber:    user.NationalIdentityNumber,
		BirthPlace:                user.BirthPlace,
		BirthDate:                 user.BirthDate,
		Gender:                    user.Gender,
		Address:                   user.Address,
		NeighborhoodCommunityUnit: user.NeighborhoodCommunityUnit,
		Village:                   user.Village,
		District:                  user.District,
		Religion:                  user.Religion,
		MaritalStatus:             user.MaritalStatus,
		Profession:                user.Profession,
		Citizenship:               user.Citizenship,
		CardValidUntil:            user.CardValidUntil,
		PhoneNumber:               user.PhoneNumber,
		ProfilePhotoURL:           profilePhotoURL,
		IsVerified:                user.IsVerified,
		CreatedAt:                 user.CreatedAt,
		UpdatedAt:                 user.UpdatedAt,
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, response)
	}
}

func (h *AuthHandler) HandleGetProfilePhoto(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get profile photo request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	repo, err := h.authService.GetRepository().NewClient(false)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_repository")
	}

	user, err := repo.Users.GetByID(c, userData.ID)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_user_by_id")
	}

	profilePhotoURL := user.ProfilePhotoURL
	if profilePhotoURL != "" {
		presignedURL, err := h.s3Client.PresignUrl(profilePhotoURL)
		if err != nil {
			h.log.WithFields(log.Fields{
				"request_id": requestID,
				"error":      err.Error(),
				"path":       ctx.Path(),
			}).Warn("Failed to presign URL for profile photo")
		} else {
			profilePhotoURL = presignedURL
		}
	}

	result := auth.ProfilePhotoResponse{
		ID:              user.ID,
		ProfilePhotoURL: profilePhotoURL,
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}

func (h *AuthHandler) HandleUpdateFacePhoto(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 30*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing update face photo request")

	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	file, err := ctx.FormFile("photo")
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_form_file")
	}

	err = h.authService.User().UpdateFacePhoto(c, userData.ID, file)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_face_photo")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, nil)
	}
}
