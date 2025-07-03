package handler

import (
	"net/http"
	"url_shortener/internal/service"
)

func urlCreate(urlService service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			LongUrl string `json:"long_url" validate:"required,url"`
		}

		err := decodeAndValidate(&request, r.Body)
		if err != nil {
			errorResponse(w, err)
			return
		}

		url, err := urlService.Create(r.Context(), request.LongUrl)
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusCreated, url)
	}
}

func urlShow(urlService service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCode(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}

		successResponse(w, http.StatusOK, url)
	}
}

func urlRedirect(urlService service.Url) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCodeAndIncrementClicks(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}

		http.Redirect(w, r, url.LongUrl, http.StatusFound)
	}
}
