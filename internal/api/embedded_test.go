package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEmbeddedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<script>window.Assignments.model = `+
			`Utilities.convert(JSON.parse('{\"AssignmentsList\":[],`+
			`\"Title\":\"Parent\\u0027s Day\"}'));</script>`)
	}))
	defer server.Close()

	var model struct {
		Assignments []any  `json:"AssignmentsList"`
		Title       string `json:"Title"`
	}
	err := New(server.URL).getEmbeddedJSON(
		context.Background(),
		"/assignments",
		"window.Assignments.model",
		&model,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Assignments) != 0 || model.Title != "Parent's Day" {
		t.Fatalf("model = %#v", model)
	}
}

func TestFindEmbeddedJSONRequiresMarker(t *testing.T) {
	if _, err := findEmbeddedJSON([]byte(`<html></html>`), "window.Model"); err == nil {
		t.Fatal("expected missing marker error")
	}
}
