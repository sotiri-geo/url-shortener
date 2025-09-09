package handler

import (
	"net/http"
	"path"

	"github.com/sotiri-geo/url-shortener/internal/storage"
)

type Redirector struct {
	store storage.URLStore
}

func NewRedirector(store storage.URLStore) *Redirector {
	return &Redirector{store}
}

func (rd *Redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalURL, exists := rd.store.GetOriginalURL(path.Base(r.URL.Path))
	if !exists {
		errResponse := NewErrorResponse(http.StatusNotFound, ERR_SHORT_CODE_NOT_FOUND, ERR_SHORT_CODE_NOT_FOUND_CODE, ERR_SHORT_CODE_NOT_FOUND_DETAILS)
		errResponse.WriteError(w)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}
