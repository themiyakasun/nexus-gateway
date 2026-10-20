package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/themiyakasun/nexus-gateway/internal/events"
)

type Handler struct {
	repo *Repository
	broadcaster *events.Broadcaster
}

func NewHandler(repo *Repository, broardcaster *events.Broadcaster) *Handler {
	return &Handler{repo: repo, broadcaster: broardcaster}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Services CRUD
	mux.HandleFunc("POST /admin/services", h.handleCreateService)
	mux.HandleFunc("GET /admin/services", h.handleListServices)
	mux.HandleFunc("GET /admin/services/{id}", h.handleGetService)
	mux.HandleFunc("DELETE /admin/services/{id}", h.handleDeleteService)

	// Routes CRUD
	mux.HandleFunc("POST /admin/routes", h.handleCreateRoute)
	mux.HandleFunc("GET /admin/routes", h.handleListRoutes)
	mux.HandleFunc("DELETE /admin/routes/{id}", h.handleDeleteRoute)

 // Organizations CRUD
	mux.HandleFunc("POST /admin/organizations", h.handleCreateOrg)
	mux.HandleFunc("GET /admin/organizations", h.handleListOrgs)

	// API Keys CRUD
	mux.HandleFunc("POST /admin/api-keys", h.handleCreateAPIKey)
	mux.HandleFunc("GET /admin/api-keys", h.handleListAPIKeys)
	mux.HandleFunc("DELETE /admin/api-keys/{id}", h.handleRevokeAPIKey)

}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Service handlers
func (h *Handler) handleCreateService(w http.ResponseWriter, r *http.Request) {
	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.URL == "" {
		respondError(w, http.StatusBadRequest, "name and url are required")
		return
	}

	service, err := h.repo.CreateService(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.broadcaster != nil {
		_ = h.broadcaster.Publish(r.Context(), events.EventAPIKeyCreated, service)
	}

	respondJSON(w, http.StatusCreated, service)
}

func (h *Handler) handleListServices(w http.ResponseWriter, r *http.Request) {
	services, err := h.repo.ListServices(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if services == nil {
		services = []Service{}
	}
	respondJSON(w, http.StatusOK, services)
}

func (h *Handler) handleGetService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	service, err := h.repo.GetServiceByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if service == nil {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}
	respondJSON(w, http.StatusOK,service)
}

func (h *Handler) handleDeleteService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.repo.DeleteService(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "service not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "service deleted successfully"})
}


//Routes handlers
func (h *Handler) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ServiceID == "" || req.PathPrefix == "" {
		respondError(w, http.StatusBadRequest, "service_id and path_prefix are required")
		return
	}
	route, err := h.repo.CreateRoute(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.broadcaster != nil {
		_ = h.broadcaster.Publish(r.Context(), events.EventRouteCreated, route)
	}

	respondJSON(w, http.StatusCreated, route)
}
func (h *Handler) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.repo.ListRoutes(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if routes == nil {
		routes = []Route{}
	}
	respondJSON(w, http.StatusOK, routes)
}
func (h *Handler) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.repo.DeleteRoute(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "route not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "route deleted successfully"})
}


func (h *Handler) handleCreateOrg(w http.ResponseWriter, r *http.Request) {
	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Slug == "" {
		respondError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	org, err := h.repo.CreateOrganization(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, org)
}
func (h *Handler) handleListOrgs(w http.ResponseWriter, r *http.Request) {
	orgs, err := h.repo.ListOrganizations(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if orgs == nil {
		orgs = []Organization{}
	}
	respondJSON(w, http.StatusOK, orgs)
}

func (h *Handler) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.OrganizationID == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "organization_id and name are required")
		return
	}
	res, err := h.repo.CreateAPIKey(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.broadcaster != nil {
		_ = h.broadcaster.Publish(r.Context(), events.EventAPIKeyCreated, err)
	}
	
	respondJSON(w, http.StatusCreated, res)
}
func (h *Handler) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		respondError(w, http.StatusBadRequest, "organization_id query param is required")
		return
	}
	keys, err := h.repo.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if keys == nil {
		keys = []APIKey{}
	}
	respondJSON(w, http.StatusOK, keys)
}
func (h *Handler) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.repo.RevokeAPIKey(r.Context(), id)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "api key not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"message": "api key revoked successfully"})
}