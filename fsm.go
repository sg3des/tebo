package tebo

import (
	"fmt"
	"time"
)

type FSM struct {
	Text string

	id     string
	parent *FSM

	buttons []*FSMbutton
	handler UpdatesHandleFunc
}

type FSMbutton struct {
	Text string
	fsm  *FSM
}

func NewFSM(text string, parent *FSM) *FSM {
	return &FSM{
		Text:   text,
		id:     fmt.Sprintf("%x", time.Now().UnixNano()),
		parent: parent,
	}
}

func (fsm *FSM) AddButton(text string) *FSM {
	btn := &FSMbutton{
		Text: text,
		fsm:  NewFSM("", fsm),
	}
	fsm.buttons = append(fsm.buttons, btn)

	return btn.fsm
}

func (fsm *FSM) SetText(text string) {
	fsm.Text = text
}

func (fsm *FSM) Handle(h UpdatesHandleFunc) {
	fsm.handler = h
}

func (fsm *FSM) Start(ctx *Context) error {
	if ctx.CallbackQuery == nil {
		return nil
	}

	id := ctx.CallbackQuery.Data

	fsm, ok := fsm.root().lookupState(id)
	if !ok {
		return fmt.Errorf("state id:%s not found", id)
	}

	text, opt := fsm.message()

	_, err := ctx.Bot.SendMessage(ctx.Chat.ID, text, opt)
	if err != nil {
		return err
	}

	return nil
}

func (fsm *FSM) root() *FSM {
	root := fsm
	for root.parent != nil {
		root = root.parent
	}

	return root
}

func (fsm *FSM) lookupState(id string) (*FSM, bool) {
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

func (fsm *FSM) message() (string, SendOptions) {
	keyboard := NewInlineKeyboard(2)
	for _, btn := range fsm.buttons {
		keyboard.AddButton(btn.Text, btn.fsm.id)
	}

	return fsm.Text, SendOptions{
		ReplyMarkup: keyboard.ToReplyMarkup(),
	}
}
