package dsl

import (
	"strings"
	"testing"

	"github.com/phi42/ad-enforcement-tool/rule"
)

func TestParseDSL_CodeRuleMultipleAssertions(t *testing.T) {
	src := `
adr "0001" "Test"

component "A" = "com.example.a"
component "B" = "com.example.b"
component "C" = "com.example.c"

code "two_assertions" {
  A must not depend on B
  A must not depend on C
  severity error
}
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected an error for code rule with multiple assertions, got nil")
	}
	if !strings.Contains(err.Error(), "more than one assertion") {
		t.Fatalf("expected error to mention 'more than one assertion', got: %v", err)
	}
}

func TestParseDSL_FileRuleMultipleChecks(t *testing.T) {
	src := `
adr "0001" "Test"

path "Lic" = "LICENSE"
path "Readme" = "README.md"

file "two_checks" {
  Lic must exist
  Readme must exist
  severity warning
}
`
	_, err := ParseDSL(src)
	if err != nil {
		t.Fatalf("file rules should allow multiple assertions, got error: %v", err)
	}
}

func TestParseDSL_CodeRuleSingleAssertion(t *testing.T) {
	src := `
adr "0001" "Test"

component "A" = "com.example.a"
component "B" = "com.example.b"

code "one_assertion" {
  A must not depend on B
  severity error
}
`
	_, err := ParseDSL(src)
	if err != nil {
		t.Fatalf("code rule with a single assertion should be valid, got error: %v", err)
	}
}

// TestParseDSL_SpecFields verifies the parsed IR has the expected ADR, selector, and rule fields.
func TestParseDSL_SpecFields(t *testing.T) {
	src := `
adr "0042" "Layering"

component "Domain" = "com.example.domain"
component "Application" = "com.example.app"

code "no_upward_deps" {
  Application must not depend on Domain
  severity warning
}
`
	spec, err := ParseDSL(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if spec.Adr == nil {
		t.Fatal("spec.Adr is nil")
	}
	if spec.Adr.Id != "0042" {
		t.Errorf("ADR id: got %q, want %q", spec.Adr.Id, "0042")
	}
	if spec.Adr.Title != "Layering" {
		t.Errorf("ADR title: got %q, want %q", spec.Adr.Title, "Layering")
	}

	if len(spec.Selectors) != 2 {
		t.Fatalf("selector count: got %d, want 2", len(spec.Selectors))
	}
	if spec.Selectors[0].Name != "Domain" || spec.Selectors[0].Pattern != "com.example.domain" {
		t.Errorf("selector[0]: got %+v", spec.Selectors[0])
	}
	if spec.Selectors[0].Kind != rule.SelectorKind_SELECTOR_COMPONENT {
		t.Errorf("selector[0] kind: got %v, want SELECTOR_COMPONENT", spec.Selectors[0].Kind)
	}

	if len(spec.Rules) != 1 {
		t.Fatalf("rule count: got %d, want 1", len(spec.Rules))
	}
	r := spec.Rules[0]
	if r.Name != "no_upward_deps" {
		t.Errorf("rule name: got %q, want %q", r.Name, "no_upward_deps")
	}
	if r.Kind != rule.RuleKind_RULE_NOT_DEPEND {
		t.Errorf("rule kind: got %v, want RULE_NOT_DEPEND", r.Kind)
	}
	if r.From == nil || r.From.Value != "Application" {
		t.Errorf("rule From: got %v, want Application", r.From)
	}
	if len(r.Targets) != 1 || r.Targets[0].Value != "Domain" {
		t.Errorf("rule Targets: got %v, want [Domain]", r.Targets)
	}
	if r.Severity != rule.Severity_SEVERITY_WARNING {
		t.Errorf("rule severity: got %v, want SEVERITY_WARNING", r.Severity)
	}
}

// TestParseDSL_MissingADR checks that omitting the adr declaration is rejected.
func TestParseDSL_MissingADR(t *testing.T) {
	src := `
component "A" = "com.example.a"

code "some_rule" {
  A must not depend on A
  severity error
}
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected error for missing ADR declaration, got nil")
	}
	// The grammar requires adr to appear first, so the parser emits a syntax error.
	// Just verify that an error is returned; the exact message is an ANTLR diagnostic.
}

