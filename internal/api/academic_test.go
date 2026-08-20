package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAssignments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<script>window.Assignments.model = convert(JSON.parse(`+
			`'{\"AssignmentsList\":[{\"Assignment\":{\"ID\":\"a1\",`+
			`\"Name\":\"Essay\"}}]}'));</script>`)
	}))
	defer server.Close()

	page, err := New(server.URL).GetAssignments(context.Background(), "student", "section")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Assignments) != 1 || page.Assignments[0].Assignment.Name != "Essay" {
		t.Fatalf("page = %#v", page)
	}
}

func TestGetAttendance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<script>window.AttendanceModule.model = convert(JSON.parse(`+
			`'{\"classDays\":[\"2026-01-01\"],\"studentAttendance\":`+
			`{\"Attendance\":[{\"ID\":\"record-1\"}]}}'));</script>`)
	}))
	defer server.Close()

	page, err := New(server.URL).GetAttendance(context.Background(), "student", "section")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.StudentAttendance.Attendance) != 1 {
		t.Fatalf("page = %#v", page)
	}
}

func TestListGradeSummariesRejectsInvalidWeek(t *testing.T) {
	_, err := New("").ListGradeSummaries(context.Background(), "not-a-date", "")
	if err == nil {
		t.Fatal("expected invalid week error")
	}
}

func TestListGradeSummaries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"UserInfos":[{"UserID":"student-1","Name":"Student",`+
			`"Sections":[{"Course":{"ID":"course-1","Name":"Math"},`+
			`"Section":{"ID":"section-1","SectionNumber":"01"},"LetterGrade":"A"}]}]}`)
	}))
	defer server.Close()

	grades, err := New(server.URL).ListGradeSummaries(
		context.Background(),
		"2026-01-01",
		"student-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(grades) != 1 || grades[0].LetterGrade != "A" {
		t.Fatalf("grades = %#v", grades)
	}
}
