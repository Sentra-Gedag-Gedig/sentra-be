package blogHandler

import (
	"ProjectGolang/internal/api/blog"
	contextPkg "ProjectGolang/pkg/context"
	"ProjectGolang/pkg/handlerUtil"
	jwtPkg "ProjectGolang/pkg/jwt"
	"ProjectGolang/pkg/log"
	"errors"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
	"strconv"
	"time"
)

func (h *BlogsHandler) CreateBlog(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing create blog request")

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	// Get form values
	title := ctx.FormValue("title")
	body := ctx.FormValue("body")
	blogCategory := ctx.FormValue("blog_category")

	// Validate required fields
	if title == "" || body == "" || blogCategory == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("title, body, and blog_category are required"), ctx.Path())
	}

	// Create request object
	req := blogs.CreateBlogRequest{
		Title:        title,
		Body:         body,
		BlogCategory: blogCategory,
	}

	// Validate with validator
	if err := h.validator.Struct(req); err != nil {
		return errHandler.HandleValidationError(ctx, requestID, err, ctx.Path())
	}

	// Get image file if provided
	imageFile, err := ctx.FormFile("image")
	// Ignore error - image is optional

	// Create blog
	if err := h.blogsService.CreateBlog(c, req, userData.ID, imageFile); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "create_blog")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusCreated, fiber.Map{
			"message": "Blog created successfully",
		})
	}
}

func (h *BlogsHandler) GetBlogByID(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get blog by ID request")

	id := ctx.Params("id")
	if id == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("blog ID is required"), ctx.Path())
	}

	blog, err := h.blogsService.GetBlogByID(c, id)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_blog")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, blogs.BlogResponse{
			ID:           blog.ID,
			Title:        blog.Title,
			Body:         blog.Body,
			ImageURL:     blog.ImageURL,
			Author:       blog.Author,
			BlogCategory: blog.BlogCategory,
			CreatedAt:    blog.CreatedAt,
			UpdatedAt:    blog.UpdatedAt,
		})
	}
}

func (h *BlogsHandler) GetAllBlogs(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get all blogs request")

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.blogsService.GetAllBlogs(c, page, limit)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_all_blogs")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}

func (h *BlogsHandler) GetBlogsByCategory(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get blogs by category request")

	categoryID := ctx.Params("id")
	if categoryID == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("category ID is required"), ctx.Path())
	}

	page, err := strconv.Atoi(ctx.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	result, err := h.blogsService.GetBlogsByCategory(c, categoryID, page, limit)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_blogs_by_category")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}

func (h *BlogsHandler) UpdateBlog(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing update blog request")

	id := ctx.Params("id")
	if id == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("blog ID is required"), ctx.Path())
	}

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	// Get form values
	title := ctx.FormValue("title", "")
	body := ctx.FormValue("body", "")
	blogCategory := ctx.FormValue("blog_category", "")
	imageURL := ctx.FormValue("image_url", "")

	// Create request object with form values
	req := blogs.UpdateBlogRequest{
		Title:        title,
		Body:         body,
		BlogCategory: blogCategory,
		ImageURL:     imageURL,
	}

	// Get image file if provided
	imageFile, err := ctx.FormFile("image")
	// Ignore error - image is optional

	// Update blog
	if err := h.blogsService.UpdateBlog(c, id, req, userData.ID, imageFile); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "update_blog")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"message": "Blog updated successfully",
		})
	}
}

func (h *BlogsHandler) DeleteBlog(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing delete blog request")

	id := ctx.Params("id")
	if id == "" {
		return errHandler.HandleValidationError(ctx, requestID,
			errors.New("blog ID is required"), ctx.Path())
	}

	// Get authenticated user
	userData, err := jwtPkg.GetUserLoginData(ctx)
	if err != nil {
		return errHandler.HandleUnauthorized(ctx, requestID, "Unauthorized")
	}

	// Delete blog
	if err := h.blogsService.DeleteBlog(c, id, userData.ID); err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "delete_blog")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, fiber.Map{
			"message": "Blog deleted successfully",
		})
	}
}

func (h *BlogsHandler) GetAllCategories(ctx *fiber.Ctx) error {
	requestID := h.middleware.GetRequestID(ctx)
	c, cancel := context.WithTimeout(contextPkg.FromFiberCtx(ctx), 10*time.Second)
	defer cancel()

	errHandler := handlerUtil.New(h.log)

	h.log.WithFields(log.Fields{
		"request_id": requestID,
		"path":       ctx.Path(),
	}).Debug("Processing get all categories request")

	result, err := h.blogsService.GetAllCategories(c)
	if err != nil {
		return errHandler.Handle(ctx, requestID, err, ctx.Path(), "get_all_categories")
	}

	select {
	case <-c.Done():
		return errHandler.HandleRequestTimeout(ctx)
	default:
		return errHandler.HandleSuccess(ctx, fiber.StatusOK, result)
	}
}
