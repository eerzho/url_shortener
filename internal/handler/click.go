package handler

import (
	"net/http"
	"strconv"
	"url_shortener/internal/handler/helper"
)

type Click struct {
	clickService ClickService
}

func NewClick(
	clickService ClickService,
) *Click {
	return &Click{
		clickService: clickService,
	}
}

// List godoc
//
//	@Summary	get clicks by url
//	@Tags		click
//	@Accept		json
//	@Produce	json
//	@Param		short_code	path		string	true	"short code"
//	@Success	200			{object}	response.List{data=[]model.Click}
//	@Failure	400			{object}	response.Fail
//	@Failure	500			{object}	response.Fail
//	@Router		/urls/{short_code}/clicks [get]
func (c *Click) List(w http.ResponseWriter, r *http.Request) {
	shortCode := r.PathValue("short_code")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	list, pagination, err := c.clickService.GetList(r.Context(), shortCode, page, size)
	if err != nil {
		helper.Fail(w, err)
		return
	}
	helper.List(w, list, pagination)
}
