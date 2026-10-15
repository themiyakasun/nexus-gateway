package admin

import (
	"context"
	"database/sql"
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
						RETURNING id, name, url, health_check_path, is active, created_at, updated_at
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