package blogService

import (
	"ProjectGolang/internal/api/blog"
	blogsRepository "ProjectGolang/internal/api/blog/repository"
	"ProjectGolang/internal/entity"
	"ProjectGolang/pkg/s3"
	"ProjectGolang/pkg/utils"
	"context"
	"github.com/sirupsen/logrus"
	"mime/multipart"
)

type IBlogsService interface {
	CreateBlog(ctx context.Context, req blogs.CreateBlogRequest, userID string, imageFile *multipart.FileHeader) error
	GetBlogByID(ctx context.Context, id string) (entity.Blog, error)
	GetAllBlogs(ctx context.Context, page, limit int) (*blogs.BlogListResponse, error)
	GetBlogsByCategory(ctx context.Context, category string, page, limit int) (*blogs.BlogListResponse, error)
	UpdateBlog(ctx context.Context, id string, req blogs.UpdateBlogRequest, userID string, imageFile *multipart.FileHeader) error
	DeleteBlog(ctx context.Context, id string, userID string) error
	GetAllCategories(ctx context.Context) (*blogs.CategoryListResponse, error)
}

type blogsService struct {
	log       *logrus.Logger
	blogsRepo blogsRepository.Repository
	s3Client  s3.ItfS3
	utils     utils.IUtils
}

func NewBlogsService(
	log *logrus.Logger,
	blogsRepo blogsRepository.Repository,
	s3Client s3.ItfS3,
	utils utils.IUtils,
) IBlogsService {
	return &blogsService{
		log:       log,
		blogsRepo: blogsRepo,
		s3Client:  s3Client,
		utils:     utils,
	}
}
