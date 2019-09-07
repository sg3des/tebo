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
