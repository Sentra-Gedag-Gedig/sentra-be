package blogHandler

import (
	blogsService "ProjectGolang/internal/api/blog/service"
	"ProjectGolang/internal/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type BlogsHandler struct {
	log          *logrus.Logger
	validator    *validator.Validate
	middleware   middleware.Middleware
	blogsService blogsService.IBlogsService
}

func New(
	log *logrus.Logger,
	validate *validator.Validate,
	middleware middleware.Middleware,
	bs blogsService.IBlogsService,
) *BlogsHandler {
	return &BlogsHandler{
		log:          log,
		validator:    validate,
		middleware:   middleware,
		blogsService: bs,
	}
}

func (h *BlogsHandler) Start(srv fiber.Router) {
	blogs := srv.Group("/blogs")

	// Create blog (requires auth)
	blogs.Post("/", h.middleware.NewTokenMiddleware, h.CreateBlog)

	// Public endpoints (no auth required)
	blogs.Get("", h.GetAllBlogs)
	blogs.Get("/categories", h.GetAllCategories)
	blogs.Get("/category/:id", h.GetBlogsByCategory)
	blogs.Get("/:id", h.GetBlogByID)

	// Update and delete (requires auth)
	blogs.Put("/:id", h.middleware.NewTokenMiddleware, h.UpdateBlog)
	blogs.Delete("/:id", h.middleware.NewTokenMiddleware, h.DeleteBlog)
}
