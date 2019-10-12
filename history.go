package tebo

import (
	"bufio"
	"fmt"
	"os"

	"github.com/vmihailenco/msgpack"
)

func (b *Bot) addChat(u Update) {
	if u.CallbackQuery != nil {
		b.Chats.Get(u.CallbackQuery.Message.Chat)
	} else {
		b.Chats.Get(u.Message.Chat)
	}
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

		b.addChat(u)
	}

	return nil
}

func (b *Bot) updateHistory(updates []Update) error {
	for _, u := range updates {
		if u.UpdateID > b.UpdateID {
			b.UpdateID = u.UpdateID
		}

		b.addChat(u)

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
