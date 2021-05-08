package main

import (
	"github.com/gotd/td/tg"
)

func (t *Troll) Register(dispatcher tg.UpdateDispatcher) {
	dispatcher.OnUserStatus(t.OnUserStatus)
	dispatcher.OnUserName(t.OnUserName)
	dispatcher.OnUserPhoto(t.OnUserPhoto)
	dispatcher.OnUserTyping(t.OnUserTyping)
}
