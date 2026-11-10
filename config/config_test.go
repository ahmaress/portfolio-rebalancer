package config

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"a:9092":             {"a:9092"},
		"a:9092,b:9092":      {"a:9092", "b:9092"},
		" a:9092 , ,b:9092 ": {"a:9092", "b:9092"},
		"":                   nil,
	}
	for in, want := range cases {
		if got := splitList(in); !reflect.DeepEqual(got, want) {
			t.Errorf("splitList(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLoad_RequiresJWTSecretUnlessDisabled(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("AUTH_DISABLED", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when JWT_SECRET is missing")
	}

	t.Setenv("AUTH_DISABLED", "true")
	if _, err := Load(); err != nil {
		t.Fatalf("expected no error with AUTH_DISABLED=true, got %v", err)
	}

	t.Setenv("AUTH_DISABLED", "")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if _, err := Load(); err != nil {
		t.Fatalf("expected no error with a valid secret, got %v", err)
	}
}
