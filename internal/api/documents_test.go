package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSchoolDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("folderID") != "folder-1" {
			http.Error(w, "missing folder", http.StatusBadRequest)
			return
		}
		_, _ = fmt.Fprint(w, `{
			"SchoolDocumentsFoldersListItems":[{"ID":"folder-2","Name":"Forms"}],
			"SchoolDocumentsListItems":[{"ID":"document-1","Name":"Calendar",
			"FileInfo":{"DownloadUrl":"/Files/document-1"}}]
		}`)
	}))
	defer server.Close()

	documents, err := New(server.URL).GetSchoolDocuments(context.Background(), "folder-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(documents.Documents) != 1 || documents.Documents[0].DownloadURL() != "/Files/document-1" {
		t.Fatalf("documents = %#v", documents)
	}
}

func TestDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("file contents"))
	}))
	defer server.Close()

	data, err := New(server.URL).Download(context.Background(), "/file")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "file contents" {
		t.Fatalf("download = %q", data)
	}
}
