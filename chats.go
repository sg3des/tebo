package tebo

import (
	"strings"
	"sync"
)

type chat struct {
	ID       int
	Username string

	lastMessageIsBot bool
	editMessageID    int

	expectContext chan *Context
	expectCancel  chan bool

	fsm *FSM
}

type chats struct {
	chats sync.Map
	names sync.Map
}

func (c *chats) Get(tc Chat) *chat {
	if tc.ID == 0 {
		log.Criticalf("message chat id is 0: %+v", tc)
	}
	if ch, ok := c.chats.Load(tc.ID); ok {
		return ch.(*chat)
	}

	username := tc.Username
	if tc.Username == "" && tc.Title != "" {
		username = tc.Title
	}

	ch := &chat{
		ID:       tc.ID,
		Username: username,
	}

	c.chats.Store(tc.ID, ch)
	c.names.Store(tc.Username, ch)

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

func (c *chats) Range(f func(id int, username string) bool) {
	c.chats.Range(func(_, val interface{}) bool {
		ch := val.(*chat)
		return f(ch.ID, ch.Username)
	})
}

// ExpectAnswer wait next message, intercept it if this message not a command
// return false if next message is command
func (c *chat) ExpectAnswer() (ctx *Context, ok bool) {
	c.lastMessageIsBot = false

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

func (c *chat) setEditMessageID(msgid int) {
	c.editMessageID = msgid
	c.lastMessageIsBot = true
}
