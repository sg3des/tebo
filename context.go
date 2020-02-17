package tebo

import (
	"fmt"
	"sync"
)

// Context argument for handlers
type Context struct {
	Bot *Bot

	Update
	Message

	chat *chat

	sync.Map
}

func (b *Bot) newContext(u Update) *Context {
	ctx := &Context{
		Bot:     b,
		Update:  u,
		Message: u.Message,
	}

	if u.CallbackQuery != nil {
		ctx.Message = u.CallbackQuery.Message
	}

	ctx.chat = b.Chats.Get(ctx.Message.Chat)

	if u.CallbackQuery != nil {
		if u.CallbackQuery.Data != "" {
			if fsm, ok := b.lookupFSM(u.CallbackQuery.Data); ok {
				ctx.chat.fsm = fsm.root
			}
		}

		// if u.CallbackQuery.From.IsBot {
		ctx.chat.setEditMessageID(u.CallbackQuery.Message.MessageID)
		// }
	}

	return ctx
}

func (b *Bot) ContextFromMessage(msg Message) *Context {
	ctx := &Context{
		Bot:     b,
		Message: msg,
		chat:    b.Chats.Get(msg.Chat),
	}

	return ctx
}

// Send prepared message to the current chat
func (ctx *Context) Send(smsg *SendMessage) (int, error) {
	msgid, err := ctx.Bot.SendMessage(ctx.Chat.ID, smsg)
	if err != nil {
		return msgid, err
	}

	ctx.chat.setEditMessageID(msgid)

	return msgid, err
}

// Edit message of the current chat
func (ctx *Context) Edit(messageid int, smsg *SendMessage) error {
	_, err := ctx.Bot.EditMessage(ctx.Chat.ID, messageid, smsg)
	return err
}

// SendMessage with text and optional Options, as ParseMode or Keyboard
func (ctx *Context) SendMessage(text string, opt ...SendOptions) (int, error) {
	return ctx.Send(ctx.NewMessage(text, opt...))
}

// SendTextMessage with formating text
func (ctx *Context) SendTextMessage(text string, a ...interface{}) (int, error) {
	return ctx.Send(ctx.NewMessage(fmt.Sprintf(text, a...)))
}

func (ctx *Context) EditMessage(messageid int, text string, opt ...SendOptions) error {
	_, err := ctx.Bot.EditMessage(ctx.Chat.ID, messageid, ctx.NewMessage(text, opt...))
	return err
}

// Expect answer of this user
func (ctx *Context) ExpectAnswer() (*Context, bool) {
	return ctx.chat.ExpectAnswer()
}

func (ctx *Context) NewMessage(text string, opt ...SendOptions) *SendMessage {
	return NewMessage(text, opt...)
}

func (ctx *Context) NewTextMessage(text string, a ...interface{}) *SendMessage {
	return NewMessage(fmt.Sprintf(text, a...))
}

func (ctx *Context) EditOrSendMessage(text string, opt ...SendOptions) (int, error) {
	if ctx.chat.lastMessageIsBot {
		err := ctx.EditMessage(ctx.chat.editMessageID, text, opt...)
		return ctx.chat.editMessageID, err
	}

	msgid, err := ctx.SendMessage(text, opt...)
	if err != nil {
		return msgid, err
	}

	ctx.chat.setEditMessageID(msgid)

	return msgid, err
}

func (ctx *Context) EditOrSend(smsg *SendMessage) (int, error) {
	if ctx.chat.lastMessageIsBot {
		err := ctx.Edit(ctx.chat.editMessageID, smsg)
		return ctx.chat.editMessageID, err
	}

	msgid, err := ctx.Send(smsg)
	if err != nil {
		return msgid, err
	}

	ctx.chat.setEditMessageID(msgid)

	return msgid, err
}
