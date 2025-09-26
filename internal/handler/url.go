package handler

import (
	"net/http"
	"url_shortener/internal/handler/helper"
	"url_shortener/internal/handler/request"
)

type URL struct {
	urlService URLService
}

func NewURL(
	urlService URLService,
) *URL {
	return &URL{
		urlService: urlService,
	}
}

// Create godoc
//
//	@Summary	create url
//	@Tags		url
//	@Accept		json
//	@Produce	json
//	@Param		input	body		request.CreateURL	true	"create url"
//	@Success	201		{object}	response.Ok{data=model.URL}
//	@Failure	400		{object}	response.Fail
//	@Failure	500		{object}	response.Fail
//	@Router		/urls [post].
func (u *URL) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateURL
	err := helper.ParseJSON(&req, r.Body)
	if err != nil {
		helper.Fail(w, err)
		return
	}

	url, err := u.urlService.Create(
		r.Context(),
		req.OriginalURL,
	)
	if err != nil {
		helper.Fail(w, err)
		return
	}

	helper.Ok(w, http.StatusCreated, url)
}

// Redirect godoc
//
//	@Summary	redirect to url
//	@Tags		url
//	@Param		short_code	path	string	true	"short code"
//	@Success	302
//	@Failure	400	{object}	response.Fail
//	@Failure	500	{object}	response.Fail
//	@Router		/{short_code} [get].
func (u *URL) Redirect(w http.ResponseWriter, r *http.Request) {
	original, err := u.urlService.GetOriginalURL(
		r.Context(),
		r.PathValue("short_code"),
	)
	if err != nil {
		helper.Fail(w, err)
		return
	}

	http.Redirect(w, r, original, http.StatusFound)
}
