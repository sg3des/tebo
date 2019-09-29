package tebo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/structs"
	"github.com/op/go-logging"
)

var (
	addr     = "https://api.telegram.org/bot%s/"
	fileaddr = "https://api.telegram.org/file/bot%s/"
	log      = logging.MustGetLogger("TELEGRAM_BOT")
)

type Bot struct {
	User

	addr     string
	fileaddr string

	Chats    []Chat
	UpdateID int

	historyFile *os.File

	handlers        []handler
	middlewares     []MiddlewareFunc
	updatesHandlers []UpdatesHandleFunc

	expectUpdate chan Update
	expectCancel chan bool
}

func NewBot(token, historyfile string) (b *Bot, err error) {
	b = &Bot{
		addr:     fmt.Sprintf(addr, token),
		fileaddr: fmt.Sprintf(fileaddr, token),
	}

	b.User, err = b.GetMe()
	if err != nil {
		return b, fmt.Errorf("connection failed: %v", err)
	}

	if err = b.readHistory(historyfile); err != nil {
		return b, fmt.Errorf("history initialize failed: %v", err)
	}

	return
}

func (b *Bot) LookupChatID(name string) (int, bool) {
	name = strings.TrimPrefix(name, "@")

	for _, chat := range b.Chats {
		if chat.Username == name || chat.Title == name {
			return chat.ID, true
		}
	}

	return 0, false
}

type ErrorResponse struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func (e *ErrorResponse) Error() string {
	return e.Description
}

func (b *Bot) Request(method string, payload, v interface{}) error {
	var body bytes.Buffer
	if payload != nil {
		err := json.NewEncoder(&body).Encode(payload)
		if err != nil {
			return fmt.Errorf("payload encode failed: %v", err)
		}
	}

	resp, err := http.Post(b.addr+method, "application/json", &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return handleResponseError(resp)
	}

	// respBody, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(respBody))

	var r Response
	// if err := json.Unmarshal(respBody, &r); err != nil {
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("response decode failed: %v", err)
	}

	return json.Unmarshal(r.Result, &v)
}

func handleResponseError(resp *http.Response) error {
	data, _ := ioutil.ReadAll(resp.Body)

	var err ErrorResponse
	json.Unmarshal(data, &err)
	if err.Description != "" {
		return &err
	}

	return fmt.Errorf("%s: %s", resp.Status, string(data))
}

func (b *Bot) FileRequest(method string, file io.Reader, payload interface{}, v interface{}) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	part, _ := w.CreateFormFile("photo", "image.png")
	io.Copy(part, file)

	if payload != nil {
		for key, val := range structs.Map(payload) {
			w.WriteField(key, fmt.Sprintf("%v", val))
		}
	}
	w.Close()

	resp, err := http.Post(b.addr+method, w.FormDataContentType(), &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		log.Error(string(data))

		return errors.New(resp.Status)
	}

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("response decode failed: %v", err)
	}

	return json.Unmarshal(r.Result, &v)
}

func (b *Bot) GetMe() (me User, err error) {
	err = b.Request("getMe", nil, &me)
	return
}

type ReqUpdates struct {
	Offset int `json:"offset"`
}

func (b *Bot) GetUpdates(offset int) (updates []Update, err error) {
	err = b.Request("getUpdates", ReqUpdates{Offset: offset}, &updates)
	return
}

type ReqSendMessage struct {
	ChatID      int    `json:"chat_id"`
	Text        string `json:"text"`
	SendOptions `json:",omitempty"`
}

const (
	ParseModeHTML     = "HTML"
	ParseModeMarkdown = "Markdown"
)

type SendOptions struct {
	ParseMode string `json:"parse_mode,omitempty"`
	// disable_web_page_preview
	// disable_notification
	// reply_to_message_id
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// type ReplyMarkup struct {
// 	InlineKeyboardMarkup
// 	ReplyKeyboardMarkup
// 	ReplyKeyboardRemove
// 	ForceReply
// }

func (b *Bot) SendMessage(chatid int, text string, opt ...SendOptions) (msg Message, err error) {
	if len(text) == 0 {
		return
	}

	reqmsg := ReqSendMessage{ChatID: chatid, Text: text}
	if len(opt) > 0 {
		reqmsg.SendOptions = opt[0]
	}

	err = b.Request("sendMessage", reqmsg, &msg)
	return
}

type ReqSendPhoto struct {
	ChatID  int    `json:"chat_id" structs:"chat_id"`
	Caption string `json:"caption,omitempty" structs:"caption"`
	// ...
}

func (b *Bot) SendPhoto(chatid int, photo io.Reader, caption string) error {
	return b.FileRequest("sendPhoto", photo, ReqSendPhoto{ChatID: chatid, Caption: caption}, nil)
}

//
// EditMessage
//

type ReqEditMessage struct {
	ChatID      int    `json:"chat_id"`
	MessageID   int    `json:"message_id,omitempty"`
	Text        string `json:"text"`
	SendOptions `json:",omitempty"`
}

func (b *Bot) EditMessage(chatid int, messageid int, text string, opt ...SendOptions) (msg Message, err error) {
	if len(text) == 0 {
		return msg, errors.New("text is empty")
	}

	reqmsg := ReqEditMessage{ChatID: chatid, MessageID: messageid, Text: text}
	if len(opt) > 0 {
		reqmsg.SendOptions = opt[0]
	}

	err = b.Request("editMessageText", reqmsg, &msg)
	return
}

//
// GetFile
//

func (b *Bot) GetFile(fileid string) (f File, err error) {
	resp, err := http.Get(b.addr + "getFile?file_id=" + fileid)
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// TODO: read error from response body
		return f, errors.New(resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&f)
	return
}

func (b *Bot) DownloadFile(filepath string, w io.Writer) error {
	resp, err := http.Get(b.fileaddr + filepath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(w, resp.Body)
	return err
}

func (b *Bot) LoadFile(fileid string, w io.Writer) error {
	f, err := b.GetFile(fileid)
	if err != nil {
		return err
	}

	return b.DownloadFile(f.FilePath, w)
}

//
// Keyboards
//

type InlineKeyboardConstuctor struct {
	columns int

	buttons []InlineKeyboardButton
}

func NewInlineKeyboard(columns int) *InlineKeyboardConstuctor {
	return &InlineKeyboardConstuctor{columns: columns}
}

func (k *InlineKeyboardConstuctor) AddButton(text, data string) {
	k.buttons = append(k.buttons, InlineKeyboardButton{
		Text:         text,
		CallbackData: data,
	})
}

func (k *InlineKeyboardConstuctor) ToReplyMarkup() InlineKeyboardMarkup {
	var keyboard [][]InlineKeyboardButton

	for i := 0; i < len(k.buttons); i += k.columns {
		line := make([]InlineKeyboardButton, 0, k.columns)
		for j := i; j < i+k.columns && j < len(k.buttons); j++ {
			line = append(line, k.buttons[j])
		}
		keyboard = append(keyboard, line)
	}

	return InlineKeyboardMarkup{keyboard}
}
