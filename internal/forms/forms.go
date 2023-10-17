package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// Form represents a web form and its data, including any validation errors.
type Form struct {
	url.Values
	Errors errors
}

// New creates a new Form from the provided data and initializes an errors map.
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// Has checks if a field exists in the form data.
func (f *Form) Has(field string) bool {
	x := f.Get(field)

	if x == "" {
		return false
	}
	return true
}

// Valid checks if the form has any validation errors.
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Required checks if specified fields are not empty and adds errors if they are.
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank "+field)
		}
	}
}

// MinLength checks if a field has a minimum length and adds an error if not.
func (f *Form) MinLength(field string, length int) bool {
	v := f.Get(field)
	if len(v) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// IsEmail checks if a field's value is a valid email address.
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email format")
	}
}
