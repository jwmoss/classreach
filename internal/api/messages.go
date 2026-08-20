package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type MessageList struct {
	MessageThreads []MessageThreadInfo `json:"MessageThreads"`
	PagingInfo     PagingInfo          `json:"PagingInfo"`
}

type PagingInfo struct {
	CurrentPage  int `json:"CurrentPage"`
	ItemsPerPage int `json:"ItemsPerPage"`
	TotalItems   int `json:"TotalItems"`
	TotalPages   int `json:"TotalPages"`
}

type MessageThreadInfo struct {
	MessageCount                int                 `json:"MessageCount"`
	MessageThread               MessageThread       `json:"MessageThread"`
	MessageThreadUserAttributes MessageThreadStatus `json:"MessageThreadUserAttributes"`
	SenderNameHistory           string              `json:"SenderNameHistory"`
	ShowThreadURL               string              `json:"ShowThreadURL"`
	TopMessage                  MessageInfo         `json:"TopMessage"`
}

type MessageThread struct {
	CreatedOn     string  `json:"CreatedOn"`
	ID            string  `json:"ID"`
	LastUpdatedOn string  `json:"LastUpdatedOn"`
	SectionID     *string `json:"Section_ID"`
	Subject       string  `json:"Subject"`
}

type MessageThreadStatus struct {
	IsArchived      bool   `json:"IsArchived"`
	IsRead          bool   `json:"IsRead"`
	MessageThreadID string `json:"MessageThreadID"`
}

type MessageInfo struct {
	Files      []FileInfo      `json:"Files"`
	Message    Message         `json:"Message"`
	Recipients []RecipientInfo `json:"Recipients"`
	Sender     MessageUser     `json:"Sender"`
}

type Message struct {
	Body            string  `json:"Body"`
	CreatedOn       string  `json:"CreatedOn"`
	ID              string  `json:"ID"`
	IsDraft         bool    `json:"IsDraft"`
	MessageThreadID string  `json:"MessageThread_ID"`
	ParentMessageID *string `json:"ParentMessage_ID"`
	SenderID        string  `json:"Sender_ID"`
}

type MessageUser struct {
	Email    string `json:"Email"`
	FullName string `json:"FullName"`
	ID       string `json:"ID"`
}

type RecipientInfo struct {
	Recipient MessageRecipient `json:"Recipient"`
}

type MessageRecipient struct {
	Name          string `json:"Name"`
	RecipientType string `json:"RecipientType"`
	UserID        string `json:"User_ID"`
}

type FileInfo struct {
	DownloadLink string `json:"FileDownloadLink"`
	ID           string `json:"ID"`
	Name         string `json:"Name"`
	URL          string `json:"Url"`
}

func (f FileInfo) DownloadURL() string {
	if f.URL != "" {
		return f.URL
	}
	return f.DownloadLink
}

type MessageThreadDetail struct {
	Labels        []any         `json:"Labels"`
	MessageThread MessageThread `json:"MessageThread"`
	Messages      []MessageInfo `json:"Messages"`
}

type ThreadMessagesResponse struct {
	MessageThreadViewModel MessageThreadDetail `json:"MessageThreadViewModel"`
	UnreadMessagesCount    int                 `json:"UnreadMessagesCount"`
}

type messageListRequest struct {
	Label           string  `json:"Label"`
	MessageThreadID string  `json:"MessageThreadID"`
	Page            int     `json:"Page,omitempty"`
	ProxyUserID     *string `json:"ProxyUserID"`
	SearchTerm      string  `json:"SearchTerm"`
}

func (c *Client) ListMessages(
	ctx context.Context,
	label, searchTerm string,
	page int,
) (*MessageList, error) {
	if strings.TrimSpace(label) == "" {
		return nil, fmt.Errorf("message label is required")
	}
	body := messageListRequest{
		Label: label, Page: page, ProxyUserID: nil, SearchTerm: searchTerm,
	}
	var messages MessageList
	request := JSONRequest{
		Body: body, Method: http.MethodPost, Path: "/Messages/GetMessageThreads",
	}
	err := c.DoJSON(ctx, request, &messages)
	if err != nil {
		return nil, err
	}
	return &messages, nil
}

func (c *Client) GetMessageThread(
	ctx context.Context,
	id string,
) (*ThreadMessagesResponse, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("message thread ID is required")
	}
	body := map[string]any{"messageThreadID": id, "userID": nil}
	var response ThreadMessagesResponse
	request := JSONRequest{
		Body: body, Method: http.MethodPost, Path: "/Messages/GetThreadMessages",
	}
	err := c.DoJSON(ctx, request, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
