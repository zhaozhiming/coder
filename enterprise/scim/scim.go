package scim

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/imulab/go-scim/pkg/v2/handlerutil"
	scimjson "github.com/imulab/go-scim/pkg/v2/json"
	"github.com/imulab/go-scim/pkg/v2/service"
	"github.com/imulab/go-scim/pkg/v2/spec"

	"cdr.dev/slog"
	"github.com/coder/coder/coderd"
	"github.com/coder/coder/coderd/database"
	"github.com/coder/coder/coderd/httpapi"
	agplscim "github.com/coder/coder/coderd/scim"
	"github.com/coder/coder/codersdk"
)

func NewHandler(
	logger slog.Logger,
	db database.Store,
	createUser func(ctx context.Context, store database.Store, req coderd.CreateUserRequest) (database.User, uuid.UUID, error),
	scimAPIKey []byte,
) agplscim.Handler {
	r := chi.NewRouter()
	h := &handler{
		Handler: r,

		log:        logger,
		db:         db,
		createUser: createUser,
		scimAPIKey: scimAPIKey,
	}

	r.Post("/Users", h.postUser)
	r.Route("/Users", func(r chi.Router) {
		r.Get("/", h.getUsers)
		r.Post("/", h.postUser)
		r.Get("/{id}", h.getUser)
		r.Patch("/{id}", h.patchUser)
	})

	return h
}

var _ agplscim.Handler = handler{}

type handler struct {
	http.Handler

	db  database.Store
	log slog.Logger

	createUser func(ctx context.Context, store database.Store, req coderd.CreateUserRequest) (database.User, uuid.UUID, error)

	scimAPIKey []byte
}

func (s *handler) verifyAuthHeader(r *http.Request) bool {
	hdr := []byte(r.Header.Get("Authorization"))

	return len(s.scimAPIKey) != 0 && subtle.ConstantTimeCompare(hdr, s.scimAPIKey) == 1
}

// getUsers intentionally always returns no users. This is done to always force
// Okta to try and create each user individually, this way we don't need to
// implement fetching users twice.
//
//nolint:revive
func (s *handler) getUsers(w http.ResponseWriter, r *http.Request) {
	if !s.verifyAuthHeader(r) {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusUnauthorized, Type: "invalidAuthorization"})
		return
	}

	_ = handlerutil.WriteSearchResultToResponse(w, &service.QueryResponse{
		TotalResults: 0,
		StartIndex:   1,
		ItemsPerPage: 0,
		Resources:    []scimjson.Serializable{},
	})
}

// getUser intentionally always returns an error saying the user wasn't found.
// This is done to always force Okta to try and create the user, this way we
// don't need to implement fetching users twice.
//
//nolint:revive
func (s *handler) getUser(w http.ResponseWriter, r *http.Request) {
	if !s.verifyAuthHeader(r) {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusUnauthorized, Type: "invalidAuthorization"})
		return
	}

	_ = handlerutil.WriteError(w, spec.ErrNotFound)
}

// We currently use our own struct instead of using the SCIM package. This was
// done mostly because the SCIM package was almost impossible to use. We only
// need these fields, so it was much simpler to use our own struct. This was
// tested only with Okta.
type scimUser struct {
	Schemas  []string `json:"schemas"`
	ID       string   `json:"id"`
	UserName string   `json:"userName"`
	Name     struct {
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"name"`
	Emails []struct {
		Primary bool   `json:"primary"`
		Value   string `json:"value"`
		Type    string `json:"type"`
		Display string `json:"display"`
	} `json:"emails"`
	Active bool          `json:"active"`
	Groups []interface{} `json:"groups"`
	Meta   struct {
		ResourceType string `json:"resourceType"`
	} `json:"meta"`
}

// postUser creates a new user, or returns the existing user if it exists.
func (s *handler) postUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !s.verifyAuthHeader(r) {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusUnauthorized, Type: "invalidAuthorization"})
		return
	}

	var sUser scimUser
	err := json.NewDecoder(r.Body).Decode(&sUser)
	if err != nil {
		_ = handlerutil.WriteError(w, err)
		return
	}

	email := ""
	for _, e := range sUser.Emails {
		if e.Primary {
			email = e.Value
			break
		}
	}

	if email == "" {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusBadRequest, Type: "invalidEmail"})
		return
	}

	user, _, err := s.createUser(ctx, s.db, coderd.CreateUserRequest{
		CreateUserRequest: codersdk.CreateUserRequest{
			Username: sUser.UserName,
			Email:    email,
		},
		LoginType: database.LoginTypeOIDC,
	})
	if err != nil {
		_ = handlerutil.WriteError(w, err)
		return
	}

	sUser.ID = user.ID.String()
	sUser.UserName = user.Username

	httpapi.Write(w, http.StatusOK, sUser)
}

// patchUser supports suspending and activating users only.
func (s *handler) patchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if !s.verifyAuthHeader(r) {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusUnauthorized, Type: "invalidAuthorization"})
		return
	}

	id := chi.URLParam(r, "id")

	var sUser scimUser
	err := json.NewDecoder(r.Body).Decode(&sUser)
	if err != nil {
		_ = handlerutil.WriteError(w, err)
		return
	}
	sUser.ID = id

	uid, err := uuid.Parse(id)
	if err != nil {
		_ = handlerutil.WriteError(w, spec.Error{Status: http.StatusBadRequest, Type: "invalidId"})
		return
	}

	dbUser, err := s.db.GetUserByID(ctx, uid)
	if err != nil {
		_ = handlerutil.WriteError(w, err)
		return
	}

	var status database.UserStatus
	if sUser.Active {
		status = database.UserStatusActive
	} else {
		status = database.UserStatusSuspended
	}

	_, err = s.db.UpdateUserStatus(r.Context(), database.UpdateUserStatusParams{
		ID:        dbUser.ID,
		Status:    status,
		UpdatedAt: database.Now(),
	})
	if err != nil {
		_ = handlerutil.WriteError(w, err)
		return
	}

	httpapi.Write(w, http.StatusOK, sUser)
}