// TestValidateIR_MissingADR exercises the semantic validator directly.
func TestValidateIR_MissingADR(t *testing.T) {
	spec := &rule.Spec{} // Adr == nil
	err := validateIR(spec)
	if err == nil {
		t.Fatal("expected error for nil Adr, got nil")
	}
	if !strings.Contains(err.Error(), "missing ADR declaration") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseDSL_DuplicateSelector checks that two selectors with the same name are rejected.
func TestParseDSL_DuplicateSelector(t *testing.T) {
	src := `
adr "0001" "Test"

component "Domain" = "com.example.a"
component "Domain" = "com.example.b"
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected error for duplicate selector, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate selector") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseDSL_DuplicateRuleName checks that two rules sharing a name are rejected.
func TestParseDSL_DuplicateRuleName(t *testing.T) {
	src := `
adr "0001" "Test"

component "A" = "com.example.a"
component "B" = "com.example.b"

code "same_name" {
  A must not depend on B
}

code "same_name" {
  B must not depend on A
}
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected error for duplicate rule name, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate rule name") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseDSL_UnknownSelectorRef checks that referencing an undeclared selector is rejected.
func TestParseDSL_UnknownSelectorRef(t *testing.T) {
	src := `
adr "0001" "Test"

component "A" = "com.example.a"

code "bad_ref" {
  A must not depend on Undefined
  severity error
}
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected error for unknown selector reference, got nil")
	}
	if !strings.Contains(err.Error(), "unknown selector") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestParseDSL_CustomBlock checks that custom blocks are extracted into rules with IsCustomRule set.
func TestParseDSL_CustomBlock(t *testing.T) {
	src := `
adr "0001" "Test"

custom "my_plugin_rule" {
  arbitrary plugin-defined body
  with multiple lines
}
`
	spec, err := ParseDSL(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Rules) != 1 {
		t.Fatalf("rule count: got %d, want 1", len(spec.Rules))
	}
	r := spec.Rules[0]
	if r.Name != "my_plugin_rule" {
		t.Errorf("custom rule name: got %q, want %q", r.Name, "my_plugin_rule")
	}
	if !r.IsCustomRule {
		t.Error("expected IsCustomRule=true")
	}
	if !strings.Contains(r.RawBody, "arbitrary plugin-defined body") {
		t.Errorf("RawBody missing expected content: %q", r.RawBody)
	}
}

// TestParseDSL_UnterminatedCustomBlock checks that a custom block without a closing brace is rejected.
func TestParseDSL_UnterminatedCustomBlock(t *testing.T) {
	src := `
adr "0001" "Test"

custom "broken" {
  no closing brace
`
	_, err := ParseDSL(src)
	if err == nil {
		t.Fatal("expected error for unterminated custom block, got nil")
	}
	if !strings.Contains(err.Error(), "unterminated custom block") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestUnquote checks the DSL string unquoting helper.
func TestUnquote(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`"hello"`, "hello"},
		{`"with \"escaped\" quotes"`, `with "escaped" quotes`},
		{`"no-quotes"`, "no-quotes"},
		{"notquoted", "notquoted"},
		{`""`, ""},
	}
	for _, tc := range cases {
		got := unquote(tc.in)
		if got != tc.want {
			t.Errorf("unquote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFindMatchingBrace checks that the brace-matching helper handles nesting and missing braces.
func TestFindMatchingBrace(t *testing.T) {
	cases := []struct {
		name    string
		src     string
		openPos int
		want    int
	}{
		{"simple", "{}", 0, 1},
		{"nested", "{ {} }", 0, 5},
		{"offset", "xx{yy{}zz}aa", 2, 9},
		{"unmatched", "{no close", 0, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := findMatchingBrace(tc.src, tc.openPos)
			if got != tc.want {
				t.Errorf("findMatchingBrace(%q, %d) = %d, want %d", tc.src, tc.openPos, got, tc.want)
			}
		})
	}
}
