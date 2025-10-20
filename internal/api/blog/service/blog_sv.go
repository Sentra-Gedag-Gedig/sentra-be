package blogService

import (
	"ProjectGolang/internal/api/blog"
	"ProjectGolang/internal/entity"
	contextPkg "ProjectGolang/pkg/context"
	"errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"
)

func (s *blogsService) CreateBlog(ctx context.Context, req blogs.CreateBlogRequest, userID string, imageFile *multipart.FileHeader) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	
	_, err = repo.Categories.GetCategoryByID(ctx, req.BlogCategory)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"category_id": req.BlogCategory,
			"error":       err.Error(),
		}).Warn("Blog category not found")
		return blogs.ErrCategoryNotFound
	}

	var imageURL string
	if imageFile != nil {
		
		if err := s.validateImageFile(imageFile); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Invalid image file")
			return err
		}

		
		uploadedURL, err := s.s3Client.UploadFile(imageFile)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to upload image")
			return blogs.ErrFailedToUpload
		}

		imageURL = uploadedURL
	}

	blogID, err := s.utils.NewULIDFromTimestamp(time.Now())
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to generate ULID")
		return err
	}

	now := time.Now()

	blog := entity.Blog{
		ID:           blogID,
		Title:        req.Title,
		Body:         req.Body,
		ImageURL:     imageURL,
		Author:       userID,
		BlogCategory: req.BlogCategory,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := repo.Blogs.CreateBlog(ctx, blog); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create blog")
		return blogs.ErrCreateBlog
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return blogs.ErrCreateBlog
	}

	return nil
}

func (s *blogsService) GetBlogByID(ctx context.Context, id string) (entity.Blog, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return entity.Blog{}, err
	}

	blog, err := repo.Blogs.GetBlogByID(ctx, id)
	if err != nil {
		if errors.Is(err, blogs.ErrBlogNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
			}).Warn("Blog not found")
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			}).Error("Failed to get blog")
		}
		return entity.Blog{}, err
	}

	
	if blog.ImageURL != "" {
		presignedURL, err := s.s3Client.PresignUrl(blog.ImageURL)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
				"image_url":  blog.ImageURL,
				"error":      err.Error(),
			}).Warn("Failed to create presigned URL for image")
			
		} else {
			blog.ImageURL = presignedURL
		}
	}

	return blog, nil
}

func (s *blogsService) GetAllBlogs(ctx context.Context, page, limit int) (*blogs.BlogListResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	blogsList, total, err := repo.Blogs.GetAllBlogs(ctx, limit, offset)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"page":       page,
			"limit":      limit,
			"error":      err.Error(),
		}).Error("Failed to get blogs")
		return nil, err
	}

	response := &blogs.BlogListResponse{
		Blogs: make([]blogs.BlogResponse, 0, len(blogsList)),
		Total: total,
	}

	for _, blog := range blogsList {
		
		imageURL := blog.ImageURL
		if imageURL != "" {
			presignedURL, err := s.s3Client.PresignUrl(imageURL)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"id":         blog.ID,
					"image_url":  imageURL,
					"error":      err.Error(),
				}).Warn("Failed to create presigned URL for image")
				
			} else {
				imageURL = presignedURL
			}
		}

		response.Blogs = append(response.Blogs, blogs.BlogResponse{
			ID:           blog.ID,
			Title:        blog.Title,
			Body:         blog.Body,
			ImageURL:     imageURL,
			Author:       blog.Author,
			BlogCategory: blog.BlogCategory,
			CreatedAt:    blog.CreatedAt,
			UpdatedAt:    blog.UpdatedAt,
		})
	}

	return response, nil
}

func (s *blogsService) GetBlogsByCategory(ctx context.Context, category string, page, limit int) (*blogs.BlogListResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	
	_, err = repo.Categories.GetCategoryByID(ctx, category)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"category_id": category,
			"error":       err.Error(),
		}).Warn("Blog category not found")
		return nil, blogs.ErrCategoryNotFound
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	blogsList, total, err := repo.Blogs.GetBlogsByCategory(ctx, category, limit, offset)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"category":   category,
			"page":       page,
			"limit":      limit,
			"error":      err.Error(),
		}).Error("Failed to get blogs by category")
		return nil, err
	}

	response := &blogs.BlogListResponse{
		Blogs: make([]blogs.BlogResponse, 0, len(blogsList)),
		Total: total,
	}

	for _, blog := range blogsList {
		
		imageURL := blog.ImageURL
		if imageURL != "" {
			presignedURL, err := s.s3Client.PresignUrl(imageURL)
			if err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"id":         blog.ID,
					"image_url":  imageURL,
					"error":      err.Error(),
				}).Warn("Failed to create presigned URL for image")
				
			} else {
				imageURL = presignedURL
			}
		}

		response.Blogs = append(response.Blogs, blogs.BlogResponse{
			ID:           blog.ID,
			Title:        blog.Title,
			Body:         blog.Body,
			ImageURL:     imageURL,
			Author:       blog.Author,
			BlogCategory: blog.BlogCategory,
			CreatedAt:    blog.CreatedAt,
			UpdatedAt:    blog.UpdatedAt,
		})
	}

	return response, nil
}

