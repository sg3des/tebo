package tebo

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"time"
)

var PollInterval = 30 * time.Second

type HandleFunc func(Message) string

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
UPDATES:
	for _, u := range updates {

		for _, h := range b.updatesHandlers {
			if pass := h(&u); !pass {
				continue UPDATES
			}
		}

		if len(u.Message.Text) == 0 {
			continue
		}

		// if expect enable
		if b.expect != nil && b.expectCancel != nil {
			// if message is not command then send this message to expect channel
			// and then do not lookup handler for this message
			if u.Message.Text[0] != '/' {
				b.expect <- u.Message
				b.closeExpectChannels()
				continue
			} else {
				// if message is command, started from '/', then cancel expect channels
				// and lookup appropriate handler for this command
				b.expectCancel <- true
				b.closeExpectChannels()
			}
		}

		go func(msg Message) {
			if err := b.ExecuteHandler(msg); err != nil {
				log.Error(err)
			}
		}(u.Message)
	}
}

// ExecuteHandler parse the incoming message text, lookup a suitable handler,
// execute middlewares and send the reponse
func (b *Bot) ExecuteHandler(msg Message) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%s\n%s", e, debug.Stack())
		}
	}()

	// lookup a handler by the received command
	h, ok := b.lookupHandler(msg.Text)
	if !ok {
		return fmt.Errorf("command %s, handler not found", msg.Text)
	}

	f := h.callback

	// execute global middlewares
	for _, mid := range b.middlewares {
		if f, ok = mid(f, msg); !ok {
			return nil
		}
	}

	// execute handler middlewares
	for _, mid := range h.middlewares {
		if f, ok = mid(f, msg); !ok {
			return nil
		}
	}

	// execute handler
	resp := f(msg)

	// send a response to this chat
	if err := b.SendMessage(msg.Chat.ID, resp); err != nil {
		return fmt.Errorf("failed send response: %v", err)
	}

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

type MiddlewareFunc func(next HandleFunc, msg Message) (HandleFunc, bool)

// Pre method is add middleware function executed for all handlers
// and before handler middlewares
func (b *Bot) Pre(mid MiddlewareFunc) {
	b.middlewares = append(b.middlewares, mid)
}

type UpdatesHandleFunc func(u *Update) bool

func (b *Bot) UpdatesHandle(h UpdatesHandleFunc) {
	b.updatesHandlers = append(b.updatesHandlers, h)
}

//
// Expect answer
//

// ExpectAnswer wait next message intercept it if this message not a command
// return false if next message is command
func (b *Bot) ExpectAnswer(msg Message, text string, opt ...SendOptions) (answer Message, ok bool) {
	if err := b.SendMessage(msg.Chat.ID, text, opt...); err != nil {
		return
	}

	b.expect = make(chan Message)
	b.expectCancel = make(chan bool)

	select {
	case answer = <-b.expect:
		ok = true
	case <-b.expectCancel:
		ok = false
	}

	return
}

func (b *Bot) closeExpectChannels() {
	close(b.expect)
	close(b.expectCancel)

	b.expect = nil
	b.expectCancel = nil
}
