package tebo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/schema"
	"github.com/imroc/req"
	"github.com/op/go-logging"
)

var (
	addr     = "https://api.telegram.org/bot%s/"
	fileaddr = "https://api.telegram.org/file/bot%s/"
	log      = logging.MustGetLogger("TEBO")
)

type Bot struct {
	User

	addr     string
	fileaddr string

	// Chats    []Chat
	UpdateID int

	historyFile *os.File

	handlers        []handler
	middlewares     []MiddlewareFunc
	updatesHandlers []UpdatesFunc

	Chats *chats
	fsm   []*FSM

	closed bool

	// expectContext chan *Context
	// expectCancel  chan bool
}

func NewBot(token, historyfile string) (b *Bot, err error) {
	b = &Bot{
		addr:     fmt.Sprintf(addr, token),
		fileaddr: fmt.Sprintf(fileaddr, token),
		Chats:    new(chats),
	}

	b.User, err = b.GetMe()
	if err != nil {
		return b, fmt.Errorf("connection failed: %v", err)
	}

	log = logging.MustGetLogger("TEBO:" + b.Username)

	if err = b.readHistory(historyfile); err != nil {
		return b, fmt.Errorf("history initialize failed: %v", err)
	}

	return
}

func (b *Bot) Close() {
	if b == nil {
		return
	}

	b.closed = true
}

func (b *Bot) LookupChatID(name string) (int, bool) {
	ch, ok := b.Chats.LookupByName(name)
	if !ok {
		return 0, false
	}

	return ch.ID, true
}

