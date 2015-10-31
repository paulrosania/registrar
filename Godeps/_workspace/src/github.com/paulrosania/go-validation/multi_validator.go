package validation

type MultiValidator struct {
	fields map[string]*Validator
}

func NewMultiValidator() *MultiValidator {
	return &MultiValidator{
		fields: make(map[string]*Validator),
	}
}

func (v *MultiValidator) ensureField(name string) {
	if _, ok := v.fields[name]; !ok {
		v.fields[name] = &Validator{}
	}
}

func (v *MultiValidator) Assert(valid bool, field, msg string) {
	v.ensureField(field)
	v.fields[field].Assert(valid, msg)
}

func (v *MultiValidator) Present(s string, field string) {
	v.ensureField(field)
	v.fields[field].Present(s, field)
}

func (v *MultiValidator) Valid() bool {
	for _, v := range v.fields {
		if !v.Valid() {
			return false
		}
	}
	return true
}

func (v *MultiValidator) Errors() map[string][]string {
	errs := make(map[string][]string)
	for k, v := range v.fields {
		if !v.Valid() {
			errs[k] = v.Errors()
		}
	}
	return errs
}
