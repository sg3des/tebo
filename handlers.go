package tebo

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"time"
)

var PollInterval = 30 * time.Second

type HandleFunc func(*Context)

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
			b.route(updates)
		}

		time.Sleep(PollInterval)
	}
}

func (b *Bot) route(updates []Update) {
	for _, u := range updates {
		go func(u Update) {
			ctx := &Context{
				Bot:     b,
				Update:  u,
				Message: u.Message,
			}

			for _, h := range b.updatesHandlers {
				if pass := h(ctx); !pass {
					return
				}
			}

			// if expect enable
			if b.expectUpdate != nil && b.expectCancel != nil {
				// if message is command, started from '/', then cancel expect channels
				// and lookup appropriate handler for this command
				if len(u.Message.Text) > 0 && u.Message.Text[0] == '/' {
					b.expectCancel <- true
					b.closeExpectChannels()
				} else {

					// if message is not command then send this message to expect channel
					// and then do not lookup handler for this message
					b.expectUpdate <- u
					b.closeExpectChannels()
					return
				}
			}

			if err := b.ExecuteHandler(ctx); err != nil {
				log.Error("failed to send response: ", err)
			}
		}(u)
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
	h, ok := b.lookupHandler(ctx.Text)
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
	f(ctx)

	return nil
}

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

type MiddlewareFunc func(next HandleFunc, ctx *Context) (HandleFunc, bool)

// Pre method is add middleware function executed for all handlers
// and before handler middlewares
func (b *Bot) Pre(mid MiddlewareFunc) {
	b.middlewares = append(b.middlewares, mid)
}

type UpdatesHandleFunc func(ctx *Context) bool

func (b *Bot) UpdatesHandle(h UpdatesHandleFunc) {
	b.updatesHandlers = append(b.updatesHandlers, h)
}

//
// Expect answer
//

// ExpectAnswer wait next message, intercept it if this message not a command
// return false if next message is command
func (b *Bot) ExpectAnswer(chatid int) (answer Update, ok bool) {
	b.expectUpdate = make(chan Update)
	b.expectCancel = make(chan bool)

	select {
	case answer = <-b.expectUpdate:
		ok = true
	case <-b.expectCancel:
		ok = false
	}

	return
}

func (b *Bot) closeExpectChannels() {
	close(b.expectUpdate)
	close(b.expectCancel)

	b.expectUpdate = nil
	b.expectCancel = nil
}

//
// Context
//

// Context argument for handlers
type Context struct {
	Bot *Bot

	Update
	Message
}

func (ctx *Context) SendMessage(text string, opt ...SendOptions) (Message, error) {
	return ctx.Bot.SendMessage(ctx.Chat.ID, text, opt...)
}
