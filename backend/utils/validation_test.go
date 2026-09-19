package utils

import "testing"

func TestValidatePoll(t *testing.T) {
	if err := ValidatePoll("Favourite language?", []string{"Go", "JavaScript"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePoll("", []string{"Go", "JavaScript"}); err == nil {
		t.Fatal("expected empty question to fail")
	}
	if err := ValidatePoll("Q", []string{"Go", "go"}); err == nil {
		t.Fatal("expected duplicate option to fail")
	}
}
func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("GoodPass1"); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePassword("password"); err == nil {
		t.Fatal("expected weak password to fail")
	}
}
