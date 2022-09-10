package scim

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/coder/coder/coderd/features"
	"github.com/coder/coder/coderd/httpapi"
)

type Handler interface {
	http.Handler
}

func NewNop() Handler {
	return nop{Router: chi.NewRouter()}
}

type nop struct {
	chi.Router
}

func Mount(feats features.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scimFeat := struct {
			SCIM Handler
		}{}
		err := feats.Get(&scimFeat)
		if err != nil {
			httpapi.InternalServerError(w, err)
			return
		}

		scimFeat.SCIM.ServeHTTP(w, r)
	})
}
