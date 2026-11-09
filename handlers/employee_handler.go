package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/employee-management-rest-api/models"
	"github.com/employee-management-rest-api/services"
)

type EmployeeHandler struct {
	svc *services.EmployeeService
}

func NewEmployeeHandler(svc *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/employees", h.Create)
	r.GET("/employees", h.List)
}

type createRequest struct {
	Name       string   `json:"name"`
	Department string   `json:"department"`
	Salary     *float64 `json:"salary"`
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body: " + err.Error()})
		return
	}

	e, err := h.svc.Create(c.Request.Context(), services.CreateInput{
		Name:       req.Name,
		Department: req.Department,
		Salary:     req.Salary,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.Header("Location", fmt.Sprintf("/employees/%d", e.ID))
	c.JSON(http.StatusCreated, e)
}

func (h *EmployeeHandler) List(c *gin.Context) {
	f, err := parseFilter(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	emps, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": len(emps), "employees": emps})
}

func parseFilter(c *gin.Context) (models.Filter, error) {
	f := models.Filter{
		Name:       strings.TrimSpace(c.Query("name")),
		Department: strings.TrimSpace(c.Query("department")),
	}

	var err error
	if f.MinSalary, err = parseFloatParam(c.Query("min_salary"), "min_salary"); err != nil {
		return f, err
	}
	if f.MaxSalary, err = parseFloatParam(c.Query("max_salary"), "max_salary"); err != nil {
		return f, err
	}

	switch sortBy := strings.ToLower(c.Query("sort_by")); sortBy {
	case "", "name", "department", "salary":
		f.SortBy = sortBy
	default:
		return f, fmt.Errorf("invalid sort_by %q: must be one of name, department, salary", sortBy)
	}

	switch order := strings.ToLower(c.Query("order")); order {
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

func writeServiceError(c *gin.Context, err error) {
	if services.IsValidation(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
