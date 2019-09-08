package api

import (
	"github.com/go-chi/render"
	"net/http"
	"strconv"
	"ursus/store"
)

type public struct {
	store store.ProxyStore
}

type ProxyList struct {
	Proxies []store.Proxy `json:"proxies"`
}

const defaultPageSize int64 = 10

func (s *public) getProxyList(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 0
	}
	result, err := s.store.FindAll(r.Context(), int64(page), defaultPageSize)

	if err != nil {
		SendErrorJSON(w, r, http.StatusBadRequest, err, "error retrieving data from storage")
		return
	}

	render.JSON(w, r, ProxyList{result})
}
