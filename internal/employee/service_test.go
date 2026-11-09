package employee

import (
	"context"
	"testing"
)

func TestServiceCreateValidation(t *testing.T) {
	svc := NewService(NewStore())
	_, err := svc.Create(context.Background(), CreateInput{Name: " ", Department: "Eng"})
	if !IsValidation(err) {
		t.Fatalf("want validation error, got %v", err)
	}
	if got := err.Error(); got != "name is required; salary is required" {
		t.Fatalf("message = %q", got)
	}

	salary := 10.0
	e, err := svc.Create(context.Background(), CreateInput{Name: " Ann ", Department: "Eng", Salary: &salary})
	if err != nil || e.ID != 1 || e.Name != "Ann" {
		t.Fatalf("got %+v, %v", e, err)
	}
}

func TestServiceListSalaryRange(t *testing.T) {
	lo, hi := 10.0, 5.0
	if _, err := NewService(NewStore()).List(context.Background(), Filter{MinSalary: &lo, MaxSalary: &hi}); !IsValidation(err) {
		t.Fatalf("want validation error, got %v", err)
	}
}
