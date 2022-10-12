package coderd

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"golang.org/x/xerrors"

	"github.com/coder/coder/coderd/database"
	"github.com/coder/coder/coderd/httpapi"
	"github.com/coder/coder/coderd/httpmw"
	"github.com/coder/coder/coderd/rbac"
	"github.com/coder/coder/codersdk"
)

func parseUUID(ctx context.Context, rw http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id := chi.URLParam(r, param)
	uid, err := uuid.Parse(id)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: fmt.Sprintf("Invalid UUID %q.", id),
			Detail:  err.Error(),
			Validations: []codersdk.ValidationError{
				{Field: param, Detail: "Invalid UUID"},
			},
		})
		return uuid.Nil, false
	}
	return uid, true
}

func parseGitAuthStatus(ctx context.Context, rw http.ResponseWriter, r *http.Request) (string, bool) {
	status := chi.URLParam(r, "status")
	switch status {
	case "resolved", "pending", "expired", "":
		return status, true
	default:
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Invalid status.",
			Validations: []codersdk.ValidationError{
				{Field: "status", Detail: "Invalid status."},
			},
		})
		return "", false
	}
}

func convertUserLinkRequestsToGitAuthRequests(reqs []database.UserLinkRequest, now time.Time) []codersdk.GitAuthRequestResponse {
	res := make([]codersdk.GitAuthRequestResponse, len(reqs))
	for i, req := range reqs {
		res[i] = convertUserLinkRequestToGitAuthRequest(req, now)
	}
	return res
}

func convertUserLinkRequestToGitAuthRequest(req database.UserLinkRequest, now time.Time) codersdk.GitAuthRequestResponse {
	return codersdk.GitAuthRequestResponse{
		ID:        req.ID,
		UserID:    req.UserID,
		AgentID:   req.AgentID,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
		ExpiresAt: req.ExpiresAt,
		Provider:  req.Provider,
		LoginURL:  req.LoginUrl,
		LoginUser: req.LoginUser,
		Resolved:  req.Resolved,
		Expired:   !req.Resolved && req.ExpiresAt.Before(now),
	}
}

func (api *API) gitAuthRequests(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := httpmw.UserParam(r)

	if !api.Authorize(r, rbac.ActionRead, rbac.ResourceUserData.WithOwner(user.ID.String())) {
		httpapi.ResourceNotFound(rw)
		return
	}

	status, ok := parseGitAuthStatus(ctx, rw, r)
	if !ok {
		return
	}

	// Reference now before query so that when expired is set, it is
	// consistent with the rows returned by the query.
	now := time.Now()

	ulReqs, err := api.Database.GetUserLinkRequestsByUserID(ctx, database.GetUserLinkRequestsByUserIDParams{
		UserID: user.ID,
		Status: status,
	})
	if err != nil && !xerrors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Fetch user link requests failed.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, convertUserLinkRequestsToGitAuthRequests(ulReqs, now))
}

func (api *API) gitAuthRequest(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := httpmw.UserParam(r)

	if !api.Authorize(r, rbac.ActionRead, rbac.ResourceUserData.WithOwner(user.ID.String())) {
		httpapi.ResourceNotFound(rw)
		return
	}

	requestID, ok := parseUUID(ctx, rw, r, "request_id")
	if !ok {
		return
	}

	now := time.Now()
	ulReq, err := api.Database.GetUserLinkRequestByIDAndUserID(ctx, database.GetUserLinkRequestByIDAndUserIDParams{
		ID:     requestID,
		UserID: user.ID,
	})
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			httpapi.ResourceNotFound(rw)
			return
		}
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Fetch user link request failed.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, convertUserLinkRequestToGitAuthRequest(ulReq, now))
}

func (api *API) gitAuthRequestConfirm(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := httpmw.UserParam(r)

	if !api.Authorize(r, rbac.ActionRead, rbac.ResourceUserData.WithOwner(user.ID.String())) {
		httpapi.ResourceNotFound(rw)
		return
	}

	requestID, ok := parseUUID(ctx, rw, r, "request_id")
	if !ok {
		return
	}
	provider := chi.URLParam(r, "provider")
	redirect := chi.URLParam(r, "redirect")
	if redirect == "" {
		redirect = "/"
	}

	ulReq, err := api.Database.GetUserLinkRequestByIDAndUserID(ctx, database.GetUserLinkRequestByIDAndUserIDParams{
		ID:     requestID,
		UserID: user.ID,
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Fetch user link request failed.",
			Detail:  err.Error(),
		})
		return
	}

	if time.Now().After(ulReq.ExpiresAt) {
		httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
			Message: "User link request expired.",
		})
		return
	}

	if ulReq.Resolved {
		httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
			Message: "Link request already resolved.",
		})
		return
	}

	// If automatic provider selection is not possible, it must be specified.
	if len(ulReq.Provider) > 1 && provider == "" {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "Provider required.",
		})
		return
	}

	if provider != "" {
		if !slices.Contains(ulReq.Provider, provider) {
			httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
				Message: "The given provider does not match the request.",
			})
			return
		}
	}

	var config GitProviderConfig
	if provider == "" {
		provider = ulReq.Provider[0]
	}
	for _, c := range api.GitProviderConfigs {
		if c.Name != provider {
			continue
		}
		config = c
		break
	}
	if config.Name == "" {
		httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
			Message: "Link request provider config for URL not found.",
		})
		return
	}

	var redirectURL string
	switch {
	case config.Github != nil:
		redirectURL = fmt.Sprintf("/api/v2/users/oauth2/github/callback?git_auth=%s&redirect=%s", config.Name, url.QueryEscape(redirect))
	case config.OIDC != nil:
		redirectURL = fmt.Sprintf("/api/v2/users/oidc/callback?git_auth=%s&redirect=%s", config.Name, url.QueryEscape(redirect))
	default:
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Unknown git authentication provider configuration.",
		})
	}
	http.Redirect(rw, r, redirectURL, http.StatusTemporaryRedirect)
}

