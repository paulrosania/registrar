package validation

import (
	"fmt"
	"strings"
)

type Validator struct {
	errs []string
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) Assert(valid bool, msg string) {
	if !valid {
		v.errs = append(v.errs, msg)
	}
}

func (v *Validator) Present(s string, field string) {
	if s == "" {
		v.errs = append(v.errs, fmt.Sprintf("%s must be present", field))
	}
}

func (v *Validator) Valid() bool {
	return len(v.errs) == 0
}

func (v *Validator) Errors() []string {
	return v.errs
}

func (v *Validator) String() string {
	switch len(v.errs) {
	case 0:
		return ""
	case 1:
		return v.errs[0]
	default:
		return fmt.Sprintf("multiple validation errors: %s", strings.Join(v.errs, ", "))
	}
}

func (v *Validator) Err() error {
	if v.String() == "" {
		return nil
	} else {
		return fmt.Errorf(v.String())
	}
}
