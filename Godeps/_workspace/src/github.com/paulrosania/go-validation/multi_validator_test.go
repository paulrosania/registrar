package validation_test

import (
	"testing"

	"github.com/paulrosania/go-validation"
)

func TestMultiValidator(t *testing.T) {
	v := validation.NewMultiValidator()
	v.Assert(true, "field-1", "should be true")
	v.Present("foo", "field-2")

	if !v.Valid() {
		t.Error("expected validator to be valid")
	}

	v.Assert(false, "field-3", "should be true")
	v.Present("", "field-4")

	if v.Valid() {
		t.Error("expected validator to be invalid")
	}

	errs := v.Errors()
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}

	if len(errs["field-3"]) != 1 || errs["field-3"][0] != "should be true" {
		t.Errorf("incorrect error for field-3:\n\texpected: %s\n\tgot: %v", "should be true", errs["field-3"])
	}

	if len(errs["field-4"]) != 1 || errs["field-4"][0] != "field-4 must be present" {
		t.Errorf("incorrect error for field-4:\n\texpected: %s\n\tgot: %v", "field-4 must be present", errs["field-4"])
	}
}
