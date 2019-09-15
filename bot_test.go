package tebo

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/op/go-logging"
)

var (
	token    = ""
	chatid   = 278037961
	username = ""

	bot *Bot
	err error
)

func init() {
	// logging.SetFormatter(logging.MustStringFormatter(
	// 	`%{color}[%{module} %{shortfile}] %{message}%{color:reset}`,
	// ))
	logging.SetBackend(logging.NewLogBackend(os.Stderr, "", 0))

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

	PollInterval = 10 * time.Second

	go bot.Start()
}

func TestReplyKeyboard(t *testing.T) {
	if token == "" || chatid == 0 {
		t.SkipNow()
	}

	opt := SendOptions{
		ReplyMarkup: ReplyKeyboardMarkup{
			Keboard: [][]KeyboardButton{
				[]KeyboardButton{
					KeyboardButton{Text: "1"},
					KeyboardButton{Text: "2"},
					KeyboardButton{Text: "3"},
				},
				[]KeyboardButton{
					KeyboardButton{Text: "4"},
					KeyboardButton{Text: "5"},
					KeyboardButton{Text: "6"},
				},
				[]KeyboardButton{
					KeyboardButton{Text: "7"},
					KeyboardButton{Text: "8"},
					KeyboardButton{Text: "9"},
				},
			},
		},
	}

	if err := bot.SendMessage(chatid, "keyboard", opt); err != nil {
		t.Error(err)
	}
}

func TestInlineKeyboard(t *testing.T) {
	if token == "" || chatid == 0 {
		t.SkipNow()
	}

	opt := SendOptions{
		ReplyMarkup: InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{
				[]InlineKeyboardButton{
					InlineKeyboardButton{Text: "text", URL: "https://github.com/sg3des/tebo"},
					InlineKeyboardButton{
						Text: "login",
						LoginURL: &LoginURL{
							URL: "http://45.76.39.223",
						},
						CallbackData: "some callback data",
					},
				},
			},
		},
	}

	bot.UpdatesHandle(func(u *Update) bool {
		if u.CallbackQuery != nil {
			log.Notice("TestInlineKeyboard, callback data:", u.CallbackQuery.Data)
			return false
		}
		return true
	})

	if err := bot.SendMessage(chatid, "keyboard", opt); err != nil {
		t.Error(err)
	}
}

func TestRouting(t *testing.T) {
	if token == "" {
		t.SkipNow()
	}
	bot.Handle("/start", func(m Message) string {
		return fmt.Sprintf("Hello %s, ChatID: %d\nPassport: %+v", m.From.Username, m.Chat.ID, m.PassportData)
	})

	time.Sleep(30 * time.Second)
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
