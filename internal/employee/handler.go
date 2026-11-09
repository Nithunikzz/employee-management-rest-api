package employee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Handler exposes the employee REST API.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mounts the routes on mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /employees", h.create)
	mux.HandleFunc("GET /employees", h.list)
}

type createRequest struct {
	Name       string   `json:"name"`
	Department string   `json:"department"`
	Salary     *float64 `json:"salary"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return
	}

	e, err := h.svc.Create(r.Context(), CreateInput{Name: req.Name, Department: req.Department, Salary: req.Salary})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("/employees/%d", e.ID))
	writeJSON(w, http.StatusCreated, e)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	emps, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(emps), "employees": emps})
}

func parseFilter(r *http.Request) (Filter, error) {
	q := r.URL.Query()
	f := Filter{
		Name:       strings.TrimSpace(q.Get("name")),
		Department: strings.TrimSpace(q.Get("department")),
	}

	var err error
	if f.MinSalary, err = parseFloatParam(q.Get("min_salary"), "min_salary"); err != nil {
		return f, err
	}
	if f.MaxSalary, err = parseFloatParam(q.Get("max_salary"), "max_salary"); err != nil {
		return f, err
	}

	switch sortBy := strings.ToLower(q.Get("sort_by")); sortBy {
	case "", "name", "department", "salary":
		f.SortBy = sortBy
	default:
		return f, fmt.Errorf("invalid sort_by %q: must be one of name, department, salary", sortBy)
	}

	switch order := strings.ToLower(q.Get("order")); order {
	case "", "asc":
	case "desc":
		f.Desc = true
	default:
		return f, fmt.Errorf("invalid order %q: must be asc or desc", order)
	}
	return f, nil
}

func parseFloatParam(v, name string) (*float64, error) {
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", name)
	}
	return &n, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeServiceError(w http.ResponseWriter, err error) {
	if IsValidation(err) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal server error")
}
