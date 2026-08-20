package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jwmoss/classreach/internal/api"
)

func newMessagesCommand(rc *runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "messages", Short: "Read message threads"}
	cmd.AddCommand(
		newMessagesListCommand(rc),
		newMessagesGetCommand(rc),
		newMessagesDownloadCommand(rc),
	)
	return cmd
}

func newMessagesListCommand(rc *runtime) *cobra.Command {
	label, searchTerm, page := "Inbox", "", 1
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List message threads",
		RunE: func(cmd *cobra.Command, args []string) error {
			messages, err := rc.client.ListMessages(cmd.Context(), label, searchTerm, page)
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(messages)
			}
			writeMessageTable(rc, messages.MessageThreads)
			rc.out.Printf(
				"Page %d of %d, %d total\n",
				messages.PagingInfo.CurrentPage,
				messages.PagingInfo.TotalPages,
				messages.PagingInfo.TotalItems,
			)
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "label", label, "message label")
	cmd.Flags().StringVar(&searchTerm, "search", "", "search message threads")
	cmd.Flags().IntVar(&page, "page", page, "page number")
	return cmd
}

func newMessagesGetCommand(rc *runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "get <thread-id>",
		Short: "Get one message thread",
		Long:  "Get one message thread. ClassReach can mark an unread thread as read.",
		Args:  usageArgs(cobra.ExactArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			response, err := rc.client.GetMessageThread(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			if rc.out.IsJSON() {
				return rc.out.JSON(response)
			}
			thread := response.MessageThreadViewModel
			rc.out.Printf("Subject: %s\n", thread.MessageThread.Subject)
			for _, message := range thread.Messages {
				writeMessage(rc, message)
			}
			return nil
		},
	}
}

func newMessagesDownloadCommand(rc *runtime) *cobra.Command {
	outputPath, force := "", false
	cmd := &cobra.Command{
		Use:   "download <thread-id> <file-id>",
		Short: "Download one message attachment",
		Long: "Download one message attachment. " +
			"ClassReach can mark an unread thread as read.",
		Args: usageArgs(cobra.ExactArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputPath == "" {
				return fmt.Errorf("%w: --output is required", errUsage)
			}
			if fileExists(outputPath) && !force {
				return fmt.Errorf("output exists at %s; use --force to overwrite", outputPath)
			}
			file, err := findMessageFile(cmd, rc, args[0], args[1])
			if err != nil {
				return err
			}
			data, err := rc.client.Download(cmd.Context(), file.URL)
			if err != nil {
				return err
			}
			if err := os.WriteFile(outputPath, data, 0600); err != nil {
				return fmt.Errorf("write attachment %s: %w", outputPath, err)
			}
			rc.out.Success("attachment downloaded")
			rc.out.Printf("%s\n", outputPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing file")
	return cmd
}

func findMessageFile(
	cmd *cobra.Command,
	rc *runtime,
	threadID, fileID string,
) (*api.FileInfo, error) {
	response, err := rc.client.GetMessageThread(cmd.Context(), threadID)
	if err != nil {
		return nil, err
	}
	for _, message := range response.MessageThreadViewModel.Messages {
		for _, file := range message.Files {
			if file.ID == fileID {
				if file.URL == "" {
					return nil, fmt.Errorf("message file %q has no download URL", fileID)
				}
				return &file, nil
			}
		}
	}
	return nil, fmt.Errorf("message file %q is not in thread %q", fileID, threadID)
}

func writeMessageTable(rc *runtime, threads []api.MessageThreadInfo) {
	rows := make([][]string, 0, len(threads))
	for _, thread := range threads {
		rows = append(rows, []string{
			thread.MessageThread.ID,
			strconv.FormatBool(!thread.MessageThreadUserAttributes.IsRead),
			thread.MessageThread.Subject,
			thread.TopMessage.Sender.FullName,
			thread.MessageThread.LastUpdatedOn,
		})
	}
	rc.out.Table([]string{"ID", "UNREAD", "SUBJECT", "SENDER", "UPDATED"}, rows)
}

func writeMessage(rc *runtime, info api.MessageInfo) {
	rc.out.Printf("\n%s  %s\n", info.Sender.FullName, info.Message.CreatedOn)
	rc.out.Printf("%s\n", cleanHTML(info.Message.Body))
	if len(info.Files) > 0 {
		rc.out.Printf("Attachments: %d\n", len(info.Files))
	}
}
