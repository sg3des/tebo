package tebo

import (
	"bufio"
	"fmt"
	"os"

	"github.com/vmihailenco/msgpack"
)

func (b *Bot) addChat(chat Chat) {
	for _, c := range b.Chats {
		if c.ID == chat.ID {
			return
		}
	}

	b.Chats = append(b.Chats, chat)
}

func (b *Bot) readHistory(filename string) (err error) {
	b.historyFile, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(b.historyFile)
	for s.Scan() {
		var u Update
		if err := msgpack.Unmarshal(s.Bytes(), &u); err != nil {
			// log.Error(err)
			continue
		}

		if u.UpdateID > b.UpdateID {
			b.UpdateID = u.UpdateID
		}

		b.addChat(u.Message.Chat)
	}

	return nil
}

func (b *Bot) updateHistory(updates []Update) error {
	for _, u := range updates {
		if u.UpdateID > b.UpdateID {
			b.UpdateID = u.UpdateID
		}

		b.addChat(u.Message.Chat)

		line, err := msgpack.Marshal(u)
		if err != nil {
			return fmt.Errorf("encode message failed: %v", err)
		}

		b.historyFile.Write(line)
		_, err = b.historyFile.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("failed write to history: %v", err)
		}
	}

	return nil
}

func (b *Bot) loadUpdates() (updates []Update, err error) {
	updates, err = b.GetUpdates(b.UpdateID + 1)
	if err != nil {
		return
	}

	err = b.updateHistory(updates)
	return
}
