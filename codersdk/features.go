package codersdk

import (
	"context"
	"encoding/json"
	"net/http"
)

type EnterpriseFeatures struct {
	Warnings       []string `json:"warnings"`
	UserLimit      bool     `json:"user_limit"`
	AuditLog       bool     `json:"audit_log"`
	BrowserOnly    bool     `json:"browser_only"`
	SCIM           bool     `json:"scim"`
	WorkspaceQuota bool     `json:"workspace_quota"`
	TemplateRBAC   bool     `json:"template_rbac"`
}

type EnterpriseFeature struct {
	Name        string      `json:"name"`
	Entitlement Entitlement `json:"entitlement"`
	// Default     bool        `json:"default"`
	Enabled bool `json:"enabled"`
}

type Entitlement string

const (
	EntitlementEntitled    Entitlement = "entitled"
	EntitlementGracePeriod Entitlement = "grace_period"
	EntitlementNotEntitled Entitlement = "not_entitled"
)

const (
	FeatureUserLimit      = "user_limit"
	FeatureAuditLog       = "audit_log"
	FeatureBrowserOnly    = "browser_only"
	FeatureSCIM           = "scim"
	FeatureWorkspaceQuota = "workspace_quota"
	FeatureTemplateRBAC   = "template_rbac"
)

var FeatureNames = []string{
	FeatureUserLimit,
	FeatureAuditLog,
	FeatureBrowserOnly,
	FeatureSCIM,
	FeatureWorkspaceQuota,
	FeatureTemplateRBAC,
}

type Feature struct {
	Entitlement Entitlement `json:"entitlement"`
	Enabled     bool        `json:"enabled"`
	Limit       *int64      `json:"limit,omitempty"`
	Actual      *int64      `json:"actual,omitempty"`
}

type Entitlements struct {
	Features     map[string]Feature `json:"features"`
	Warnings     []string           `json:"warnings"`
	HasLicense   bool               `json:"has_license"`
	Experimental bool               `json:"experimental"`
	Trial        bool               `json:"trial"`
}

func (c *Client) Entitlements(ctx context.Context) (Entitlements, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/entitlements", nil)
	if err != nil {
		return Entitlements{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Entitlements{}, readBodyAsError(res)
	}
	var ent Entitlements
	return ent, json.NewDecoder(res.Body).Decode(&ent)
}

func (c *Client) EnterpriseFeatures(ctx context.Context) (EnterpriseFeatures, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/features", nil)
	if err != nil {
		return EnterpriseFeatures{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return EnterpriseFeatures{}, readBodyAsError(res)
	}
	var ent EnterpriseFeatures
	return ent, json.NewDecoder(res.Body).Decode(&ent)
}
