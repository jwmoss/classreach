package cli

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

const dateLayout = "2006-01-02"

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

type courseRecord struct {
	StudentID     string   `json:"student_id"`
	StudentName   string   `json:"student_name"`
	CourseID      string   `json:"course_id"`
	CourseName    string   `json:"course_name"`
	SectionID     string   `json:"section_id"`
	SectionNumber string   `json:"section_number,omitempty"`
	Grade         *float64 `json:"grade,omitempty"`
	LetterGrade   string   `json:"letter_grade,omitempty"`
	URL           string   `json:"url"`
}

func newOverviewCommand(rc *runtime) *cobra.Command {
	week := defaultWeek()
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Show the guardian quick view",
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(map[string]any{"week": week, "quick_view": quickView})
			}
			rc.out.Printf("Week of %s\n", week)
			rc.out.Printf("Students: %d\n", len(quickView.Students))
			rc.out.Printf("Announcements: %d\n", len(quickView.Announcements))
			for _, student := range quickView.Students {
				rc.out.Printf("- %s: %d courses\n", student.Name, len(student.Sections))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	return cmd
}

func newStudentsCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "students", Short: "Read visible students"}
	cmd.AddCommand(newStudentsListCommand(rc), newStudentsGetCommand(rc))
	return cmd
}

func newStudentsListCommand(rc *runtime) *cobra.Command {
	week := defaultWeek()
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List visible students",
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(quickView.Students)
			}
			rows := make([][]string, 0, len(quickView.Students))
			for _, student := range quickView.Students {
				rows = append(rows, []string{student.ID, student.Name, strconv.Itoa(len(student.Sections))})
			}
			rc.out.Table([]string{"ID", "NAME", "COURSES"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	return cmd
}

func newStudentsGetCommand(rc *runtime) *cobra.Command {
	week := defaultWeek()
	cmd := &cobra.Command{
		Use:   "get <student-id>",
		Short: "Get one visible student",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			for _, student := range quickView.Students {
				if student.ID == args[0] {
					return writeStudent(rc, student)
				}
			}
			return fmt.Errorf("student %q is not visible to this account", args[0])
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	return cmd
}

func writeStudent(rc *runtime, student api.Student) error {
	if rc.out.IsJSON() {
		return rc.out.JSON(student)
	}
	rc.out.Table([]string{"KEY", "VALUE"}, [][]string{
		{"id", student.ID},
		{"name", student.Name},
		{"courses", strconv.Itoa(len(student.Sections))},
		{"summary_url", student.UserSummaryLink},
	})
	return nil
}

func newCoursesCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "courses", Short: "Read visible courses"}
	cmd.AddCommand(newCoursesListCommand(rc), newCoursesGetCommand(rc))
	return cmd
}

func newCoursesListCommand(rc *runtime) *cobra.Command {
	week, studentID := defaultWeek(), ""
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List visible courses",
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			records := courseRecords(quickView.Students, studentID)
			if rc.out.IsJSON() {
				return rc.out.JSON(records)
			}
			writeCourseTable(rc, records)
			return nil
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	cmd.Flags().StringVar(&studentID, "student", "", "filter by student ID")
	return cmd
}

func newCoursesGetCommand(rc *runtime) *cobra.Command {
	week := defaultWeek()
	cmd := &cobra.Command{
		Use:   "get <section-id>",
		Short: "Get one visible course section",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			for _, record := range courseRecords(quickView.Students, "") {
				if record.SectionID == args[0] {
					if rc.out.IsJSON() {
						return rc.out.JSON(record)
					}
					writeCourseTable(rc, []courseRecord{record})
					return nil
				}
			}
			return fmt.Errorf("section %q is not visible to this account", args[0])
		},
	}
	cmd.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	return cmd
}

