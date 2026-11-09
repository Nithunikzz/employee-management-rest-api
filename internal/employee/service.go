package employee

import (
	"context"
	"errors"
	"strings"
)

type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Problems, "; ")
}

// Repository is the persistence the service depends on. *Store implements it.
type Repository interface {
	Create(e Employee) Employee
	List(f Filter) []Employee
}

// CreateInput is the data needed to create an employee.
type CreateInput struct {
	Name       string
	Department string
	Salary     *float64
}

// Service holds the employee business rules, independent of HTTP.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create validates in and stores a new employee.
func (s *Service) Create(_ context.Context, in CreateInput) (Employee, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Department = strings.TrimSpace(in.Department)

	var problems []string
	if in.Name == "" {
		problems = append(problems, "name is required")
	}
	if in.Department == "" {
		problems = append(problems, "department is required")
	}
	if in.Salary == nil {
		problems = append(problems, "salary is required")
	} else if *in.Salary < 0 {
		problems = append(problems, "salary must be >= 0")
	}
	if len(problems) > 0 {
		return Employee{}, &ValidationError{Problems: problems}
	}

	return s.repo.Create(Employee{Name: in.Name, Department: in.Department, Salary: *in.Salary}), nil
}

func (s *Service) List(_ context.Context, f Filter) ([]Employee, error) {
	if f.MinSalary != nil && f.MaxSalary != nil && *f.MinSalary > *f.MaxSalary {
		return nil, &ValidationError{Problems: []string{"min_salary must be <= max_salary"}}
	}
	return s.repo.List(f), nil
}

func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
