package scim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cdr.dev/slog/sloggers/slogtest"
	"github.com/coder/coder/coderd"
	"github.com/coder/coder/coderd/database"
	"github.com/coder/coder/coderd/database/databasefake"
	"github.com/coder/coder/cryptorand"
)

func setupScimService(t *testing.T) *handler {
	t.Helper()

	log := slogtest.Make(t, nil)
	db := databasefake.New()

	return &handler{
		db:  db,
		log: log,
		createUser: func(ctx context.Context, store database.Store, req coderd.CreateUserRequest) (database.User, uuid.UUID, error) {
			now := database.Now()
			user, err := store.InsertUser(ctx, database.InsertUserParams{
				ID:        uuid.New(),
				Email:     req.Email,
				Username:  req.Username,
				CreatedAt: now,
				UpdatedAt: now,
				// All new users are defaulted to members of the site.
				RBACRoles: []string{},
				LoginType: req.LoginType,
			})

			return user, uuid.Nil, err
		},
		scimAPIKey: []byte("secret"),
	}
}

//nolint:revive
func makeScimUser(t testing.TB) *scimUser {
	rstr, err := cryptorand.String(10)
	require.NoError(t, err)

	return &scimUser{
		UserName: rstr,
		Name: struct {
			GivenName  string "json:\"givenName\""
			FamilyName string "json:\"familyName\""
		}{
			GivenName:  rstr,
			FamilyName: rstr,
		},
		Emails: []struct {
			Primary bool   "json:\"primary\""
			Value   string "json:\"value\""
			Type    string "json:\"type\""
			Display string "json:\"display\""
		}{
			{Primary: true, Value: fmt.Sprintf("%s@coder.com", rstr)},
		},
		Active: true,
	}
}

func doScimRequest(ctx context.Context, t testing.TB, svc *handler, fn func(w http.ResponseWriter, r *http.Request), body interface{}) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	require.NoError(t, err)

	chi.NewRouteContext()

	r := httptest.NewRequest("GET", "/", bytes.NewReader(raw))
	r.Header.Set("Authorization", string(svc.scimAPIKey))
	w := httptest.NewRecorder()

	fn(w, r.WithContext(ctx))
	return w
}

func TestScim(t *testing.T) {
	t.Parallel()

	t.Run("postUser", func(t *testing.T) {
		t.Parallel()

		t.Run("noAuth", func(t *testing.T) {
			t.Parallel()

			svc := setupScimService(t)
			r := httptest.NewRequest("GET", "/", bytes.NewReader(nil))
			w := httptest.NewRecorder()

			svc.postUser(w, r)

			require.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("OK", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := setupScimService(t)
			expectedUser := makeScimUser(t)

			res := doScimRequest(ctx, t, svc, svc.postUser, expectedUser)
			require.Equal(t, http.StatusOK, res.Code)

			gotUser := &scimUser{}
			err := json.Unmarshal(res.Body.Bytes(), gotUser)
			require.NoError(t, err)

			dbuser, err := svc.db.GetUserByID(ctx, uuid.MustParse(gotUser.ID))
			require.NoError(t, err)

			require.Equal(t, expectedUser.Emails[0].Value, dbuser.Email)
			require.Equal(t, expectedUser.UserName, dbuser.Username)
		})
	})

	t.Run("patchUser", func(t *testing.T) {
		t.Parallel()

		t.Run("noAuth", func(t *testing.T) {
			t.Parallel()

			svc := setupScimService(t)
			r := httptest.NewRequest("GET", "/", bytes.NewReader(nil))
			w := httptest.NewRecorder()

			svc.patchUser(w, r)

			require.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("OK", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			svc := setupScimService(t)
			expectedUser := makeScimUser(t)

			res := doScimRequest(ctx, t, svc, svc.postUser, expectedUser)
			require.Equal(t, http.StatusOK, res.Code)

			gotUser := &scimUser{}
			err := json.Unmarshal(res.Body.Bytes(), gotUser)
			require.NoError(t, err)

			expectedUser.Active = false

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", gotUser.ID)
			ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)

			res = doScimRequest(ctx, t, svc, svc.patchUser, expectedUser)
			require.Equal(t, http.StatusOK, res.Code)

			dbUser, err := svc.db.GetUserByID(ctx, uuid.MustParse(gotUser.ID))
			require.NoError(t, err)
			assert.Equal(t, database.UserStatusSuspended, dbUser.Status)
		})
	})
}
