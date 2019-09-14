package tebo

import "encoding/json"

type Response struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result"`
}

type User struct {
	ID           int    `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int             `json:"message_id"`
	From      User            `json:"from,omitempty"`
	Chat      Chat            `json:"chat"`
	Date      int64           `json:"date"`
	Text      string          `json:"text"`
	Entities  []MessageEntity `json:"entities,omitempty"`

	// ...

	ConnectedWebsite string                `json:"connected_website,omitempty"`
	PassportData     *PassportData         `json:"passport_data,omitempty"`
	ReplyMarkup      *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

func (m Message) BotCommand() (string, bool) {
	for _, e := range m.Entities {
		if e.Type == "bot_command" {
			return m.Text, true
		}
	}

	return "", false
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
	User   User   `json:"user,omitempty"`
}

type Chat struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	// ...
}

//
//
//

type PassportData struct {
	Data        []EncryptedPassportElement `json:"data"`
	Credentials EncryptedCredentials       `json:"credentials"`
}

type EncryptedPassportElement struct {
	Type        string         `json:"type"`
	Data        []byte         `json:"data,omitempty"`
	PhoneNumber string         `json:"phone_number,omitempty"`
	Email       string         `json:"email,omitempty"`
	Files       []PassportFile `json:"files,omitempty"`
	FrontSide   PassportFile   `json:"front_side,omitempty"`
	ReverseSide PassportFile   `json:"reverse_side,omitempty"`
	Selfie      PassportFile   `json:"selfie,omitempty"`
	Translation []PassportFile `json:"translation,omitempty"`
	Hash        []byte         `json:"hash"`
}

type PassportFile struct {
	FileID   string `json:"file_id"`
	FileSize int64  `json:"file_size"`
	FileDate int64  `json:"file_date"`
}

type EncryptedCredentials struct {
	Data   []byte `json:"data"`
	Hash   []byte `json:"hash"`
	Secret []byte `json:"secret"`
}

//
// InlineKeyboardMarkup
//

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string    `json:"text"`
	URL          string    `json:"url,omitempty"`
	LoginURL     *LoginURL `json:"login_url,omitempty"`
	CallbackData string    `json:"callback_data,omitempty"`
	// switch_inline_query
	// switch_inline_query_current_chat
	// callback_game
	// pay
}

type LoginURL struct {
	URL                string `json:"url"`
	ForwardText        string `json:"forward_text,omitempty"`
	BotUsername        string `json:"bot_username,omitempty"`
	RequestWriteAccess bool   `json:"request_write_access,omitempty"`
}

//
// ReplyKeyboardMarkup
//

type ReplyKeyboardMarkup struct {
	Keboard         [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard,omitempty"`
	OneTimeKeyboard bool               `json:"one_time_keyboard,omitempty"`
	Selective       bool               `json:"selective,omitempty"`
}

type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact,omitempty"`
	RequestLocation bool   `json:"request_contact,omitempty"`
}

type ReplyKeyboardRemove struct {
	RemoveKeyboard bool `json:"remove_keyboard"`
	Selective      bool `json:"selective,omitempty"`
}

type ForceReply struct {
	ForceReply bool `json:"force_reply"`
	Selective  bool `json:"selective,omitempty"`
}
