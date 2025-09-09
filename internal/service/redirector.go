package service

import (
	"net/http"
	"path"
)

type Redirector struct {
	store URLStore
}

func NewRedirector(store URLStore) *Redirector {
	return &Redirector{store}
}

func (rd *Redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shortCode, exists := rd.store.GetOriginalURL(path.Base(r.URL.Path))
	if !exists {
		errResponse := NewErrorResponse(http.StatusNotFound, ERR_SHORT_CODE_NOT_FOUND, ERR_SHORT_CODE_NOT_FOUND_CODE, ERR_SHORT_CODE_NOT_FOUND_DETAILS)
		errResponse.WriteError(w)
	}
	w.WriteHeader(http.StatusFound)
	http.Redirect(w, r, shortCode, http.StatusFound)
}
