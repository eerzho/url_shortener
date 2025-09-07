package request

type CreateURL struct {
	OriginalURL string `json:"original_url" validate:"required,url"`
}
