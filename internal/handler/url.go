package handler

import (
	"net/http"
	"url_shortener/internal/handler/request"
	"url_shortener/internal/handler/response"
)

type Url struct {
	urlService UrlService
	ipService  IpService
}

func NewUrl(urlService UrlService, ipService IpService) *Url {
	return &Url{
		urlService: urlService,
		ipService:  ipService,
	}
}

// Create godoc
// @Summary      create url
// @Tags         url
// @Accept       json
// @Produce      json
// @Param        input  body  request.CreateUrl  true "create url"
// @Success      201    {object}  response.Url
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /urls [post]
func (u *Url) Create(w http.ResponseWriter, r *http.Request) {
	var request request.CreateUrl
	err := decodeAndValidate(&request, r.Body)
	if err != nil {
		errorResponse(w, err)
		return
	}
	url, err := u.urlService.Create(
		r.Context(),
		request.LongUrl,
		u.ipService.GetIp(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		errorResponse(w, err)
		return
	}
	successResponse(w, http.StatusCreated, response.NewUrl(url))
}

// Stats godoc
// @Summary      url stats
// @Tags         url
// @Produce      json
// @Param        short_code  path  string  true "short code"
// @Success      200    {object}  response.UrlStats
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /urls/{short_code} [get]
func (u *Url) Stats(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.GetStats(
		r.Context(),
		r.PathValue("short_code"),
	)
	if err != nil {
		errorResponse(w, err)
		return
	}
	successResponse(w, http.StatusOK, response.NewUrlStats(url))
}

// Click godoc
// @Summary      click short code
// @Tags         url
// @Param        short_code  path  string  true "short code"
// @Success      302
// @Failure      400    {object}  response.Error
// @Failure      500    {object}  response.Error
// @Router       /{short_code} [get]
func (u *Url) Click(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.Click(
		r.Context(),
		r.PathValue("short_code"),
		u.ipService.GetIp(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		errorResponse(w, err)
		return
	}
	http.Redirect(w, r, url.LongUrl, http.StatusFound)
}
