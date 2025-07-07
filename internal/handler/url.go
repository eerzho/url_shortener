package handler

import (
	"net/http"
	"url_shortener/internal/constant"
	"url_shortener/internal/handler/request"
	"url_shortener/internal/handler/response"
)

// Additional URL-related handlers that complement the main handler functions

// urlBatch godoc
// @Summary      Create multiple short URLs
// @Description  Create multiple short URLs in a single request
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input  body      []request.CreateUrl  true  "Batch URL creation request"
// @Success      201    {object}  []response.Url
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /urls/batch [post]
func urlBatch(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requests []request.CreateUrl

		if err := decodeAndValidate(&requests, r.Body); err != nil {
			errorResponse(w, r, err)
			return
		}

		// Validate batch size
		if len(requests) == 0 {
			errorResponse(w, r, &ValidationError{
				Field:   "batch",
				Message: "batch cannot be empty",
			})
			return
		}

		if len(requests) > 100 {
			errorResponse(w, r, &ValidationError{
				Field:   "batch",
				Message: "batch size cannot exceed 100",
			})
			return
		}

		results := make([]response.Url, 0, len(requests))
		for _, req := range requests {
			url, err := urlService.Create(r.Context(), req.LongUrl)
			if err != nil {
				errorResponse(w, r, err)
				return
			}
			results = append(results, *response.NewUrl(url))
		}

		successResponse(w, http.StatusCreated, results)
	}
}

// urlDelete godoc
// @Summary      Delete a short URL
// @Description  Delete a short URL by short code
// @Tags         url
// @Param        short_code  path  string  true  "Short code"
// @Success      204         "No Content"
// @Failure      400         {object}  response.Error
// @Failure      404         {object}  response.Error
// @Failure      500         {object}  response.Error
// @Router       /urls/{short_code} [delete]
func urlDelete(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := r.PathValue("short_code")

		if shortCode == "" {
			errorResponse(w, r, &ValidationError{
				Field:   "short_code",
				Message: "short code is required",
			})
			return
		}

		// First check if URL exists
		_, err := urlService.GetByShortCode(r.Context(), shortCode)
		if err != nil {
			if err == constant.ErrNotFound {
				errorResponse(w, r, constant.ErrNotFound)
			} else {
				errorResponse(w, r, err)
			}
			return
		}

		// Note: Delete functionality would need to be implemented in the service layer
		// For now, we'll return a not implemented error
		errorResponse(w, r, &ValidationError{
			Field:   "operation",
			Message: "delete operation not implemented",
		})
	}
}

// urlList godoc
// @Summary      List URLs
// @Description  Get a paginated list of URLs
// @Tags         url
// @Produce      json
// @Param        page     query     int     false  "Page number"     default(1)
// @Param        per_page query     int     false  "Items per page"  default(10)
// @Success      200      {object}  response.UrlList
// @Failure      400      {object}  response.Error
// @Failure      500      {object}  response.Error
// @Router       /urls [get]
func urlList(urlService UrlService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Note: List functionality would need to be implemented in the service layer
		// For now, we'll return a not implemented error
		errorResponse(w, r, &ValidationError{
			Field:   "operation",
			Message: "list operation not implemented",
		})
	}
}
