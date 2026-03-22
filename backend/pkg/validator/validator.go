package validator

import "github.com/go-playground/validator/v10"

// Validate is the shared validator instance.
var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidateStruct validates a struct and returns a slice of human-readable error messages.
func ValidateStruct(s any) []string {
	err := Validate.Struct(s)
	if err == nil {
		return nil
	}

	var messages []string
	for _, e := range err.(validator.ValidationErrors) {
		messages = append(messages, e.Field()+" failed on '"+e.Tag()+"' validation")
	}
	return messages
}
