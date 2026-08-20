package cli

import (
	"testing"

	"github.com/jwmoss/classreach/internal/api"
)

func TestCourseRecordsFiltersStudent(t *testing.T) {
	students := []api.Student{
		{ID: "student-1", Name: "One", Sections: []api.SectionSummary{{
			Course:  api.Course{ID: "course-1", Name: "Math"},
			Section: api.Section{ID: "section-1"},
		}}},
		{ID: "student-2", Name: "Two", Sections: []api.SectionSummary{{
			Course:  api.Course{ID: "course-2", Name: "History"},
			Section: api.Section{ID: "section-2"},
		}}},
	}

	records := courseRecords(students, "student-2")
	if len(records) != 1 || records[0].CourseName != "History" {
		t.Fatalf("records = %#v", records)
	}
}

func TestValidateDateRange(t *testing.T) {
	if err := validateDateRange("2026-08-20", "2026-08-19"); err == nil {
		t.Fatal("expected invalid date range")
	}
	if err := validateDateRange("2026-08-20", "2026-08-21"); err != nil {
		t.Fatal(err)
	}
}

func TestCleanHTML(t *testing.T) {
	if got := cleanHTML("<p>Hello &amp; welcome</p>"); got != "Hello & welcome" {
		t.Fatalf("clean HTML = %q", got)
	}
}
