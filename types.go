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

	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID int             `json:"message_id"`
	From      User            `json:"from,omitempty"`
	Chat      Chat            `json:"chat"`
	Date      int64           `json:"date"`
	Text      string          `json:"text"`
	Entities  []MessageEntity `json:"entities,omitempty"`

	// ...

	Document *Document   `json:"document,omitempty"`
	Photo    []PhotoSize `json:"photo,omitempty"`

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

type CallbackQuery struct {
	ID              string   `json:"id"`
	From            User     `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	ChatInstance    string   `json:"chat_instance"`
	Data            string   `json:"data,omitempty"`
	GameShortName   string   `json:"game_short_name,omitempty"`
}

//
// Files
//

type File struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
	FileSize int    `json:"file_size"`
}

type Document struct {
	FileID   string    `json:"file_id"`
	Thumb    PhotoSize `json:"thumb,omitempty"`
	FileName string    `json:"file_name,omitempty"`
	MIMEType string    `json:"mime_type,omitempty"`
	FileSize int       `json:"file_size,omitempty"`
}

type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size,omitempty"`
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
	FrontSide   *PassportFile  `json:"front_side,omitempty"`
	ReverseSide *PassportFile  `json:"reverse_side,omitempty"`
	Selfie      *PassportFile  `json:"selfie,omitempty"`
	Translation []PassportFile `json:"translation,omitempty"`
	Hash        []byte         `json:"hash"`
}

const (
	PassportElemType_PersonalDetails       = "personal_details"
	PassportElemType_Passport              = "passport"
	PassportElemType_DriverLicense         = "driver_license"
	PassportElemType_IdentityCard          = "identity_card"
	PassportElemType_InternalPassport      = "internal_passport"
	PassportElemType_Address               = "address"
	PassportElemType_UtilityBill           = "utility_bill"
	PassportElemType_BankStatement         = "bank_statement"
	PassportElemType_RentalAgreement       = "rental_agreement"
	PassportElemType_PassportRegistration  = "passport_registration"
	PassportElemType_TemporaryRegistration = "temporary_registration"
	PassportElemType_PhoneNumber           = "phone_number"
	PassportElemType_Email                 = "email"
)

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
