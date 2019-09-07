package tebo

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	token    = ""
	chatid   = 278037961
	username = ""

	bot *Bot
	err error
)

func init() {
	token = os.Getenv("TEST_TOKEN")
	username = os.Getenv("TEST_USERNAME")
	if _chatid := os.Getenv("TEST_CHATID"); _chatid != "" {
		chatid, _ = strconv.Atoi(_chatid)
	}

	if token == "" || chatid == 0 {
		log.Warning("SKIP tebo tests, env variable `TEST_TOKEN` or `TEST_CHATID` not specified")
		return
	}

	bot, err = NewBot(token, "testdata/history")
	if err != nil {
		log.Fatal(err)
	}
}

func TestRouting(t *testing.T) {
	if token == "" {
		t.SkipNow()
	}
	bot.Handle("/start", func(m Message) string {
		return fmt.Sprintf("Hello %s, ChatID: %d", m.From.Username, m.Chat.ID)
	})

	time.Sleep(15 * time.Second)
}

func TestGetUpdates(t *testing.T) {
	if token == "" {
		t.SkipNow()
	}

	updates, err := bot.GetUpdates(0)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, u := range updates {
		t.Logf("%+v", u.Message)
		chatid = u.Message.Chat.ID
		username = u.Message.Chat.Username
	}
}

func TestLookupChatID(t *testing.T) {
	if token == "" || username == "" {
		t.SkipNow()
	}

	id, ok := bot.LookupChatID(username)
	if !ok {
		t.Error("chat not found")
	} else if id != chatid {
		t.Error("invalid chatid")
	}

	t.Log(id)
}

func TestSendMessage(t *testing.T) {
	if token == "" || chatid == 0 {
		t.SkipNow()
	}

	if err := bot.SendMessage(chatid, "Hello"); err != nil {
		t.Error(err)
	}
}

func TestSendPhoto(t *testing.T) {
	if token == "" || chatid == 0 {
		t.SkipNow()
	}

	f, err := os.Open("testdata/gopher.png")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if err := bot.SendPhoto(chatid, f, "Image"); err != nil {
		t.Error(err)
	}
}
