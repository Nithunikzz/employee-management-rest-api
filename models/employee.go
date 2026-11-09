package models

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
