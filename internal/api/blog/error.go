package blogs

import "ProjectGolang/pkg/response"

var (
	ErrBlogNotFound     = response.NewError(404, "blog not found")
	ErrCategoryNotFound = response.NewError(404, "blog category not found")
	ErrCreateBlog       = response.NewError(500, "failed to create blog")
	ErrUpdateBlog       = response.NewError(500, "failed to update blog")
	ErrDeleteBlog       = response.NewError(500, "failed to delete blog")
	ErrInvalidFileType  = response.NewError(400, "invalid file type")
	ErrFileTooLarge     = response.NewError(500, "file too large")
	ErrFailedToUpload   = response.NewError(500, "failed to upload file")
	ErrBlogNotOwned     = response.NewError(403, "blog does not belong to user")
	ErrInvalidBlogData  = response.NewError(400, "invalid blog data")
)
