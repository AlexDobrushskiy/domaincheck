package main

import (
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

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(results, false)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 JSON lines, got %d: %s", len(lines), output)
	}

	var jr jsonResult
	json.Unmarshal([]byte(lines[0]), &jr)
	if jr.Domain != "example.com" || jr.Available != false {
		t.Errorf("unexpected first result: %+v", jr)
	}

	json.Unmarshal([]byte(lines[1]), &jr)
	if jr.Domain != "free.com" || jr.Available != true {
		t.Errorf("unexpected second result: %+v", jr)
	}

	json.Unmarshal([]byte(lines[2]), &jr)
	if jr.Domain != "err.com" || jr.Error != "lookup failed" {
		t.Errorf("unexpected third result: %+v", jr)
	}
}

func TestPrintJSON_AvailableOnly(t *testing.T) {
	now := time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
	results := []Result{
		{Domain: "taken.com", Available: false, CheckedAt: now},
		{Domain: "free.com", Available: true, CheckedAt: now},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(results, true)

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 1 {
		t.Fatalf("expected 1 JSON line, got %d: %s", len(lines), output)
	}

	var jr jsonResult
	json.Unmarshal([]byte(lines[0]), &jr)
	if jr.Domain != "free.com" || jr.Available != true {
		t.Errorf("unexpected result: %+v", jr)
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
