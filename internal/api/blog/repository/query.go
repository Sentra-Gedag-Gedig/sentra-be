package blogRepository

const (
	queryCreateBlog = `
		INSERT INTO blogs (
			id,
			title,
			body,
			image_url,
			author,
			blog_category,
			created_at,
			updated_at
		) VALUES (
			:id,
			:title,
			:body,
			:image_url,
			:author,
			:blog_category,
			:created_at,
			:updated_at
		)
	`

	queryGetBlogByID = `
		SELECT
			id,
			title,
			body,
			image_url,
			author,
			blog_category,
			created_at,
			updated_at
		FROM blogs
		WHERE id = :id
	`

	queryGetAllBlogs = `
		SELECT
			id,
			title,
			body,
			image_url,
			author,
			blog_category,
			created_at,
			updated_at
		FROM blogs
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryCountAllBlogs = `
		SELECT COUNT(*)
		FROM blogs
	`

	queryGetBlogsByCategory = `
		SELECT
			id,
			title,
			body,
			image_url,
			author,
			blog_category,
			created_at,
			updated_at
		FROM blogs
		WHERE blog_category = :blog_category
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	queryCountBlogsByCategory = `
		SELECT COUNT(*)
		FROM blogs
		WHERE blog_category = :blog_category
	`

	queryUpdateBlog = `
		UPDATE blogs
		SET
			title = CASE WHEN :title = '' THEN title ELSE :title END,
			body = CASE WHEN :body = '' THEN body ELSE :body END,
			image_url = CASE WHEN :image_url = '' THEN image_url ELSE :image_url END,
			blog_category = CASE WHEN :blog_category = '' THEN blog_category ELSE :blog_category END,
			updated_at = :updated_at
		WHERE id = :id
	`

	queryDeleteBlog = `
		DELETE FROM blogs
		WHERE id = :id
	`

	queryGetAllCategories = `
		SELECT
			id,
			name,
			created_at
		FROM blog_categories
		ORDER BY name ASC
	`

	queryGetCategoryByID = `
		SELECT
			id,
			name,
			created_at
		FROM blog_categories
		WHERE id = :id
	`

	queryGetCategoryByName = `
		SELECT
			id,
			name,
			created_at
		FROM blog_categories
		WHERE name = :name
	`
)
