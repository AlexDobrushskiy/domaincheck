package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckDomain_Taken(t *testing.T) {
	result := CheckDomain("google.com")
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Available {
		t.Error("expected google.com to be taken")
	}
}

func TestCheckDomain_Available(t *testing.T) {
	result := CheckDomain("notarealdomain12345.com")
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Available {
		t.Error("expected notarealdomain12345.com to be available")
	}
}

func TestCheckDomains_Concurrency(t *testing.T) {
	domains := []string{"google.com", "notarealdomain12345.com"}
	results := CheckDomains(domains, 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Results should be in order
	if results[0].Domain != "google.com" {
		t.Errorf("expected first result to be google.com, got %s", results[0].Domain)
	}
	if results[1].Domain != "notarealdomain12345.com" {
		t.Errorf("expected second result to be notarealdomain12345.com, got %s", results[1].Domain)
	}
}

func TestReadDomainsFromFile(t *testing.T) {
	content := "example.com\n\n# comment\ntest.net\n  spaced.org  \n"
	tmp := filepath.Join(t.TempDir(), "domains.txt")
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	domains, err := readDomainsFromFile(tmp)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"example.com", "test.net", "spaced.org"}
	if len(domains) != len(expected) {
		t.Fatalf("expected %d domains, got %d", len(expected), len(domains))
	}
	for i, d := range domains {
		if d != expected[i] {
			t.Errorf("domain[%d]: expected %q, got %q", i, expected[i], d)
		}
	}
}
