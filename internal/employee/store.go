package employee

import (
	"cmp"
	"slices"
	"strings"
	"sync"
)

// Employee is a single employee record.
type Employee struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Salary     float64 `json:"salary"`
}


type Filter struct {
	Name       string   
	Department string   
	MinSalary  *float64 
	MaxSalary  *float64 
	SortBy     string   
	Desc       bool
}

// Store is a concurrency-safe in-memory employee store.
type Store struct {
	mu        sync.RWMutex
	nextID    int64
	employees []Employee
}

func NewStore() *Store {
	return &Store{nextID: 1}
}

// Create assigns an ID to e and saves it.
func (s *Store) Create(e Employee) Employee {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = s.nextID
	s.nextID++
	s.employees = append(s.employees, e)
	return e
}

// List returns employees matching f, sorted as requested.
func (s *Store) List(f Filter) []Employee {
	s.mu.RLock()
	out := make([]Employee, 0, len(s.employees))
	name := strings.ToLower(f.Name)
	for _, e := range s.employees {
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
	s.mu.RUnlock()

	slices.SortStableFunc(out, func(a, b Employee) int {
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
			c = cmp.Compare(a.ID, b.ID) // deterministic tie-break
		}
		if f.Desc {
			return -c
		}
		return c
	})
	return out
}
