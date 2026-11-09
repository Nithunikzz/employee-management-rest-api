package repository

import (
	"cmp"
	"slices"
	"strings"
	"sync"

	"github.com/employee-management-rest-api/models"
)

type EmployeeRepository interface {
	Create(e models.Employee) models.Employee
	List(f models.Filter) []models.Employee
}

type InMemoryEmployeeRepository struct {
	mu        sync.RWMutex
	nextID    int64
	employees []models.Employee
}

func NewInMemoryEmployeeRepository() *InMemoryEmployeeRepository {
	return &InMemoryEmployeeRepository{nextID: 1}
}

func (r *InMemoryEmployeeRepository) Create(e models.Employee) models.Employee {
	r.mu.Lock()
	defer r.mu.Unlock()
	e.ID = r.nextID
	r.nextID++
	r.employees = append(r.employees, e)
	return e
}

func (r *InMemoryEmployeeRepository) List(f models.Filter) []models.Employee {
	r.mu.RLock()
	out := make([]models.Employee, 0, len(r.employees))
	name := strings.ToLower(f.Name)
	for _, e := range r.employees {
		if name != "" && !strings.Contains(strings.ToLower(e.Name), name) {
			continue
		}
		if f.Department != "" && !strings.EqualFold(e.Department, f.Department) {
			continue
		}
		if f.MinSalary != nil && e.Salary < *f.MinSalary {
			continue
		}
		if f.MaxSalary != nil && e.Salary > *f.MaxSalary {
			continue
		}
		out = append(out, e)
	}
	r.mu.RUnlock()

	slices.SortStableFunc(out, func(a, b models.Employee) int {
		var c int
		switch f.SortBy {
		case "name":
			c = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case "department":
			c = cmp.Compare(strings.ToLower(a.Department), strings.ToLower(b.Department))
		case "salary":
			c = cmp.Compare(a.Salary, b.Salary)
		}
		if c == 0 {
			c = cmp.Compare(a.ID, b.ID) 
		}
		if f.Desc {
			return -c
		}
		return c
	})
	return out
}
