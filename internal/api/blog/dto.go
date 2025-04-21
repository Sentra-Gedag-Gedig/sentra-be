package blogs

import "time"

type CreateBlogRequest struct {
	Title        string `json:"title" validate:"required,min=3,max=256"`
	Body         string `json:"body" validate:"required"`
	BlogCategory string `json:"blog_category" validate:"required"`
}

type UpdateBlogRequest struct {
	Title        string `json:"title" validate:"omitempty,min=3,max=256"`
	Body         string `json:"body" validate:"omitempty"`
	BlogCategory string `json:"blog_category" validate:"omitempty"`
	ImageURL     string `json:"image_url" validate:"omitempty"`
}

type BlogResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	ImageURL     string    `json:"image_url"`
	Author       string    `json:"author"`
	BlogCategory string    `json:"blog_category"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type BlogListResponse struct {
	Blogs []BlogResponse `json:"blogs"`
	Total int            `json:"total"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}
