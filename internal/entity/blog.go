package entity

import "time"

type Blog struct {
	ID           string    `db:"id"`
	Title        string    `db:"title"`
	Body         string    `db:"body"`
	ImageURL     string    `db:"image_url"`
	Author       string    `db:"author"`
	BlogCategory string    `db:"blog_category"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type BlogCategory struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}
