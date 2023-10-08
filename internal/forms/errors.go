package forms

type errors map[string][]string

// add error to given field.
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// Get return 1st error messsgaee
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
