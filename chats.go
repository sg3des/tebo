package tebo

import (
	"strings"
	"sync"
)

type chat struct {
	ID       int
	Username string

	// lastMessageIsBot bool
	// editMessageID    int

	expectContext chan *Context
	expectCancel  chan bool

	fsm *FSM
}

type chats struct {
	chats sync.Map
	names sync.Map
}

func (c *chats) Get(msg Message) *chat {
	if msg.Chat.ID == 0 {
		log.Criticalf("message chat id is 0: %+v", msg)
	}
	if ch, ok := c.chats.Load(msg.Chat.ID); ok {
		return ch.(*chat)
	}

	ch := &chat{
		ID:       msg.Chat.ID,
		Username: msg.Chat.Username,
	}

	c.chats.Store(msg.Chat.ID, ch)
	c.names.Store(msg.Chat.Username, ch)

	return ch
}

func (c *chats) LookupByName(name string) (*chat, bool) {
	name = strings.TrimPrefix(name, "@")

	ch, ok := c.names.Load(name)
	if !ok {
		return nil, false
	}

	return ch.(*chat), true
}

// ExpectAnswer wait next message, intercept it if this message not a command
// return false if next message is command
func (c *chat) ExpectAnswer() (ctx *Context, ok bool) {
	c.expectContext = make(chan *Context)
	c.expectCancel = make(chan bool)

	select {
	case ctx = <-c.expectContext:
		ok = true
	case <-c.expectCancel:
		ok = false
	}

	return
}

func (c *chat) closeExpectChannels() {
	close(c.expectContext)
	close(c.expectCancel)

	c.expectContext = nil
	c.expectCancel = nil
}
