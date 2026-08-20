package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

func newAssignmentsCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "assignments", Short: "Read assignments"}
	cmd.AddCommand(newAssignmentsListCommand(rc), newAssignmentsGetCommand(rc))
	return cmd
}

func newAssignmentsListCommand(rc *runtime) *cobra.Command {
	studentID, sectionID := "", ""
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List assignments for a section",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStudentSection(studentID, sectionID); err != nil {
				return err
			}
			page, err := rc.client.GetAssignments(cmd.Context(), studentID, sectionID)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(page.Assignments)
			}
			writeAssignmentTable(rc, page.Assignments)
			return nil
		},
	}
	cmd.Flags().StringVar(&studentID, "student", "", "student ID")
	cmd.Flags().StringVar(&sectionID, "section", "", "section ID")
	return cmd
}

func newAssignmentsGetCommand(rc *runtime) *cobra.Command {
	studentID, sectionID := "", ""
	cmd := &cobra.Command{
		Use:   "get <assignment-id>",
		Short: "Get one assignment",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStudentSection(studentID, sectionID); err != nil {
				return err
			}
			page, err := rc.client.GetAssignments(cmd.Context(), studentID, sectionID)
			if err != nil {
				return err
			}
			assignment, err := findAssignment(page.Assignments, args[0])
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(assignment)
			}
			writeAssignmentTable(rc, []api.AssignmentSummary{*assignment})
			return nil
		},
	}
	cmd.Flags().StringVar(&studentID, "student", "", "student ID")
	cmd.Flags().StringVar(&sectionID, "section", "", "section ID")
	return cmd
}

func findAssignment(
	assignments []api.AssignmentSummary,
	assignmentID string,
) (*api.AssignmentSummary, error) {
	for _, assignment := range assignments {
		if assignment.Assignment.ID == assignmentID {
			return &assignment, nil
		}
	}
	return nil, fmt.Errorf("assignment %q was not found", assignmentID)
}

func newGradesCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "grades", Short: "Read course grade summaries"}
	cmd.AddCommand(newGradesListCommand(rc))
	return cmd
}

func newGradesListCommand(rc *runtime) *cobra.Command {
	studentID, week := "", defaultWeek()
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List course grade summaries",
		RunE: func(cmd *cobra.Command, args []string) error {
			grades, err := rc.client.ListGradeSummaries(cmd.Context(), week, studentID)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(grades)
			}
			writeGradeTable(rc, grades)
			return nil
		},
	}
	cmd.Flags().StringVar(&studentID, "student", "", "student ID")
	cmd.Flags().StringVar(&week, "week", week, "week date (YYYY-MM-DD)")
	return cmd
}

func newAttendanceCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "attendance", Short: "Read section attendance"}
	cmd.AddCommand(newAttendanceListCommand(rc))
	return cmd
}

func newAttendanceListCommand(rc *runtime) *cobra.Command {
	studentID, sectionID := "", ""
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List section attendance",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireStudentSection(studentID, sectionID); err != nil {
				return err
			}
			page, err := rc.client.GetAttendance(cmd.Context(), studentID, sectionID)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(page)
			}
			writeAttendanceTable(rc, page)
			return nil
		},
	}
	cmd.Flags().StringVar(&studentID, "student", "", "student ID")
	cmd.Flags().StringVar(&sectionID, "section", "", "section ID")
	return cmd
}

func requireStudentSection(studentID, sectionID string) error {
	if studentID == "" {
		return fmt.Errorf("%w: --student is required", errUsage)
	}
	if sectionID == "" {
		return fmt.Errorf("%w: --section is required", errUsage)
	}
	return nil
}

func writeAssignmentTable(rc *runtime, assignments []api.AssignmentSummary) {
	rows := make([][]string, 0, len(assignments))
	for _, assignment := range assignments {
		rows = append(rows, []string{
			assignment.Assignment.ID,
			assignment.Assignment.Name,
			assignment.SectionGradingCategoryName,
			stringValue(assignment.SectionGradingCategoryAssignment.DateDue),
			assignment.AssignmentState,
		})
	}
	rc.out.Table([]string{"ID", "NAME", "CATEGORY", "DUE", "STATE"}, rows)
}

func writeGradeTable(rc *runtime, grades []api.GradeSummary) {
	rows := make([][]string, 0, len(grades))
	for _, grade := range grades {
		rows = append(rows, []string{
			grade.SectionID,
			grade.StudentName,
			grade.CourseName,
			formatGrade(courseRecord{Grade: grade.NumericGrade, LetterGrade: grade.LetterGrade}),
		})
	}
	rc.out.Table([]string{"SECTION ID", "STUDENT", "COURSE", "GRADE"}, rows)
}

func writeAttendanceTable(rc *runtime, page *api.AttendancePage) {
	markings := make(map[string]string, len(page.AttendanceMarkings))
	for _, marking := range page.AttendanceMarkings {
		markings[marking.ID] = marking.Value
	}
	rows := make([][]string, 0, len(page.StudentAttendance.Attendance))
	for _, record := range page.StudentAttendance.Attendance {
		rows = append(rows, []string{
			record.ID, record.Date, markings[record.AttendanceMarkingID],
		})
	}
	rc.out.Table([]string{"ID", "DATE", "MARKING"}, rows)
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
