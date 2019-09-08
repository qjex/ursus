package api

import (
	"github.com/go-chi/render"
	"log"
	"net/http"
)

type Error struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

func SendErrorJSON(w http.ResponseWriter, r *http.Request, httpStatusCode int, err error, details string) {
	log.Printf("Error processing request with errorCode=%d, error=%v, details=%s", httpStatusCode, err, details)
	render.Status(r, httpStatusCode)
	render.JSON(w, r, Error{err.Error(), details})
}
