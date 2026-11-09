package services

import (
	"context"
	"errors"
	"strings"

	"github.com/employee-management-rest-api/models"
	"github.com/employee-management-rest-api/repository"
)

type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Problems, "; ")
}

type CreateInput struct {
	Name       string
	Department string
	Salary     *float64
}

type EmployeeService struct {
	repo repository.EmployeeRepository
}

func NewEmployeeService(repo repository.EmployeeRepository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) Create(_ context.Context, in CreateInput) (models.Employee, error) {
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
		return models.Employee{}, &ValidationError{Problems: problems}
	}

	return s.repo.Create(models.Employee{Name: in.Name, Department: in.Department, Salary: *in.Salary}), nil
}

func (s *EmployeeService) List(_ context.Context, f models.Filter) ([]models.Employee, error) {
	if f.MinSalary != nil && f.MaxSalary != nil && *f.MinSalary > *f.MaxSalary {
		return nil, &ValidationError{Problems: []string{"min_salary must be <= max_salary"}}
	}
	return s.repo.List(f), nil
}

func IsValidation(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
