package cli

import (
	"testing"

	"github.com/jwmoss/classreach/internal/api"
)

func TestFindAssignment(t *testing.T) {
	assignments := []api.AssignmentSummary{{
		Assignment: api.Assignment{ID: "assignment-1", Name: "Essay"},
	}}

	assignment, err := findAssignment(assignments, "assignment-1")
	if err != nil {
		t.Fatal(err)
	}
	if assignment.Assignment.Name != "Essay" {
		t.Fatalf("assignment = %#v", assignment)
	}
}

func TestRequireStudentSection(t *testing.T) {
	if err := requireStudentSection("", "section"); err == nil {
		t.Fatal("expected missing student error")
	}
	if err := requireStudentSection("student", ""); err == nil {
		t.Fatal("expected missing section error")
	}
	if err := requireStudentSection("student", "section"); err != nil {
		t.Fatal(err)
	}
}
