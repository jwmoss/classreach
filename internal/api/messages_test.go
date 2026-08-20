package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(messageListTestHandler))
	defer server.Close()

	client := New(server.URL)
	client.antiForgeryToken = "token"
	messages, err := client.ListMessages(context.Background(), "Inbox", "", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages.MessageThreads) != 1 ||
		messages.MessageThreads[0].MessageThread.Subject != "Notice" {
		t.Fatalf("messages = %#v", messages)
	}
}

func messageListTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("__RequestVerificationToken") != "token" {
		http.Error(w, "unexpected request", http.StatusBadRequest)
		return
	}
	var body messageListRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.Label != "Inbox" || body.Page != 2 {
		http.Error(w, "unexpected body", http.StatusBadRequest)
		return
	}
	_, _ = fmt.Fprint(w, `{
		"MessageThreads":[{"MessageThread":{"ID":"thread-1","Subject":"Notice"},
		"MessageThreadUserAttributes":{"IsRead":false}}],
		"PagingInfo":{"CurrentPage":2,"TotalPages":3}
	}`)
}

func TestGetMessageThread(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{
			"MessageThreadViewModel":{"MessageThread":{"ID":"thread-1","Subject":"Notice"},
			"Messages":[{"Message":{"Body":"Hello"},"Sender":{"FullName":"Teacher"}}]},
			"UnreadMessagesCount":0
		}`)
	}))
	defer server.Close()

	client := New(server.URL)
	thread, err := client.GetMessageThread(context.Background(), "thread-1")
	if err != nil {
		t.Fatal(err)
	}
	if thread.MessageThreadViewModel.Messages[0].Message.Body != "Hello" {
		t.Fatalf("thread = %#v", thread)
	}
}
