package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type AssignmentPage struct {
	Assignments       []AssignmentSummary `json:"AssignmentsList"`
	CanEditUnitGrades bool                `json:"CanEditUnitGrades"`
	SchemeAndUnitInfo SchemeAndUnitInfo   `json:"SchemeAndUnitInfo"`
	View              string              `json:"View"`
}

type AssignmentSummary struct {
	Assignment                       Assignment        `json:"Assignment"`
	AssignmentDueStatus              string            `json:"AssignmentDueStatus"`
	AssignmentState                  string            `json:"AssignmentState"`
	HasUnseenNotification            bool              `json:"HasUnseenNotification"`
	SectionGradingCategoryAssignment SectionAssignment `json:"SectionGradingCategoryAssignment"`
	SectionGradingCategoryName       string            `json:"SectionGradingCategoryName"`
}

type Assignment struct {
	Description string `json:"Description"`
	ID          string `json:"ID"`
	Name        string `json:"Name"`
}

type SectionAssignment struct {
	AllowSubmissions  bool     `json:"AllowSubmissions"`
	DateDue           *string  `json:"DateDue"`
	GradingCategoryID string   `json:"GradingCategory_ID"`
	ID                string   `json:"ID"`
	IsExtraCredit     bool     `json:"IsExtraCredit"`
	PointsPossible    *float64 `json:"PointsPossible"`
	SectionID         string   `json:"Section_ID"`
	StartDate         *string  `json:"StartDate"`
}

type SchemeAndUnitInfo struct {
	AreGradingUnitsWeighted bool   `json:"AreGradingUnitsWeighted"`
	CurrentGradingUnitID    string `json:"CurrentGradingUnit_ID"`
	GradingUnitSchemeID     string `json:"GradingUnitScheme_ID"`
}

type AttendancePage struct {
	AttendanceMarkings []AttendanceMarking `json:"attendanceMarkings"`
	ClassDays          []string            `json:"classDays"`
	StudentAttendance  StudentAttendance   `json:"studentAttendance"`
}

type AttendanceMarking struct {
	Color           string `json:"Color"`
	ID              string `json:"ID"`
	IsShownInTotals bool   `json:"IsShownInTotals"`
	ShortCode       string `json:"ShortCode"`
	Value           string `json:"Value"`
}

type StudentAttendance struct {
	Attendance []AttendanceRecord `json:"Attendance"`
	Totals     map[string]int     `json:"AttendanceMarkingTotals"`
}

type AttendanceRecord struct {
	AttendanceMarkingID string `json:"AttendanceMarking_ID"`
	Date                string `json:"Date"`
	ID                  string `json:"ID"`
	SectionID           string `json:"Section_ID"`
	StudentID           string `json:"User_ID"`
}

type GradeSummary struct {
	CourseID      string   `json:"course_id"`
	CourseName    string   `json:"course_name"`
	LetterGrade   string   `json:"letter_grade"`
	NumericGrade  *float64 `json:"numeric_grade"`
	SectionID     string   `json:"section_id"`
	SectionNumber string   `json:"section_number"`
	StudentID     string   `json:"student_id"`
	StudentName   string   `json:"student_name"`
}

func (c *Client) GetAssignments(
	ctx context.Context,
	studentID, sectionID string,
) (*AssignmentPage, error) {
	path, err := studentSectionResource(studentID, sectionID, "Assignments")
	if err != nil {
		return nil, err
	}
	var page AssignmentPage
	if err := c.getEmbeddedJSON(ctx, path, "window.Assignments.model", &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) GetAttendance(
	ctx context.Context,
	studentID, sectionID string,
) (*AttendancePage, error) {
	path, err := studentSectionResource(studentID, sectionID, "Attendance")
	if err != nil {
		return nil, err
	}
	var page AttendancePage
	marker := "window.AttendanceModule.model"
	if err := c.getEmbeddedJSON(ctx, path, marker, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

func (c *Client) ListGradeSummaries(
	ctx context.Context,
	weekDate, studentID string,
) ([]GradeSummary, error) {
	if _, err := time.Parse("2006-01-02", weekDate); err != nil {
		return nil, fmt.Errorf("invalid week date %q: use YYYY-MM-DD", weekDate)
	}
	view, err := c.GetQuickView(ctx, weekDate)
	if err != nil {
		return nil, err
	}
	var grades []GradeSummary
	for _, student := range view.Students {
		if studentID != "" && student.ID != studentID {
			continue
		}
		grades = append(grades, studentGradeSummaries(student)...)
	}
	return grades, nil
}

func studentGradeSummaries(student Student) []GradeSummary {
	grades := make([]GradeSummary, 0, len(student.Sections))
	for _, section := range student.Sections {
		grades = append(grades, GradeSummary{
			CourseID: section.Course.ID, CourseName: section.Course.Name,
			LetterGrade: section.LetterGrade, NumericGrade: section.Grade,
			SectionID: section.Section.ID, SectionNumber: section.Section.SectionNumber,
			StudentID: student.ID, StudentName: student.Name,
		})
	}
	return grades
}

func studentSectionResource(studentID, sectionID, resource string) (string, error) {
	if strings.TrimSpace(studentID) == "" {
		return "", fmt.Errorf("student ID is required")
	}
	if strings.TrimSpace(sectionID) == "" {
		return "", fmt.Errorf("section ID is required")
	}
	return fmt.Sprintf(
		"/Students/%s/Sections/%s/%s",
		url.PathEscape(studentID),
		url.PathEscape(sectionID),
		resource,
	), nil
}
