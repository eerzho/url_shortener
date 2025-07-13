package request

type CreateURL struct {
	LongURL string `json:"long_url" validate:"required,url,max=2048"`
}
