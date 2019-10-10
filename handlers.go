package tebo

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"time"
)

var PollInterval = 30 * time.Second

type HandleFunc func(*Context) *SendMessage

type handler struct {
	cmd         string
	exp         *regexp.Regexp
	callback    HandleFunc
	middlewares []MiddlewareFunc
}

// Handle command with specified function
func (b *Bot) Handle(cmd string, f HandleFunc, mid ...MiddlewareFunc) error {
	exp, err := regexp.Compile("^" + cmd + "$")
	if err != nil {
		return err
	}

	b.handlers = append(b.handlers, handler{
		cmd:         cmd,
		exp:         exp,
		callback:    f,
		middlewares: mid,
	})

	return nil
}

func (b *Bot) Start() {
	for {
		updates, err := b.loadUpdates()
		if err != nil {
			log.Error(err)
		} else {
			// try to lookup apporpriate handler for this update
			for _, u := range updates {
				go b.route(u)
			}
		}

		time.Sleep(PollInterval)
	}
}

func (b *Bot) route(u Update) {
	ctx := b.newContext(u)

	// pass context for each updates handler, if it return false then stop further process
	for _, h := range b.updatesHandlers {
		if pass := h(ctx); !pass {
			return
		}
	}

	// if expect enable
	if ctx.chat.expectContext != nil && ctx.chat.expectCancel != nil {
		// if message is command, started from '/', then cancel expect channels
		// and lookup appropriate handler for this command
		if len(u.Message.Text) > 0 && u.Message.Text[0] == '/' {
			ctx.chat.expectCancel <- true
			ctx.chat.closeExpectChannels()
		} else {

			// if message is not command then send this message to expect channel
			// and then do not lookup handler for this message
			ctx.chat.expectContext <- ctx
			ctx.chat.closeExpectChannels()
			return
		}
	}

	// if for chat enable FSM and its not a bot command, then pass context to it
	if ctx.chat.fsm != nil {
		if _, ok := u.Message.BotCommand(); !ok {
			if err := ctx.chat.fsm.handle(ctx); err != nil {
				log.Error("fsm error:", err)
			}
			return
		}
	}

	if ctx.CallbackQuery == nil {
		// lookup and execute command handler
		if err := b.ExecuteHandler(ctx); err != nil {
			log.Error("failed to send response:", err)
		}
	}
}

// ExecuteHandler parse the incoming message text, lookup a suitable handler,
// execute middlewares and send the reponse
func (b *Bot) ExecuteHandler(ctx *Context) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s\n%s", e, debug.Stack())
		}
	}()

	// lookup a handler by the received command
	h, ok := b.lookupHandler(ctx.Message.Text)
	if !ok {
		return fmt.Errorf("command %s, handler not found", ctx.Text)
	}

	f := h.callback

	// execute global middlewares
	for _, mid := range b.middlewares {
		if f, ok = mid(f, ctx); !ok {
			return nil
		}
	}

	// execute handler middlewares
	for _, mid := range h.middlewares {
		if f, ok = mid(f, ctx); !ok {
			return nil
		}
	}

	// execute handler
	smsg := f(ctx)
	if smsg == nil {
		return nil
	}

	_, err = ctx.Send(smsg)
	return err
}

// lookupHandler try to match command to by regular expression for each handler
func (b *Bot) lookupHandler(cmd string) (handler, bool) {
	for _, h := range b.handlers {
		if h.exp.MatchString(cmd) {
			return h, true
		}
	}

	return handler{}, false
}

//
// PRE
//

// MiddlewareFunc recieve selected handler and current context on incoming arguments
// it can substitute `next` handler with some other.
// if middleware function return false on second variable, it stop further proccess
type MiddlewareFunc func(next HandleFunc, ctx *Context) (HandleFunc, bool)

// Pre method is add middleware function executed for all handlers
// and before handler middlewares
func (b *Bot) Pre(mid MiddlewareFunc) {
	b.middlewares = append(b.middlewares, mid)
}

// UpdatesFunc recieve current context in start of routing,
// it can change some incoming data
// if return `false` then stop further routing process
type UpdatesFunc func(ctx *Context) bool

// UpdatesHandle add special callback function to recieve incoming updates
func (b *Bot) UpdatesHandle(h UpdatesFunc) {
	b.updatesHandlers = append(b.updatesHandlers, h)
}

//
// Context
//

// Context argument for handlers
type Context struct {
	Bot *Bot

	Update
	Message

	chat *chat
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

	ctx.chat = b.Chats.Get(ctx.Message)

	if u.CallbackQuery != nil {
		ctx.chat.setEditMessageID(u.CallbackQuery.Message.MessageID)
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
func (ctx *Context) SendTextMessage(text string, args ...interface{}) (int, error) {
	return ctx.Send(ctx.NewMessage(fmt.Sprintf(text, args...)))
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
