package config

import (
	"strings"
	"testing"
)

func TestValidateKey_KnownDefaults(t *testing.T) {
	for _, key := range KnownDefaults {
		if err := ValidateKey(key); err != nil {
			t.Errorf("ValidateKey(%q) returned unexpected error: %v", key, err)
		}
	}
}

func TestValidateKey_PluginConfigPrefix(t *testing.T) {
	cases := []string{
		PluginConfigsPrefix + "archgo.namespace",
		PluginConfigsPrefix + "fscheck.root",
	}
	for _, key := range cases {
		if err := ValidateKey(key); err != nil {
			t.Errorf("ValidateKey(%q) returned unexpected error: %v", key, err)
		}
	}
}

func TestValidateKey_UnknownKey(t *testing.T) {
	err := ValidateKey("not.a.real.key")
	if err == nil {
		t.Fatal("expected error for unknown key, got nil")
	}
	if !strings.Contains(err.Error(), "unknown config key") {
		t.Errorf("error should mention 'unknown config key', got: %v", err)
	}
	// The error should list the allowed keys as a hint.
	if !strings.Contains(err.Error(), DefaultInput) {
		t.Errorf("error should list known keys, got: %v", err)
	}
}
