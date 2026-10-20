package admin

import (
	"time"
)

type Service struct {
	ID string `json:"id"`
	Name string `json:"name"`
	URL string `json:"url"`
	HealthCheckPath string `json:"health_check_path"`
	IsActive bool `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateServiceRequest struct {
	Name string `json:"name"`
	URL string `json:"url"`
	HealthCheckPath string `json:"health_check_path,omitempty"`
}

type Route struct {
	ID string `json:"id"`
	ServiceID string `json:"service_id"`
	PathPrefix string `json:"path_prefix"`
	StripPrefix bool `json:"strip_prefix"`
	AuthRequired bool `json:"auth_required"`
	RateLimitOverrideRPM *int `json:"rate_limit_override_rpm,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRouteRequest struct {
	ServiceID string `json:"service_id"`
	PathPrefix string `json:"path_prefix"`
	StripPrefix bool `json:"strip_prefix"`
	AuthRequired bool `json:"auth_required"`
	RateLimitOverrideRPM *int `json:"rate_limit_override_rpm,omitempty"`
}

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateOrgRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type APIKey struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	KeyPrefix      string     `json:"key_prefix"`
	Name           string     `json:"name"`
	IsActive       bool       `json:"is_active"`
	RateLimitRPM   int        `json:"rate_limit_rpm"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}
type CreateAPIKeyRequest struct {
	OrganizationID string `json:"organization_id"`
	Name           string `json:"name"`
	RateLimitRPM   int    `json:"rate_limit_rpm"`
}
type CreateAPIKeyResponse struct {
	APIKey APIKey `json:"api_key"`
	RawKey string `json:"raw_key"` 
}