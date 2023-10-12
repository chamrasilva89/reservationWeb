package forms

// errors is a custom type representing validation errors, where each field can have multiple error messages.
type errors map[string][]string

// Add adds an error message to a specific field.
func (e errors) Add(field, message string) {
    // Append the error message to the list of error messages associated with the field.
    e[field] = append(e[field], message)
}

// Get retrieves the first error message associated with a field.
func (e errors) Get(field string) string {
    // Retrieve the list of error messages for the specified field.
    es := e[field]
    if len(es) == 0 {
        // If no error messages are found for the field, return an empty string.
        return ""
    }
    // Return the first error message for the field.
    return es[0]
}
