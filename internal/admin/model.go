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