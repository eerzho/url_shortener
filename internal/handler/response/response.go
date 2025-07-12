package response

import "url_shortener/internal/dto"

type Ok struct {
	Data any `json:"data"`
}

type List struct {
	Data       any             `json:"data"`
	Pagination *dto.Pagination `json:"pagination,omitempty"`
}

type Fail struct {
	Error  string   `json:"error,omitempty"`
	Errors []string `json:"errors,omitempty"`
}