func (api *API) workspaceAgentGitAuthRequest(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agent := httpmw.WorkspaceAgent(r)

	requestID, ok := parseUUID(ctx, rw, r, "request_id")
	if !ok {
		return
	}

	ulReq, err := api.Database.GetUserLinkRequestByIDAndAgentID(ctx, database.GetUserLinkRequestByIDAndAgentIDParams{
		ID:      requestID,
		AgentID: agent.ID,
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Fetch user link request failed.",
			Detail:  err.Error(),
		})
		return
	}

	authURL, ok := api.gitAuthConfirmURL(ctx, rw, ulReq.ID)
	if !ok {
		return
	}

	if !ulReq.Resolved {
		httpapi.Write(ctx, rw, http.StatusCreated, codersdk.WorkspaceAgentGitAuthResponse{
			RequestID:   ulReq.ID,
			User:        "",
			AccessToken: "",
			AuthURL:     authURL.String(),
		})
	}

	link, err := api.Database.GetUserLinkByGitAuthRequest(ctx, database.GetUserLinkByGitAuthRequestParams{
		UserID:    ulReq.UserID,
		LoginUser: ulReq.LoginUser,
		LoginUrl:  ulReq.LoginUrl,
	})
	if err != nil && !xerrors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching user link.",
			Detail:  err.Error(),
		})
		return
	}

	// TODO(mafredri): Ensure OAuth token is still valid, refresh if needed.
	httpapi.Write(ctx, rw, http.StatusCreated, codersdk.WorkspaceAgentGitAuthResponse{
		RequestID:   ulReq.ID,
		User:        link.LoginUser,
		AccessToken: link.OAuthAccessToken,
		AuthURL:     authURL.String(),
	})
}

func (api *API) postWorkspaceAgentGitAuthRequest(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agent := httpmw.WorkspaceAgent(r)

	var gaRequest codersdk.WorkspaceAgentGitAuthRequest
	if !httpapi.Read(ctx, rw, r, &gaRequest) {
		return
	}

	var config *GitProviderConfig
	for _, c := range api.GitProviderConfigs {
		if c.URL == gaRequest.URL {
			c := c
			config = &c
			break
		}
	}
	if config == nil {
		httpapi.Write(ctx, rw, http.StatusBadRequest, codersdk.Response{
			Message: "No config found for login URL.",
		})
		return
	}

	resource, err := api.Database.GetWorkspaceResourceByID(ctx, agent.ResourceID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace resource.",
			Detail:  err.Error(),
		})
		return
	}

	job, err := api.Database.GetWorkspaceBuildByJobID(ctx, resource.JobID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace build.",
			Detail:  err.Error(),
		})
		return
	}

	workspace, err := api.Database.GetWorkspaceByID(ctx, job.WorkspaceID)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching workspace.",
			Detail:  err.Error(),
		})
		return
	}

	// Check if a user link already exists, note that a new request will
	// be created in either scenario, but immediately resolved when
	// authentication is already done.
	link, err := api.Database.GetUserLinkByGitAuthRequest(ctx, database.GetUserLinkByGitAuthRequestParams{
		UserID:    workspace.OwnerID,
		LoginUser: gaRequest.User,
		LoginUrl:  gaRequest.URL,
	})
	if err != nil && !xerrors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error fetching user link.",
			Detail:  err.Error(),
		})
		return
	}

	createdAt := database.Now()
	expiresAt := createdAt.Add(1 * time.Hour)

	ulReq, err := api.Database.InsertUserLinkRequest(ctx, database.InsertUserLinkRequestParams{
		ID:        uuid.New(),
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		ExpiresAt: expiresAt,
		AgentID:   agent.ID,
		UserID:    workspace.OwnerID,
		Provider:  []string{config.Name}, // Only single match is supported for now.
		LoginUser: gaRequest.User,
		LoginUrl:  gaRequest.URL,
		Resolved:  link.OAuthAccessToken != "",
	})
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error creating user link request.",
			Detail:  err.Error(),
		})
		return
	}

	authURL, ok := api.gitAuthConfirmURL(ctx, rw, ulReq.ID)
	if !ok {
		return
	}

	httpapi.Write(ctx, rw, http.StatusCreated, codersdk.WorkspaceAgentGitAuthResponse{
		RequestID:   ulReq.ID,
		User:        link.LoginUser,
		AccessToken: link.OAuthAccessToken,
		AuthURL:     authURL.String(),
	})
}

func (api *API) gitAuthConfirmURL(ctx context.Context, rw http.ResponseWriter, reqID uuid.UUID) (*url.URL, bool) {
	// TODO(mafredri): Return the URL for the dashboard.
	authURL, err := api.AccessURL.Parse("/api/v2/users/me/git/auth/" + reqID.String() + "/confirm")
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error creating auth URL.",
			Detail:  err.Error(),
		})
		return nil, false
	}
	return authURL, true
}
