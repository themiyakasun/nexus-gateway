package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
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