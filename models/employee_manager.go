package models

type EmployeeManagerCreateOption struct {
	Managers []EmployeeManager `json:"managers"`
}

type EmployeeManager struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type EmployeeManagerDeleteOption struct {
	Managers []EmployeeManagerId `json:"managers"`
}

type EmployeeManagerId struct {
	ID string `json:"id"`
}
