package models

import "time"

type Click struct {
	ID 		 	 int64     `json:"id"`
	ShortCode    string    `json:"short_code"`
	ClickedAt    time.Time `json:"clicked_at"`
}

type ClickRequest struct {
	ShortenedURL string `json:"shortened_url" validate:"required"`
}

type ClickResponse struct {
	ShortenedURL string `json:"shortened_url"`
	ClickedAt    string `json:"clicked_at"`
}
