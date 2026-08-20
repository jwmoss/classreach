package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

func newDocumentsCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "documents", Short: "Read and download school documents"}
	cmd.AddCommand(newDocumentsListCommand(rc), newDocumentsDownloadCommand(rc))
	return cmd
}

func newDocumentsListCommand(rc *runtime) *cobra.Command {
	folderID := ""
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List school document folders and files",
		RunE: func(cmd *cobra.Command, args []string) error {
			documents, err := rc.client.GetSchoolDocuments(cmd.Context(), folderID)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(documents)
			}
			writeDocumentsTable(rc, documents)
			return nil
		},
	}
	cmd.Flags().StringVar(&folderID, "folder", "", "folder ID")
	return cmd
}

func newDocumentsDownloadCommand(rc *runtime) *cobra.Command {
	folderID, outputPath, force := "", "", false
	cmd := &cobra.Command{
		Use:   "download <document-id>",
		Short: "Download one school document",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputPath == "" {
				return fmt.Errorf("%w: --output is required", errUsage)
			}
			if fileExists(outputPath) && !force {
				return fmt.Errorf("output exists at %s; use --force to overwrite", outputPath)
			}
			document, err := findDocument(cmd, rc, folderID, args[0])
			if err != nil {
				return err
			}
			data, err := rc.client.Download(cmd.Context(), document.DownloadURL())
			if err != nil {
				return err
			}
			if err := os.WriteFile(outputPath, data, 0600); err != nil {
				return fmt.Errorf("write document %s: %w", outputPath, err)
			}
			rc.out.Success("document downloaded")
			rc.out.Printf("%s\n", outputPath)
			return nil
		},
	}
	cmd.Flags().StringVar(&folderID, "folder", "", "folder ID containing the document")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")
	return cmd
}

func findDocument(
	cmd *cobra.Command,
	rc *runtime,
	folderID, documentID string,
) (*api.SchoolDocument, error) {
	documents, err := rc.client.GetSchoolDocuments(cmd.Context(), folderID)
	if err != nil {
		return nil, err
	}
	for _, document := range documents.Documents {
		if document.ID == documentID {
			if document.DownloadURL() == "" {
				return nil, fmt.Errorf("document %q has no download URL", documentID)
			}
			return &document, nil
		}
	}
	return nil, fmt.Errorf("document %q is not in the selected folder", documentID)
}

func writeDocumentsTable(rc *runtime, documents *api.SchoolDocuments) {
	rows := make([][]string, 0, len(documents.Folders)+len(documents.Documents))
	for _, folder := range documents.Folders {
		rows = append(rows, []string{"folder", folder.ID, folder.Name, ""})
	}
	for _, document := range documents.Documents {
		rows = append(rows, []string{
			"document", document.ID, document.Name,
			strconv.FormatInt(document.FileInfo.File.Size, 10),
		})
	}
	rc.out.Table([]string{"TYPE", "ID", "NAME", "BYTES"}, rows)
}
