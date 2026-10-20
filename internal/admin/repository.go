package admin

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateService(ctx context.Context, req CreateServiceRequest) (*Service, error) {
	if req.HealthCheckPath == "" {
		req.HealthCheckPath = "/health"
	}

	query := `
						INSERT INTO services (name, url, health_check_path)
						VALUES ($1, $2, $3)
						RETURNING id, name, url, health_check_path, is_active, created_at, updated_at
	`

	var s Service
	err := r.db.QueryRowContext(ctx, query, req.Name, req.URL, req.HealthCheckPath).Scan(&s.ID, &s.Name, &s.URL, &s.HealthCheckPath, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("could not create service: %w", err)
	}

	return &s, nil
}

func (r *Repository) ListServices(ctx context.Context) ([]Service, error) {
	query := `SELECT id, name, url, health_check_path, is_active, created_at, updated_at FROM services ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.HealthCheckPath, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, rows.Err()
}

func (r *Repository) GetServiceByID(ctx context.Context, id string) (*Service, error) {
	query := `SELECT id, name, url, health_check_path, is_active, created_at, updated_at FROM services WHERE id = $1`

	var s Service
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.URL, &s.HealthCheckPath, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) DeleteService(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected,_ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Routes
func (r *Repository) CreateRoute(ctx context.Context, req CreateRouteRequest) (*Route, error) {
	query := `
	INSERT INTO routes (service_id, path_prefix, strip_prefix, auth_required, rate_limit_override_rpm) 
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, service_id, path_prefix, strip_prefix, auth_required, rate_limit_override_rpm, created_at, updated_at
	`

	var rt Route
	err := r.db.QueryRowContext(ctx, query, req.ServiceID, req.PathPrefix, req.StripPrefix, req.AuthRequired, req.RateLimitOverrideRPM).Scan(&rt.ID, &rt.ServiceID, &rt.PathPrefix, &rt.StripPrefix, &rt.AuthRequired, &rt.RateLimitOverrideRPM, &rt.CreatedAt, &rt.UpdatedAt,)
	if err != nil {
		return nil, fmt.Errorf("could not create route: %w", err)
	}

	return &rt, nil
}

func (r *Repository) ListRoutes(ctx context.Context) ([]Route, error) {
	query := `SELECT id, service_id, path_prefix, strip_prefix, auth_required, rate_limit_override_rpm, created_at, updated_at FROM routes ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []Route
	for rows.Next() {
		var rt Route 
		if err := rows.Scan(&rt.ID, &rt.ServiceID, &rt.PathPrefix, &rt.StripPrefix, &rt.AuthRequired, &rt.RateLimitOverrideRPM, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		routes = append(routes, rt)
	}

	return routes, rows.Err()
}

func (r *Repository) DeleteRoute(ctx context.Context, id string) error {
	query := `DELETE FROM routes WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}



func (r *Repository) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
	query := `
		INSERT INTO organizations (name, slug)
		VALUES ($1, $2)
		RETURNING id, name, slug, is_active, created_at, updated_at
	`
	var o Organization
	err := r.db.QueryRowContext(ctx, query, req.Name, req.Slug).Scan(
		&o.ID, &o.Name, &o.Slug, &o.IsActive, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create organization: %w", err)
	}
	return &o, nil
}
func (r *Repository) ListOrganizations(ctx context.Context) ([]Organization, error) {
	query := `SELECT id, name, slug, is_active, created_at, updated_at FROM organizations ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orgs []Organization
	for rows.Next() {
		var o Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.IsActive, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, o)
	}
	return orgs, rows.Err()
}

func (r *Repository) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	if req.RateLimitRPM <= 0 {
		req.RateLimitRPM = 60
	}

	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	rawKey := fmt.Sprintf("sk_live_%s", hex.EncodeToString(randomBytes))
	prefix := rawKey[:12] + "..."

	hashBytes := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hashBytes[:])

	query := `
		INSERT INTO api_keys (organization_id, key_hash, key_prefix, name, rate_limit_rpm)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, organization_id, key_prefix, name, is_active, rate_limit_rpm, created_at, expires_at
	`
	var k APIKey
	err := r.db.QueryRowContext(ctx, query, req.OrganizationID, keyHash, prefix, req.Name, req.RateLimitRPM).Scan(
		&k.ID, &k.OrganizationID, &k.KeyPrefix, &k.Name, &k.IsActive, &k.RateLimitRPM, &k.CreatedAt, &k.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create api key: %w", err)
	}
	return &CreateAPIKeyResponse{
		APIKey: k,
		RawKey: rawKey,
	}, nil
}
func (r *Repository) ListAPIKeys(ctx context.Context, orgID string) ([]APIKey, error) {
	query := `SELECT id, organization_id, key_prefix, name, is_active, rate_limit_rpm, created_at, expires_at FROM api_keys WHERE organization_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.OrganizationID, &k.KeyPrefix, &k.Name, &k.IsActive, &k.RateLimitRPM, &k.CreatedAt, &k.ExpiresAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
func (r *Repository) RevokeAPIKey(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET is_active = false WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}