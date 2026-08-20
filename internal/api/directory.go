package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type DirectoryList struct {
	Directories []DirectoryInfo `json:"Directories"`
}

type DirectoryInfo struct {
	ID                string `json:"ID"`
	IsFamilyDirectory bool   `json:"IsFamilyDirectory"`
	Name              string `json:"Name"`
	SortOrder         int    `json:"SortOrder"`
}

type FamilyDirectoryQuery struct {
	AcademicLevelIDs []string
	AscendingOrder   bool
	DirectoryID      string
	Page             int
	PerPage          int
	SchoolYearID     string
	SearchTerm       string
	SortProperty     string
}

type FamilyDirectory struct {
	CustomContents          *string      `json:"CustomContents"`
	DirectoryID             string       `json:"DirectoryId"`
	Families                []FamilyInfo `json:"FamilyList"`
	Name                    string       `json:"Name"`
	PagingInfo              PagingInfo   `json:"PagingInfo"`
	SchoolName              string       `json:"SchoolName"`
	ShowGuardianEmail       bool         `json:"ShowGuardianEmail"`
	ShowGuardianMobilePhone bool         `json:"ShowGuardianMobilePhone"`
	ShowStudentEmail        bool         `json:"ShowStudentEmail"`
	ShowStudentMobilePhone  bool         `json:"ShowStudentMobilePhone"`
}

type FamilyInfo struct {
	FamilyID        string            `json:"FamilyId"`
	FamilyName      string            `json:"FamilyName"`
	Fields          []DirectoryField  `json:"FieldDisplay"`
	GuardianDetails []DirectoryPerson `json:"GuardianDetails"`
	StudentDetails  []DirectoryPerson `json:"StudentDetails"`
}

type DirectoryPerson struct {
	Email                string           `json:"Email"`
	Fields               []DirectoryField `json:"FieldDisplay"`
	FullName             string           `json:"FullName"`
	MobilePhone          string           `json:"MobilePhone"`
	StudentAcademicLevel string           `json:"StudentAcademicLevel"`
	UserID               string           `json:"UserID"`
}

type DirectoryField struct {
	DirectoryFieldID string `json:"DirectoryFieldId"`
	FieldName        string `json:"FieldName"`
	FieldValue       string `json:"FieldValue"`
}

var currentSchoolYearPattern = regexp.MustCompile(
	`window\.Directory\.SchoolYearForTodayID\s*=\s*'([^']+)'`,
)

func (c *Client) ListDirectories(ctx context.Context) (*DirectoryList, error) {
	var directories DirectoryList
	request := JSONRequest{Method: http.MethodGet, Path: "/Directory/GetDirectoryInfo"}
	if err := c.DoJSON(ctx, request, &directories); err != nil {
		return nil, err
	}
	return &directories, nil
}

func (c *Client) ListFamilies(
	ctx context.Context,
	query FamilyDirectoryQuery,
) (*FamilyDirectory, error) {
	if err := c.completeFamilyDirectoryQuery(ctx, &query); err != nil {
		return nil, err
	}
	values := familyDirectoryValues(query)
	var directory FamilyDirectory
	request := JSONRequest{
		Method: http.MethodGet,
		Path:   "/Directory/GetFamilyDirectoryUserInfo",
		Query:  values,
	}
	if err := c.DoJSON(ctx, request, &directory); err != nil {
		return nil, err
	}
	return &directory, nil
}

func (c *Client) completeFamilyDirectoryQuery(
	ctx context.Context,
	query *FamilyDirectoryQuery,
) error {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 25
	}
	if query.DirectoryID == "" {
		directoryID, err := c.defaultFamilyDirectoryID(ctx)
		if err != nil {
			return err
		}
		query.DirectoryID = directoryID
	}
	if query.SchoolYearID == "" {
		schoolYearID, err := c.currentSchoolYearID(ctx)
		if err != nil {
			return err
		}
		query.SchoolYearID = schoolYearID
	}
	return nil
}

func (c *Client) defaultFamilyDirectoryID(ctx context.Context) (string, error) {
	directories, err := c.ListDirectories(ctx)
	if err != nil {
		return "", err
	}
	for _, directory := range directories.Directories {
		if directory.IsFamilyDirectory {
			return directory.ID, nil
		}
	}
	return "", fmt.Errorf("no family directory is available")
}

func (c *Client) currentSchoolYearID(ctx context.Context) (string, error) {
	body, err := c.Do(ctx, http.MethodGet, "/Directory", nil, nil)
	if err != nil {
		return "", err
	}
	match := currentSchoolYearPattern.FindSubmatch(body)
	if len(match) != 2 || strings.TrimSpace(string(match[1])) == "" {
		return "", fmt.Errorf("current directory school year was not found")
	}
	return string(match[1]), nil
}

func familyDirectoryValues(query FamilyDirectoryQuery) url.Values {
	values := url.Values{
		"AscendingOrder": {strconv.FormatBool(query.AscendingOrder)},
		"DirectoryId":    {query.DirectoryID},
		"Page":           {strconv.Itoa(query.Page)},
		"PerPage":        {strconv.Itoa(query.PerPage)},
		"SchoolYearId":   {query.SchoolYearID},
	}
	if query.SearchTerm != "" {
		values.Set("SearchTerm", query.SearchTerm)
	}
	if query.SortProperty != "" {
		values.Set("SortProperty", query.SortProperty)
	}
	for _, id := range query.AcademicLevelIDs {
		values.Add("AcademicLevelIds", id)
	}
	return values
}
