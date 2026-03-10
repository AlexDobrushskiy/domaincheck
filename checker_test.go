package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestPrintJSON(t *testing.T) {
	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	results := []Result{
		{Domain: "example.com", Available: false, CheckedAt: now},
		{Domain: "free.com", Available: true, CheckedAt: now},
		{Domain: "err.com", Err: errors.New("lookup failed"), CheckedAt: now},
	}

	var buf bytes.Buffer
	availCount, err := printJSON(&buf, results, false)
	if err != nil {
		t.Fatalf("printJSON returned error: %v", err)
	}
	if availCount != 1 {
		t.Errorf("expected availCount 1, got %d", availCount)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 JSON lines, got %d: %s", len(lines), buf.String())
	}

	expectedCheckedAt := "2026-03-06T12:00:00Z"

	var jr jsonResult
	if err := json.Unmarshal([]byte(lines[0]), &jr); err != nil {
		t.Fatalf("failed to unmarshal line 0: %v", err)
	}
	if jr.Domain != "example.com" || jr.Available == nil || *jr.Available != false {
		t.Errorf("unexpected first result: %+v", jr)
	}
	if jr.CheckedAt != expectedCheckedAt {
		t.Errorf("expected checked_at %q, got %q", expectedCheckedAt, jr.CheckedAt)
	}

	if err := json.Unmarshal([]byte(lines[1]), &jr); err != nil {
		t.Fatalf("failed to unmarshal line 1: %v", err)
	}
	if jr.Domain != "free.com" || jr.Available == nil || *jr.Available != true {
		t.Errorf("unexpected second result: %+v", jr)
	}
	if jr.CheckedAt != expectedCheckedAt {
		t.Errorf("expected checked_at %q, got %q", expectedCheckedAt, jr.CheckedAt)
	}

	if err := json.Unmarshal([]byte(lines[2]), &jr); err != nil {
		t.Fatalf("failed to unmarshal line 2: %v", err)
	}
	if jr.Domain != "err.com" || jr.Error != "lookup failed" {
		t.Errorf("unexpected third result: %+v", jr)
	}
	if jr.Available != nil {
		t.Errorf("expected available to be null for error result, got %v", *jr.Available)
	}
	if jr.CheckedAt != expectedCheckedAt {
		t.Errorf("expected checked_at %q, got %q", expectedCheckedAt, jr.CheckedAt)
	}
}

func TestPrintJSON_AvailableOnly(t *testing.T) {
	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	results := []Result{
		{Domain: "taken.com", Available: false, CheckedAt: now},
		{Domain: "free.com", Available: true, CheckedAt: now},
	}

	var buf bytes.Buffer
	availCount, err := printJSON(&buf, results, true)
	if err != nil {
		t.Fatalf("printJSON returned error: %v", err)
	}
	if availCount != 1 {
		t.Errorf("expected availCount 1, got %d", availCount)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")

	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d: %s", len(lines), buf.String())
	}

	var jr jsonResult
	if err := json.Unmarshal([]byte(lines[0]), &jr); err != nil {
		t.Fatalf("failed to unmarshal line 0: %v", err)
	}
	if jr.Domain != "free.com" || jr.Available == nil || *jr.Available != true {
		t.Errorf("unexpected result: %+v", jr)
	}
	if jr.CheckedAt != "2026-03-06T12:00:00Z" {
		t.Errorf("expected checked_at %q, got %q", "2026-03-06T12:00:00Z", jr.CheckedAt)
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		domain  string
		wantErr bool
	}{
		{"example.com", false},
		{"sub.example.com", false},
		{"my-site.co.uk", false},
		{"a.io", false},
		{"", true},
		{"nodot", true},
		{"has space.com", true},
		{"tab\there.com", true},
		{"special!char.com", true},
		{"under_score.com", true},
		{"trailing-.com", true},
		{"-leading.com", true},
		{"double..dot.com", true},
		{".leading-dot.com", true},
		{"trailing-dot.com.", true},
		{"http://example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			err := validateDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDomain(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
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
