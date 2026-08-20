package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListFamiliesUsesDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(directoryTestHandler))
	defer server.Close()

	directory, err := New(server.URL).ListFamilies(
		context.Background(),
		FamilyDirectoryQuery{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(directory.Families) != 1 || directory.Families[0].FamilyName != "Family" {
		t.Fatalf("directory = %#v", directory)
	}
}

func directoryTestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/Directory/GetDirectoryInfo":
		_, _ = fmt.Fprint(w, `{"Directories":[`+
			`{"ID":"directory-1","Name":"Families","IsFamilyDirectory":true}]}`)
	case "/Directory":
		_, _ = fmt.Fprint(w, `<script>`+
			`window.Directory.SchoolYearForTodayID = 'year-1';</script>`)
	case "/Directory/GetFamilyDirectoryUserInfo":
		if r.URL.Query().Get("DirectoryId") != "directory-1" ||
			r.URL.Query().Get("SchoolYearId") != "year-1" {
			http.Error(w, "unexpected query", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{"FamilyList":[{"FamilyId":"family-1",`+
			`"FamilyName":"Family"}],"PagingInfo":{"CurrentPage":1}}`)
	default:
		http.NotFound(w, r)
	}
}
