package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetQuickView(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Home/GetQuickView" ||
			r.URL.Query().Get("weekDate") != "2026-08-17T00:00:00" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{
			"Announcements":[{"Heading":"Notice","Description":"Body","Important":true}],
			"UserInfos":[{"UserID":"student-1","Name":"Student","Sections":[{
				"Course":{"ID":"course-1","Name":"Math"},
				"Section":{"ID":"section-1"},"LetterGrade":"A"
			}]}]
		}`)
	}))
	defer server.Close()

	quickView, err := New(server.URL).GetQuickView(context.Background(), "2026-08-17")
	if err != nil {
		t.Fatal(err)
	}
	if len(quickView.Students) != 1 || quickView.Students[0].Sections[0].Course.Name != "Math" {
		t.Fatalf("quick view = %#v", quickView)
	}
}

func TestGetCalendar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("startDate") != "2026-08-20" ||
			r.URL.Query().Get("endDate") != "2026-09-20" {
			http.Error(w, "unexpected dates", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{"CalendarEvents":[{"CalendarEvent":{"ID":"event-1","Name":"Event"}}]}`)
	}))
	defer server.Close()

	calendar, err := New(server.URL).GetCalendar(context.Background(), "2026-08-20", "2026-09-20")
	if err != nil {
		t.Fatal(err)
	}
	if len(calendar.Events) != 1 || calendar.Events[0].Event.Name != "Event" {
		t.Fatalf("calendar = %#v", calendar)
	}
}
