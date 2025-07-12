package handler

import (
	"net/http"
	"url_shortener/internal/handler/request"
)

type Url struct {
	*Handler
	urlService UrlService
	ipService  IpService
}

func NewUrl(
	handler *Handler,
	urlService UrlService,
	ipService IpService,
) *Url {
	return &Url{
		Handler:    handler,
		urlService: urlService,
		ipService:  ipService,
	}
}

// Create godoc
// @Summary    create url
// @Tags       url
// @Accept     json
// @Produce    json
// @Param      input body request.CreateUrl true "create url"
// @Success    201 {object} response.Ok{data=model.Url}
// @Failure    400 {object} response.Fail
// @Failure    500 {object} response.Fail
// @Router     /urls [post]
func (u *Url) Create(w http.ResponseWriter, r *http.Request) {
	var request request.CreateUrl
	err := u.parseJson(&request, r.Body)
	if err != nil {
		u.fail(w, err)
		return
	}
	url, err := u.urlService.Create(
		r.Context(),
		request.LongUrl,
		u.ipService.GetIp(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		u.fail(w, err)
		return
	}
	u.ok(w, http.StatusCreated, url)
}

// Stats godoc
// @Summary   url stats
// @Tags      url
// @Produce   json
// @Param     short_code path string true "short code"
// @Success   200 {object} response.Ok{data=model.UrlWithClicksCount}
// @Failure   400 {object} response.Fail
// @Failure   500 {object} response.Fail
// @Router    /urls/{short_code} [get]
func (u *Url) Stats(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.GetStats(
		r.Context(),
		r.PathValue("short_code"),
	)
	if err != nil {
		u.fail(w, err)
		return
	}
	u.ok(w, http.StatusOK, url)
}

// Click godoc
// @Summary   click short code
// @Tags      url
// @Param     short_code path string true "short code"
// @Success   302
// @Failure   400 {object} response.Fail
// @Failure   500 {object} response.Fail
// @Router    /{short_code} [get]
func (u *Url) Click(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.Click(
		r.Context(),
		r.PathValue("short_code"),
		u.ipService.GetIp(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		u.fail(w, err)
		return
	}
	http.Redirect(w, r, url.LongUrl, http.StatusFound)
}
