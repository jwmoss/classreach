package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type QuickView struct {
	Announcements            []Announcement `json:"Announcements"`
	DownloadAgendaForWeekURL string         `json:"DownloadAgendaForWeekUrl"`
	Students                 []Student      `json:"UserInfos"`
}

type Announcement struct {
	Description string `json:"Description"`
	Heading     string `json:"Heading"`
	Important   bool   `json:"Important"`
}

type Student struct {
	ID              string           `json:"UserID"`
	Name            string           `json:"Name"`
	Sections        []SectionSummary `json:"Sections"`
	UserSummaryLink string           `json:"UserSummaryLink"`
}

type SectionSummary struct {
	Course              Course   `json:"Course"`
	Grade               *float64 `json:"Grade"`
	HideNumericAverages bool     `json:"HideNumericAverages"`
	LetterGrade         string   `json:"LetterGrade"`
	NoLetterGrade       bool     `json:"NoLetterGrade"`
	Section             Section  `json:"Section"`
	SectionURL          string   `json:"SectionUrl"`
	UnitName            *string  `json:"UnitName"`
}

type Course struct {
	CourseNumber string `json:"CourseNumber"`
	ID           string `json:"ID"`
	Name         string `json:"Name"`
}

type Section struct {
	AcademicTermID string `json:"AcademicTerm_ID"`
	ID             string `json:"ID"`
	SectionNumber  string `json:"SectionNumber"`
}

type Calendar struct {
	Events []CalendarEventInfo `json:"CalendarEvents"`
}

type CalendarEventInfo struct {
	Event      CalendarEvent `json:"CalendarEvent"`
	IsEditable bool          `json:"IsEventEditable"`
}

type CalendarEvent struct {
	AllDay      bool    `json:"AllDay"`
	Description *string `json:"Description"`
	EndTime     string  `json:"EndTime"`
	ID          string  `json:"ID"`
	Location    *string `json:"Location"`
	Name        string  `json:"Name"`
	SectionID   *string `json:"Section_ID"`
	StartTime   string  `json:"StartTime"`
	URL         *string `json:"Url"`
}

func (c *Client) GetQuickView(ctx context.Context, weekDate string) (*QuickView, error) {
	if strings.TrimSpace(weekDate) == "" {
		return nil, fmt.Errorf("week date is required")
	}
	query := url.Values{"weekDate": {weekDate + "T00:00:00"}}
	var quickView QuickView
	request := JSONRequest{Method: http.MethodGet, Path: "/Home/GetQuickView", Query: query}
	if err := c.DoJSON(ctx, request, &quickView); err != nil {
		return nil, err
	}
	return &quickView, nil
}

func (c *Client) DownloadAgenda(ctx context.Context, weekDate string) ([]byte, error) {
	quickView, err := c.GetQuickView(ctx, weekDate)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(quickView.DownloadAgendaForWeekURL) == "" {
		return nil, fmt.Errorf("quick view did not include an agenda download URL")
	}
	return c.Do(ctx, http.MethodGet, quickView.DownloadAgendaForWeekURL, nil, nil)
}

func (c *Client) GetCalendar(ctx context.Context, startDate, endDate string) (*Calendar, error) {
	if strings.TrimSpace(startDate) == "" || strings.TrimSpace(endDate) == "" {
		return nil, fmt.Errorf("start and end dates are required")
	}
	query := url.Values{
		"startDate": {startDate},
		"endDate":   {endDate},
	}
	var calendar Calendar
	request := JSONRequest{Method: http.MethodGet, Path: "/Calendar/events", Query: query}
	if err := c.DoJSON(ctx, request, &calendar); err != nil {
		return nil, err
	}
	return &calendar, nil
}
