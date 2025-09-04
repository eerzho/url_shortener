package handler

import (
	"net/http"
	"url_shortener/internal/handler/helper"
	"url_shortener/internal/handler/request"
)

type URL struct {
	urlService URLService
	ipService  IPService
}

func NewURL(
	urlService URLService,
	ipService IPService,
) *URL {
	return &URL{
		urlService: urlService,
		ipService:  ipService,
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
//	@Router		/urls [post]
func (u *URL) Create(w http.ResponseWriter, r *http.Request) {
	var request request.CreateURL
	err := helper.ParseJSON(&request, r.Body)
	if err != nil {
		helper.Fail(w, err)
		return
	}
	url, err := u.urlService.Create(
		r.Context(),
		request.LongURL,
		u.ipService.GetIP(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		helper.Fail(w, err)
		return
	}
	helper.OK(w, http.StatusCreated, url)
}

// Stats godoc
//
//	@Summary	url stats
//	@Tags		url
//	@Produce	json
//	@Param		short_code	path		string	true	"short code"
//	@Success	200			{object}	response.Ok{data=model.URLWithClicksCount}
//	@Failure	400			{object}	response.Fail
//	@Failure	500			{object}	response.Fail
//	@Router		/urls/{short_code} [get]
func (u *URL) Stats(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.GetStats(
		r.Context(),
		r.PathValue("short_code"),
	)
	if err != nil {
		helper.Fail(w, err)
		return
	}
	helper.OK(w, http.StatusOK, url)
}

// Click godoc
//
//	@Summary	click short code
//	@Tags		url
//	@Param		short_code	path	string	true	"short code"
//	@Success	302
//	@Failure	400	{object}	response.Fail
//	@Failure	500	{object}	response.Fail
//	@Router		/{short_code} [get]
func (u *URL) Click(w http.ResponseWriter, r *http.Request) {
	url, err := u.urlService.Click(
		r.Context(),
		r.PathValue("short_code"),
		u.ipService.GetIP(r.Context(), r),
		r.UserAgent(),
	)
	if err != nil {
		helper.Fail(w, err)
		return
	}
	http.Redirect(w, r, url.LongURL, http.StatusFound)
}
