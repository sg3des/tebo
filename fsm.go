package tebo

import (
	"fmt"
)

const letters = "0123456789abcdefghijklmnopqrstuvwzyxABCDEFGHIJKLMNOPQRSTUVWZYX"
const fsmrootid = "."

type FSM struct {
	id      string
	handler HandleFunc
	buttons []*fsmButton

	root *FSM
}

func NewFSM(h HandleFunc) *FSM {
	fsm := &FSM{
		handler: h,
		id:      fsmrootid,
	}

	fsm.root = fsm

	return fsm
}

type FSMButtonBuilder func(ctx *Context) *InlineKeyboardButton

type fsmButton struct {
	text string
	f    FSMButtonBuilder
	fsm  *FSM
}

func (fsm *FSM) Add(text string, h HandleFunc) *FSM {
	btn := &fsmButton{
		text: text,
		fsm:  fsm.newState(h),
	}

	fsm.buttons = append(fsm.buttons, btn)

	return btn.fsm
}

func (fsm *FSM) AddFunc(f FSMButtonBuilder, h HandleFunc) *FSM {
	btn := &fsmButton{
		f:   f,
		fsm: fsm.newState(h),
	}

	fsm.buttons = append(fsm.buttons, btn)

	return btn.fsm
}

func (fsm *FSM) newState(h HandleFunc) *FSM {
	return &FSM{
		id:      fsm.nextID(),
		handler: h,
		root:    fsm.root,
	}
}

func (fsm *FSM) nextID() string {
	return fmt.Sprintf("%s%c", fsm.id, letters[len(fsm.buttons)])
}

func (fsm *FSM) newID(n int) string {
	return fmt.Sprintf("%s%c", fsm.id, letters[n])
}

func (fsm *FSM) handle(ctx *Context) (err error) {
	ctx.chat.fsm = fsm.root

	if ctx.CallbackQuery == nil {
		return fsm.initialMessage(ctx)
	}

	id := ctx.CallbackQuery.Data

	fsm, ok := fsm.root.lookupState(id)
	if !ok {
		return fmt.Errorf("state by id:%s not found", id)
	}

	var msg Message
	smsg := fsm.message(ctx)
	if ctx.chat.lastMessageIsBot {
		msg, err = ctx.Edit(ctx.chat.editMessageID, smsg)
	} else {
		msg, err = ctx.Send(smsg)
	}
	if err != nil {
		return err
	}

	ctx.chat.setEditMessageID(msg.MessageID)

	return nil
}

func (fsm *FSM) initialMessage(ctx *Context) error {
	msg, err := ctx.Send(fsm.message(ctx))
	if err != nil {
		return err
	}

	ctx.chat.setEditMessageID(msg.MessageID)

	return nil
}

func (fsm *FSM) lookupState(id string) (*FSM, bool) {
	if id == fsmrootid {
		return fsm.root, true
	}
	if fsm.id == id {
		return fsm, true
	}

	for _, btn := range fsm.buttons {
		if btnfsm, ok := btn.fsm.lookupState(id); ok {
			return btnfsm, ok
		}
	}

	return nil, false
}

func (fsm *FSM) message(ctx *Context) *SendMessage {
	smsg := fsm.handler(ctx)
	if smsg.ReplyMarkup == nil {
		smsg.ReplyMarkup = fsm.keyboard(ctx)
	} else {
		var i int
		markup := smsg.ReplyMarkup.(InlineKeyboardMarkup)
		for _, row := range markup.InlineKeyboard {
			for _, btn := range row {
				btn.CallbackData = fsm.newID(i)
				i++
			}
		}
	}

	return smsg
}

func (fsm *FSM) keyboard(ctx *Context) *InlineKeyboardMarkup {
	keyboard := NewInlineKeyboard(2)
	for _, btn := range fsm.buttons {
		if btn.f != nil {
			b := btn.f(ctx)
			if b == nil {
				continue
			}
			if b.CallbackData == "" {
				b.CallbackData = btn.fsm.id
			}

			keyboard.Add(*b)
		} else {
			keyboard.AddButton(btn.text, btn.fsm.id)
		}
	}

	if fsm.id != fsmrootid {
		keyboard.AddButton("« Back", fsm.id[:len(fsm.id)-1])
	}

	return keyboard.ToReplyMarkup()
}

func (fsm *FSM) NewMessage(ctx *Context, text string, opt ...SendOptions) *SendMessage {
	ctx.chat.fsm = fsm

	smsg := ctx.NewMessage(text, opt...)
	smsg.ReplyMarkup = fsm.keyboard(ctx)
	return smsg
}
