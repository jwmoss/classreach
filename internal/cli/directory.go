package cli

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

func newDirectoryCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "directory", Short: "Read school directories"}
	cmd.AddCommand(newDirectoryListCommand(rc), newDirectoryFamiliesCommand(rc))
	return cmd
}

func newDirectoryListCommand(rc *runtime) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List available directories",
		RunE: func(cmd *cobra.Command, args []string) error {
			directories, err := rc.client.ListDirectories(cmd.Context())
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(directories.Directories)
			}
			writeDirectoryTable(rc, directories.Directories)
			return nil
		},
	}
}

func newDirectoryFamiliesCommand(rc *runtime) *cobra.Command {
	query := api.FamilyDirectoryQuery{}
	cmd := &cobra.Command{
		Use:   "families",
		Short: "List or search the family directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			directory, err := rc.client.ListFamilies(cmd.Context(), query)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(directory)
			}
			writeFamilyTable(rc, directory.Families)
			rc.out.Printf(
				"Page %d of %d, %d total\n",
				directory.PagingInfo.CurrentPage,
				directory.PagingInfo.TotalPages,
				directory.PagingInfo.TotalItems,
			)
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringSliceVar(&query.AcademicLevelIDs, "academic-level", nil, "academic level ID")
	flags.BoolVar(&query.AscendingOrder, "ascending", false, "sort in ascending order")
	flags.StringVar(&query.DirectoryID, "directory", "", "family directory ID")
	flags.IntVar(&query.Page, "page", 1, "page number")
	flags.IntVar(&query.PerPage, "per-page", 25, "items per page")
	flags.StringVar(&query.SchoolYearID, "school-year", "", "school year ID")
	flags.StringVar(&query.SearchTerm, "search", "", "search family names")
	flags.StringVar(&query.SortProperty, "sort", "", "sort property")
	return cmd
}

func writeDirectoryTable(rc *runtime, directories []api.DirectoryInfo) {
	rows := make([][]string, 0, len(directories))
	for _, directory := range directories {
		rows = append(rows, []string{
			directory.ID,
			directory.Name,
			strconv.FormatBool(directory.IsFamilyDirectory),
		})
	}
	rc.out.Table([]string{"ID", "NAME", "FAMILY"}, rows)
}

func writeFamilyTable(rc *runtime, families []api.FamilyInfo) {
	rows := make([][]string, 0, len(families))
	for _, family := range families {
		rows = append(rows, []string{
			family.FamilyID,
			family.FamilyName,
			personNames(family.GuardianDetails),
			personNames(family.StudentDetails),
		})
	}
	rc.out.Table([]string{"ID", "FAMILY", "GUARDIANS", "STUDENTS"}, rows)
}

func personNames(people []api.DirectoryPerson) string {
	names := make([]string, 0, len(people))
	for _, person := range people {
		names = append(names, person.FullName)
	}
	return strings.Join(names, ", ")
}
