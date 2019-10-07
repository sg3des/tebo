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

type fsmButton struct {
	Text string
	fsm  *FSM
}

func (fsm *FSM) Add(text string, h HandleFunc) *FSM {
	btn := &fsmButton{
		Text: text,
		fsm:  fsm.newState(h),
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

	// smsg := fsm.message(ctx)
	// if ctx.chat.lastMessageIsBot {
	// 	_, err = ctx.editMessage(ctx.chat.editMessageID, smsg)
	// } else {
	// 	_, err = ctx.sendMessage(smsg)
	// }

	_, err = ctx.sendMessage(fsm.message(ctx))

	return err
}

func (fsm *FSM) initialMessage(ctx *Context) error {
	_, err := ctx.sendMessage(fsm.message(ctx))
	if err != nil {
		return err
	}

	// ctx.chat.editMessageID = msg.MessageID
	// ctx.chat.lastMessageIsBot = true

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
		smsg.ReplyMarkup = fsm.keyboard()
	}

	return smsg
}

func (fsm *FSM) keyboard() *InlineKeyboardMarkup {
	// if len(fsm.buttons) == 0 {
	// 	return nil
	// }

	keyboard := NewInlineKeyboard(2)
	for _, btn := range fsm.buttons {
		keyboard.AddButton(btn.Text, btn.fsm.id)
	}
	if fsm.id != fsmrootid {
		keyboard.AddButton("« Back", fsm.id[:len(fsm.id)-1])
	}

	return keyboard.ToReplyMarkup()
}

func (fsm *FSM) NewMessage(ctx *Context, text string, opt ...SendOptions) *SendMessage {
	ctx.chat.fsm = fsm

	smsg := ctx.NewMessage(text, opt...)
	smsg.ReplyMarkup = fsm.keyboard()
	return smsg
}
