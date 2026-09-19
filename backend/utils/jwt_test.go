package utils

import (
	"testing"
	"time"
)

func TestJWTGenerateAndParse(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, err := m.Generate("abc123")
	if err != nil {
		t.Fatal(err)
	}
	id, err := m.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != "abc123" {
		t.Fatalf("got %s", id)
	}
}
func TestJWTRejectsTampering(t *testing.T) {
	m := NewJWTManager("test-secret", time.Hour)
	token, _ := m.Generate("abc123")
	if _, err := m.Parse(token + "x"); err == nil {
		t.Fatal("expected tampered token to fail")
	}
}
