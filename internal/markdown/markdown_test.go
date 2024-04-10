package markdown

import (
	"strings"
	"testing"

	"github.com/rneatherway/gh-slack/internal/mocks"
	"github.com/rneatherway/gh-slack/internal/slackclient"
)

func TestFromMessagesCombinesAdjacentMessagesFromSameUser(t *testing.T) {
	mockClient := &mocks.MockClient{}
	mockClient.MockSuccessfulAuthResponse()
	client, err := slackclient.Null("test", mockClient)
	if err != nil {
		t.Fatal(err)
	}
	messages := []slackclient.Message{
		{Text: "hello", User: "82317", BotID: "", Ts: "123.456"},
		{Text: "second message", User: "82317", BotID: "", Ts: "124.567"},
	}
	history := &slackclient.HistoryResponse{Ok: true, HasMore: false, Messages: messages}
	mockClient.MockSuccessfulUsersResponse([]slackclient.User{{ID: "82317", Name: "cheshire137"}})
	actual, err := FromMessages(client, history)
	if err != nil {
		t.Fatal(err)
	}
	expected := `> **cheshire137** at 1970-01-01 00:02 UTC
>
> hello
>
> second message`
	if expected != strings.TrimSpace(actual) {
		t.Fatal("expected:\n\n", expected, "\n\ngot:\n\n", actual)
	}
}

func TestFromMessagesCombinesAdjacentMessagesFromSameBot(t *testing.T) {
	mockClient := &mocks.MockClient{}
	mockClient.MockSuccessfulAuthResponse()
	client, err := slackclient.Null("test", mockClient)
	if err != nil {
		t.Fatal(err)
	}
	messages := []slackclient.Message{
		{Text: "hello", User: "", BotID: "bot123", Ts: "123.456"},
		{Text: "second message", User: "", BotID: "bot123", Ts: "124.567"},
	}
	history := &slackclient.HistoryResponse{Ok: true, HasMore: false, Messages: messages}
	actual, err := FromMessages(client, history)
	if err != nil {
		t.Fatal(err)
	}
	expected := `> **bot bot123** at 1970-01-01 00:02 UTC
>
> hello
>
> second message`
	if expected != strings.TrimSpace(actual) {
		t.Fatal("expected:\n\n", expected, "\n\ngot:\n\n", actual)
	}
}

func TestFromMessagesSeparatesMessagesFromSameUserWhenFarApartInTime(t *testing.T) {
	mockClient := &mocks.MockClient{}
	mockClient.MockSuccessfulAuthResponse()
	client, err := slackclient.Null("test", mockClient)
	if err != nil {
		t.Fatal(err)
	}
	messages := []slackclient.Message{
		{Text: "hello", User: "82317", BotID: "", Ts: "1679058753.0"},
		{Text: "second message", User: "82317", BotID: "", Ts: "1679064168.0"},
	}
	history := &slackclient.HistoryResponse{Ok: true, HasMore: false, Messages: messages}
	mockClient.MockSuccessfulUsersResponse([]slackclient.User{{ID: "82317", Name: "cheshire137"}})
	actual, err := FromMessages(client, history)
	if err != nil {
		t.Fatal(err)
	}
	expected := `> **cheshire137** at 2023-03-17 13:12 UTC
>
> hello

> **cheshire137** at 2023-03-17 14:42 UTC
>
> second message`
	if expected != strings.TrimSpace(actual) {
		t.Fatal("expected:\n\n", expected, "\n\ngot:\n\n", actual)
	}
}

func TestFromMessagesSeparatesMessagesFromSameUserWhenNotAdjacent(t *testing.T) {
	mockClient := &mocks.MockClient{}
	mockClient.MockSuccessfulAuthResponse()
	client, err := slackclient.Null("test", mockClient)
	if err != nil {
		t.Fatal(err)
	}
	messages := []slackclient.Message{
		{Text: "first!", User: "82317", BotID: "", Ts: "123.456"},
		{Text: "Message the Second", User: "1234", BotID: "", Ts: "124.567"},
		{Text: "third message", User: "82317", BotID: "", Ts: "125.678"},
	}
	history := &slackclient.HistoryResponse{Ok: true, HasMore: false, Messages: messages}
	mockClient.MockSuccessfulUsersResponse([]slackclient.User{
		{ID: "82317", Name: "cheshire137"},
		{ID: "1234", Name: "octokatherine"},
	})
	actual, err := FromMessages(client, history)
	if err != nil {
		t.Fatal(err)
	}
	expected := `> **cheshire137** at 1970-01-01 00:02 UTC
>
> first!

> **octokatherine** at 1970-01-01 00:02 UTC
>
> Message the Second

> **cheshire137** at 1970-01-01 00:02 UTC
>
> third message`
	if expected != strings.TrimSpace(actual) {
		t.Fatal("expected:\n\n", expected, "\n\ngot:\n\n", actual)
	}
}
