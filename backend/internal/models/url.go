package models

type URLRequest struct {
	// validate:"required,url" ensures that the field is required and must be a valid URL
	URL string `json:"long_url" validate:"required,url"`
}

type URLResponse struct {
	ShortenedURL string `json:"shortened_url"`
}