type ErrorResponse struct {
	Status string `json:"-"`

	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func (e *ErrorResponse) Error() string {
	return e.Description
}

func (b *Bot) Request(method string, payload, v interface{}) error {
	resp, err := req.Post(b.addr+method, req.BodyJSON(payload))
	if err != nil {
		return err
	}
	// log.Debug(resp.Dump())

	if r := resp.Response(); r.StatusCode >= 400 {
		err := &ErrorResponse{Status: r.Status}
		resp.ToJSON(err)
		return err
	}

	if v == nil {
		return nil
	}

	var r Response
	if err := resp.ToJSON(&r); err != nil {
		return err
	}

	return json.Unmarshal(r.Result, &v)
}

type FormFile struct {
	field string

	Name string
	io.Reader
}

func (b *Bot) FileRequest(method string, file FormFile, payload interface{}, v interface{}) error {
	q := make(url.Values)
	e := schema.NewEncoder()
	e.SetAliasTag("json")
	e.Encode(payload, q)

	resp, err := req.Post(b.addr+method, req.FileUpload{
		FieldName: file.field,
		File:      ioutil.NopCloser(file),
		FileName:  file.Name,
	}, q)
	// log.Debug(resp.Dump())
	if err != nil {
		return err
	}

	var r Response
	if err := resp.ToJSON(&r); err != nil || !r.OK {
		return errors.New(resp.String())
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
	ChatID      int `json:"chat_id"`
	SendMessage `structs:",flatten`
}

type SendMessage struct {
	Text        string `json:"text"`
	SendOptions `json:",omitempty" structs:",flatten`
}

const (
	ParseModeHTML     = "HTML"
	ParseModeMarkdown = "Markdown"
)

type SendOptions struct {
	ParseMode string `json:"parse_mode,omitempty" structs:"parse_mode,omitempty"`
	// disable_web_page_preview
	DisableNotification bool `json:"disable_notification,omitempty"  structs:"disable_notification,omitempty"`
	// reply_to_message_id
	ReplyMarkup interface{} `json:"reply_markup,omitempty"  structs:"reply_markup,omitempty"`
}

// type ReplyMarkup struct {
// 	InlineKeyboardMarkup
// 	ReplyKeyboardMarkup
// 	ReplyKeyboardRemove
// 	ForceReply
// }

func NewMessage(text string, opt ...SendOptions) *SendMessage {
	msg := &SendMessage{Text: text}
	if len(opt) > 0 {
		msg.SendOptions = opt[0]
	}

	return msg
}

func (b *Bot) SendMessage(chatid int, smsg *SendMessage) (msgid int, err error) {
	if len(smsg.Text) == 0 {
		return
	}

	var msg Message
	err = b.Request("sendMessage", ReqSendMessage{ChatID: chatid, SendMessage: *smsg}, &msg)
	return msg.MessageID, err
}

func (b *Bot) SendTextMessage(chatid int, text string, a ...interface{}) (msgid int, err error) {
	if len(text) == 0 {
		return 0, errors.New("text is empty")
	}

	var msg Message
	err = b.Request("sendMessage", ReqSendMessage{
		ChatID: chatid,
		SendMessage: SendMessage{
			Text: fmt.Sprintf(text, a...),
		},
	}, &msg)

	return msg.MessageID, err
}

func (b *Bot) DeleteMessage(chatid int, messageid int) error {
	return b.Request("deleteMessage", map[string]interface{}{
		"chat_id":    chatid,
		"message_id": messageid,
	}, nil)
}

// type SendOptions struct {
// 	Caption             string `json:"caption,omitempty" structs:"caption"`
// 	DisableNotification bool   `json:"disable_notification,omitempty" structs:"disable_notification"`
// 	ParseMode           string `json:"parse_mode,omitempty" structs:"disable_notification"`
// }

type ReqSendPhoto struct {
	ChatID      int    `json:"chat_id" structs:"chat_id"`
	Caption     string `json:"caption,omitempty" structs:"caption,omitempty"`
	SendOptions `json:",omitempty" structs:",flatten"`
	// ...
}

func (b *Bot) SendPhoto(chatid int, photo FormFile, caption string, opt ...SendOptions) (msgid int, err error) {
	req := ReqSendPhoto{
		ChatID:  chatid,
		Caption: caption,
	}
	if len(opt) > 0 {
		req.SendOptions = opt[0]
	}
	photo.field = "photo"

	var msg Message
	err = b.FileRequest("sendPhoto", photo, req, &msg)
	return msg.MessageID, err
}

func (b *Bot) SendDocument(chatid int, document FormFile, caption string, opt ...SendOptions) (msgid int, err error) {
	req := ReqSendPhoto{
		ChatID:  chatid,
		Caption: caption,
	}
	if len(opt) > 0 {
		req.SendOptions = opt[0]
	}
	document.field = "document"

	var msg Message
	err = b.FileRequest("sendDocument", document, req, nil)
	return msg.MessageID, err
}

//
// EditMessage
//

type ReqEditMessage struct {
	ChatID    int `json:"chat_id"`
	MessageID int `json:"message_id,omitempty"`
	SendMessage
}

func (b *Bot) EditMessage(chatid int, messageid int, smsg *SendMessage) (msg Message, err error) {
	if len(smsg.Text) == 0 {
		return msg, errors.New("text is empty")
	}

	reqmsg := ReqEditMessage{ChatID: chatid, MessageID: messageid, SendMessage: *smsg}
	err = b.Request("editMessageText", reqmsg, &msg)
	return
}

type ReqEditMessageMedia struct {
	ChatID      int    `json:"chat_id"`
	MessageID   int    `json:"message_id"`
	Media       string `json:"media"`
	ReplyMarkup string `json:"reply_markup,omitempty"`
}

func (b *Bot) EditMessageMedia(chatid, msgid int, file FormFile, caption string, keys ...*InlineKeyboardMarkup) (msg Message, err error) {
	file.field = "photo"

	payload := ReqEditMessageMedia{
		ChatID:    chatid,
		MessageID: msgid,
		Media: InputMedia{
			Caption: caption,
			Type:    "photo",
			Media:   "attach://" + file.field,
		}.ToJSON(),
	}

	if len(keys) > 0 {
		payload.ReplyMarkup = keys[0].ToJSON()
	}

	err = b.FileRequest("editMessageMedia", file, payload, &msg)
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

	var respdata struct {
		OK     bool
		Result File
	}

	err = json.NewDecoder(resp.Body).Decode(&respdata)
	return respdata.Result, err
}

func (b *Bot) DownloadFile(filepath string, w io.Writer) error {
	resp, err := http.Get(b.fileaddr + filepath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}

func (b *Bot) LoadFile(fileid string, w io.Writer) (File, error) {
	f, err := b.GetFile(fileid)
	if err != nil {
		return f, err
	}

	return f, b.DownloadFile(f.FilePath, w)
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

func (k *InlineKeyboardConstuctor) Add(btn InlineKeyboardButton) {
	k.buttons = append(k.buttons, btn)
}

func (k *InlineKeyboardConstuctor) ToReplyMarkup() *InlineKeyboardMarkup {
	var keyboard [][]InlineKeyboardButton

	for i := 0; i < len(k.buttons); i += k.columns {
		line := make([]InlineKeyboardButton, 0, k.columns)
		for j := i; j < i+k.columns && j < len(k.buttons); j++ {
			line = append(line, k.buttons[j])
		}
		keyboard = append(keyboard, line)
	}

	return &InlineKeyboardMarkup{keyboard}
}

func (k *InlineKeyboardConstuctor) GetButtonText(data string) (string, bool) {
	for _, b := range k.buttons {
		if b.CallbackData == data {
			return b.Text, true
		}
	}

	return "", false
}