func courseRecords(students []api.Student, studentID string) []courseRecord {
	records := []courseRecord{}
	for _, student := range students {
		if studentID != "" && student.ID != studentID {
			continue
		}
		for _, section := range student.Sections {
			records = append(records, courseRecord{
				StudentID: student.ID, StudentName: student.Name,
				CourseID: section.Course.ID, CourseName: section.Course.Name,
				SectionID: section.Section.ID, SectionNumber: section.Section.SectionNumber,
				Grade: section.Grade, LetterGrade: section.LetterGrade, URL: section.SectionURL,
			})
		}
	}
	return records
}

func writeCourseTable(rc *runtime, records []courseRecord) {
	rows := make([][]string, 0, len(records))
	for _, record := range records {
		rows = append(rows, []string{
			record.SectionID, record.StudentName, record.CourseName, formatGrade(record),
		})
	}
	rc.out.Table([]string{"SECTION ID", "STUDENT", "COURSE", "GRADE"}, rows)
}

func formatGrade(record courseRecord) string {
	if record.LetterGrade != "" {
		return record.LetterGrade
	}
	if record.Grade != nil {
		return strconv.FormatFloat(*record.Grade, 'f', 1, 64)
	}
	return ""
}

func newAnnouncementsCommand(rc *runtime) *cobra.Command {
	week := defaultWeek()
	cmd := &cobra.Command{Use: "announcements", Short: "Read school announcements"}
	list := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List school announcements",
		RunE: func(cmd *cobra.Command, args []string) error {
			quickView, err := loadQuickView(cmd, rc, week)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(quickView.Announcements)
			}
			rows := make([][]string, 0, len(quickView.Announcements))
			for _, announcement := range quickView.Announcements {
				rows = append(rows, []string{
					announcement.Heading, cleanHTML(announcement.Description),
					strconv.FormatBool(announcement.Important),
				})
			}
			rc.out.Table([]string{"HEADING", "DESCRIPTION", "IMPORTANT"}, rows)
			return nil
		},
	}
	list.Flags().StringVar(&week, "week", week, "week date in YYYY-MM-DD format")
	cmd.AddCommand(list)
	return cmd
}

func newCalendarCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "calendar", Short: "Read calendar events"}
	cmd.AddCommand(newCalendarListCommand(rc))
	return cmd
}

func newCalendarListCommand(rc *runtime) *cobra.Command {
	startDate := time.Now().Format(dateLayout)
	endDate := time.Now().AddDate(0, 1, 0).Format(dateLayout)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List calendar events",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateDateRange(startDate, endDate); err != nil {
				return err
			}
			calendar, err := rc.client.GetCalendar(cmd.Context(), startDate, endDate)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(calendar.Events)
			}
			rows := make([][]string, 0, len(calendar.Events))
			for _, item := range calendar.Events {
				rows = append(rows, []string{item.Event.StartTime, item.Event.EndTime, item.Event.Name})
			}
			rc.out.Table([]string{"START", "END", "NAME"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&startDate, "start", startDate, "start date in YYYY-MM-DD format")
	cmd.Flags().StringVar(&endDate, "end", endDate, "end date in YYYY-MM-DD format")
	return cmd
}

func loadQuickView(cmd *cobra.Command, rc *runtime, week string) (*api.QuickView, error) {
	if _, err := time.Parse(dateLayout, week); err != nil {
		return nil, fmt.Errorf("invalid week date %q: use YYYY-MM-DD", week)
	}
	return rc.client.GetQuickView(cmd.Context(), week)
}

func validateDateRange(startDate, endDate string) error {
	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return fmt.Errorf("invalid start date %q: use YYYY-MM-DD", startDate)
	}
	end, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return fmt.Errorf("invalid end date %q: use YYYY-MM-DD", endDate)
	}
	if end.Before(start) {
		return fmt.Errorf("end date must be on or after start date")
	}
	return nil
}

func defaultWeek() string {
	now := time.Now()
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	return now.AddDate(0, 0, -daysSinceMonday).Format(dateLayout)
}

func cleanHTML(value string) string {
	withoutTags := htmlTagPattern.ReplaceAllString(value, " ")
	return strings.Join(strings.Fields(html.UnescapeString(withoutTags)), " ")
}
