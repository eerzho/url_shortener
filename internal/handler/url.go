package handler

import (
	"net/http"
	"url_shortener/internal/handler/request"
	"url_shortener/internal/handler/response"
)

// urlCreate godoc
// @Summary      create short code
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input  body  request.CreateUrl  true "create short code"
// @Success      201    {object}  response.Url
// @Failure      400    {object}  response.Error
// @Router       /urls [post]
func urlCreate(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request request.CreateUrl
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
		successResponse(w, http.StatusCreated, response.NewUrl(url))
	}
}

// urlShow godoc
// @Summary      get url info
// @Tags         url
// @Produce      json
// @Param        short_code  path  string  true "short code"
// @Success      200    {object}  response.Url
// @Failure      400    {object}  response.Error
// @Router       /urls/{short_code} [get]
func urlShow(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCode(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}
		successResponse(w, http.StatusOK, response.NewUrl(url))
	}
}

// urlRedirect godoc
// @Summary      redirect to long url
// @Tags         url
// @Param        short_code  path  string  true "short code"
// @Success      302
// @Failure      400    {object}  response.Error
// @Router       /{short_code} [get]
func urlRedirect(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, err := urlService.GetByShortCodeAndIncrementClicks(r.Context(), r.PathValue("short_code"))
		if err != nil {
			errorResponse(w, err)
			return
		}
		http.Redirect(w, r, url.LongUrl, http.StatusFound)
	}
}
