package main

import (
	"strings"
	"testing"
)

func TestParseCIDRs(t *testing.T) {
	got, err := parseCIDRs([]string{
		"  1.2.3.4/24  ",
		"",
		"# comment",
		"2001:db8::/32",
	})
	if err != nil {
		t.Fatalf("parseCIDRs returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d networks, want 2", len(got))
	}
	if got[0].String() != "1.2.3.0/24" {
		t.Errorf("first network = %s, want 1.2.3.0/24", got[0])
	}
	if got[1].String() != "2001:db8::/32" {
		t.Errorf("second network = %s, want 2001:db8::/32", got[1])
	}
}

func TestParseCIDRsRejectsInvalidLine(t *testing.T) {
	_, err := parseCIDRs([]string{"10.0.0.0/8", "not-a-cidr"})
	if err == nil {
		t.Fatal("expected invalid CIDR to return an error")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("error %q does not identify line 2", err)
	}
}

func TestParseCIDRsAllowsEmptyInput(t *testing.T) {
	got, err := parseCIDRs([]string{"", " # comment"})
	if err != nil {
		t.Fatalf("parseCIDRs returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d networks, want 0", len(got))
	}
}
