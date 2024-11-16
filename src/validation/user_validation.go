package validation

import "fmt"

type UserInput struct {
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Age     int    `json:"age"`
}

func ValidateUserInput(input UserInput) error {
	if input.Name == "" {
		return fmt.Errorf("name is required")
	}

	if input.Surname == "" {
		return fmt.Errorf("surname is required")
	}

	if input.Age <= 0 {
		return fmt.Errorf("aga must be greater than 0")
	}

	return nil
}
