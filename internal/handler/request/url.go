package request

type CreateUrl struct {
	LongUrl string `json:"long_url" validate:"required,url"`
}