func (s *blogsService) UpdateBlog(ctx context.Context, id string, req blogs.UpdateBlogRequest, userID string, imageFile *multipart.FileHeader) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	
	existingBlog, err := repo.Blogs.GetBlogByID(ctx, id)
	if err != nil {
		if errors.Is(err, blogs.ErrBlogNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
			}).Warn("Blog not found")
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			}).Error("Failed to get blog")
		}
		return err
	}

	
	if existingBlog.Author != userID {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"id":           id,
			"blog_author":  existingBlog.Author,
			"request_user": userID,
		}).Warn("User is not the author of the blog")
		return blogs.ErrBlogNotOwned
	}

	
	if req.BlogCategory != "" && req.BlogCategory != existingBlog.BlogCategory {
		_, err = repo.Categories.GetCategoryByID(ctx, req.BlogCategory)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id":  requestID,
				"category_id": req.BlogCategory,
				"error":       err.Error(),
			}).Warn("Blog category not found")
			return blogs.ErrCategoryNotFound
		}
	}

	imageURL := existingBlog.ImageURL

	
	if imageFile != nil {
		
		if err := s.validateImageFile(imageFile); err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Warn("Invalid image file")
			return err
		}

		
		uploadedURL, err := s.s3Client.UploadFile(imageFile)
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      err.Error(),
			}).Error("Failed to upload image")
			return blogs.ErrFailedToUpload
		}

		
		if existingBlog.ImageURL != "" {
			parts := strings.Split(existingBlog.ImageURL, "/")
			if len(parts) > 0 {
				fileName := parts[len(parts)-1]
				go func() {
					if err := s.s3Client.DeleteFile(fileName); err != nil {
						s.log.WithFields(logrus.Fields{
							"request_id": requestID,
							"file_name":  fileName,
							"error":      err.Error(),
						}).Warn("Failed to delete old image")
					}
				}()
			}
		}

		imageURL = uploadedURL
	} else if req.ImageURL != "" {
		
		if req.ImageURL == "remove" && existingBlog.ImageURL != "" {
			
			parts := strings.Split(existingBlog.ImageURL, "/")
			if len(parts) > 0 {
				fileName := parts[len(parts)-1]
				if err := s.s3Client.DeleteFile(fileName); err != nil {
					s.log.WithFields(logrus.Fields{
						"request_id": requestID,
						"file_name":  fileName,
						"error":      err.Error(),
					}).Warn("Failed to delete old image")
				}
			}
			imageURL = ""
		}
	}

	
	blog := entity.Blog{
		ID:           id,
		Title:        req.Title,
		Body:         req.Body,
		ImageURL:     imageURL,
		BlogCategory: req.BlogCategory,
		UpdatedAt:    time.Now(),
	}

	if err := repo.Blogs.UpdateBlog(ctx, blog); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Error("Failed to update blog")
		return blogs.ErrUpdateBlog
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return blogs.ErrUpdateBlog
	}

	return nil
}

func (s *blogsService) DeleteBlog(ctx context.Context, id string, userID string) error {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(true)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return err
	}
	defer repo.Rollback()

	
	existingBlog, err := repo.Blogs.GetBlogByID(ctx, id)
	if err != nil {
		if errors.Is(err, blogs.ErrBlogNotFound) {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
			}).Warn("Blog not found")
		} else {
			s.log.WithFields(logrus.Fields{
				"request_id": requestID,
				"id":         id,
				"error":      err.Error(),
			}).Error("Failed to get blog")
		}
		return err
	}

	
	if existingBlog.Author != userID {
		s.log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"id":           id,
			"blog_author":  existingBlog.Author,
			"request_user": userID,
		}).Warn("User is not the author of the blog")
		return blogs.ErrBlogNotOwned
	}

	
	if err := repo.Blogs.DeleteBlog(ctx, id); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Error("Failed to delete blog")
		return blogs.ErrDeleteBlog
	}

	
	if existingBlog.ImageURL != "" {
		parts := strings.Split(existingBlog.ImageURL, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			if err := s.s3Client.DeleteFile(fileName); err != nil {
				s.log.WithFields(logrus.Fields{
					"request_id": requestID,
					"file_name":  fileName,
					"error":      err.Error(),
				}).Warn("Failed to delete image")
				
			}
		}
	}

	if err := repo.Commit(); err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to commit transaction")
		return blogs.ErrDeleteBlog
	}

	return nil
}

func (s *blogsService) GetAllCategories(ctx context.Context) (*blogs.CategoryListResponse, error) {
	requestID := contextPkg.GetRequestID(ctx)

	repo, err := s.blogsRepo.NewClient(false)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to create repository client")
		return nil, err
	}

	categories, err := repo.Categories.GetAllCategories(ctx)
	if err != nil {
		s.log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Error("Failed to get categories")
		return nil, err
	}

	response := &blogs.CategoryListResponse{
		Categories: make([]blogs.CategoryResponse, 0, len(categories)),
	}

	for _, category := range categories {
		response.Categories = append(response.Categories, blogs.CategoryResponse{
			ID:        category.ID,
			Name:      category.Name,
			CreatedAt: category.CreatedAt,
		})
	}

	return response, nil
}

func (s *blogsService) validateImageFile(file *multipart.FileHeader) error {
	if file == nil {
		return nil
	}

	
	maxSize := int64(5 * 1024 * 1024)
	if file.Size > maxSize {
		return blogs.ErrFileTooLarge
	}

	
	ext := filepath.Ext(file.Filename)
	ext = strings.ToLower(ext)
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !allowedExtensions[ext] {
		return blogs.ErrInvalidFileType
	}

	
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return blogs.ErrInvalidFileType
	}

	return nil
}
