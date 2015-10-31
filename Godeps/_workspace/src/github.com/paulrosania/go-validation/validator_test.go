package validation_test

import (
	"testing"

	"github.com/paulrosania/go-validation"
)

func TestAssert(t *testing.T) {
	v1 := validation.NewValidator()
	v1.Assert(false, "should be true")

	if v1.Valid() {
		t.Error("expected validator to be invalid")
	}

	if v1.String() != "should be true" {
		t.Errorf("incorrect error string:\n\texpected: %s\n\tgot: %s", "should be true", v1.String())
	}

	v2 := validation.NewValidator()
	v2.Assert(true, "should be true")

	if !v2.Valid() {
		t.Error("expected validator to be valid")
	}

	if v2.String() != "" {
		t.Errorf("incorrect error string:\n\texpected: %s\n\tgot: %s", "", v2.String())
	}
}

func TestPresent(t *testing.T) {
	v1 := validation.NewValidator()
	v1.Present("", "foo")

	if v1.Valid() {
		t.Error("expected validator to be invalid")
	}

	if v1.String() != "foo must be present" {
		t.Errorf("incorrect error string:\n\texpected: %s\n\tgot: %s", "foo must be present", v1.String())
	}

	v2 := validation.NewValidator()
	v2.Present("bar", "foo")

	if !v2.Valid() {
		t.Error("expected validator to be valid")
	}

	if v2.String() != "" {
		t.Errorf("incorrect error string:\n\texpected: %s\n\tgot: %s", "", v2.String())
	}
}
