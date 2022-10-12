package codersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type WorkspaceAgentGitAuthRequest struct {
	User string `json:"user"`
	URL  string `json:"url"`
}

type WorkspaceAgentGitAuthResponse struct {
	RequestID   uuid.UUID `json:"request_id"`
	AuthURL     string    `json:"auth_url"`
	User        string    `json:"user,omitempty"`
	AccessToken string    `json:"access_token,omitempty"`
}

func (c *Client) WorkspaceAgentRequestGitAuth(ctx context.Context, req WorkspaceAgentGitAuthRequest) (WorkspaceAgentGitAuthResponse, error) {
	resp, err := c.Request(ctx, http.MethodPost, "/api/v2/workspaceagents/me/git/auth", req)
	if err != nil {
		return WorkspaceAgentGitAuthResponse{}, err
	}
	defer resp.Body.Close()

	var reply WorkspaceAgentGitAuthResponse
	return reply, json.NewDecoder(resp.Body).Decode(&reply)
}

func (c *Client) WorkspaceAgentGitAuthRequest(ctx context.Context, requestID uuid.UUID) (WorkspaceAgentGitAuthResponse, error) {
	resp, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/workspaceagents/me/git/auth/%s", requestID.String()), nil)
	if err != nil {
		return WorkspaceAgentGitAuthResponse{}, err
	}
	defer resp.Body.Close()

	var reply WorkspaceAgentGitAuthResponse
	return reply, json.NewDecoder(resp.Body).Decode(&reply)
}

type GitAuthRequestStatus string

// GitAuthRequestStatus enums.
const (
	GitAuthRequestStatusAll      GitAuthRequestStatus = ""
	GitAuthRequestStatusResolved GitAuthRequestStatus = "resolved"
	GitAuthRequestStatusPending  GitAuthRequestStatus = "pending"
	GitAuthRequestStatusExpired  GitAuthRequestStatus = "expired"
)

type GitAuthRequestsRequest struct {
	Status GitAuthRequestStatus `json:"status"`
}

type GitAuthRequestResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	AgentID   uuid.UUID `json:"agent_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Provider  []string  `json:"provider"`
	LoginUser string    `json:"login_user"`
	LoginURL  string    `json:"login_url"`
	Resolved  bool      `json:"resolved"`
	Expired   bool      `json:"expired"`
}

func (c *Client) GitAuthRequests(ctx context.Context, req GitAuthRequestsRequest) ([]GitAuthRequestResponse, error) {
	resp, err := c.Request(ctx, http.MethodGet, "/api/v2/users/me/git/auth", nil,
		func(r *http.Request) {
			q := r.URL.Query()
			q.Set("status", string(req.Status))
			r.URL.RawQuery = q.Encode()
		},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var reply []GitAuthRequestResponse
	return reply, json.NewDecoder(resp.Body).Decode(&reply)
}

func (c *Client) GitAuthRequest(ctx context.Context, requestID uuid.UUID) (GitAuthRequestResponse, error) {
	resp, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/users/me/git/auth/%s", requestID.String()), nil)
	if err != nil {
		return GitAuthRequestResponse{}, err
	}
	defer resp.Body.Close()

	var reply GitAuthRequestResponse
	return reply, json.NewDecoder(resp.Body).Decode(&reply)
}
