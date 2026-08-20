package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type SchoolDocuments struct {
	FolderHierarchy []DocumentFolder `json:"FolderHierarchy"`
	Folders         []DocumentFolder `json:"SchoolDocumentsFoldersListItems"`
	Documents       []SchoolDocument `json:"SchoolDocumentsListItems"`
}

type DocumentFolder struct {
	FolderURL      string `json:"FolderUrl"`
	ID             string `json:"ID"`
	Name           string `json:"Name"`
	ParentFolderID string `json:"ParentFolderID"`
}

type SchoolDocument struct {
	Description    string           `json:"Description"`
	EndDate        *string          `json:"EndDate"`
	FileInfo       DocumentFileInfo `json:"FileInfo"`
	ID             string           `json:"ID"`
	Name           string           `json:"Name"`
	ParentFolderID string           `json:"ParentFolderID"`
	StartDate      string           `json:"StartDate"`
}

type DocumentFileInfo struct {
	DownloadURL string       `json:"DownloadUrl"`
	File        DocumentFile `json:"File"`
}

type DocumentFile struct {
	DownloadLink string `json:"FileDownloadLink"`
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	Size         int64  `json:"Size"`
	Type         string `json:"Type"`
}

func (d SchoolDocument) DownloadURL() string {
	if d.FileInfo.DownloadURL != "" {
		return d.FileInfo.DownloadURL
	}
	return d.FileInfo.File.DownloadLink
}

func (c *Client) GetSchoolDocuments(
	ctx context.Context,
	folderID string,
) (*SchoolDocuments, error) {
	query := url.Values{}
	if strings.TrimSpace(folderID) != "" {
		query.Set("folderID", folderID)
	}
	var documents SchoolDocuments
	request := JSONRequest{Method: http.MethodGet, Path: "/SchoolDocuments", Query: query}
	if err := c.DoJSON(ctx, request, &documents); err != nil {
		return nil, err
	}
	return &documents, nil
}

func (c *Client) Download(ctx context.Context, downloadURL string) ([]byte, error) {
	if strings.TrimSpace(downloadURL) == "" {
		return nil, fmt.Errorf("download URL is required")
	}
	return c.Do(ctx, http.MethodGet, downloadURL, nil, nil)
}
