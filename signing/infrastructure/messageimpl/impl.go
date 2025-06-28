package messageimpl

import (
	"github.com/opensourceways/app-cla-server/signing/domain/message"
	"github.com/opensourceways/app-cla-server/signing/watch"
)

func NewMessageImpl() *messageImpl {
	return &messageImpl{}
}

type messageImpl struct {
}

func (impl *messageImpl) SendCLAUpdatedEvent(msg message.CLAUpdatedMsg) {
	watch.SendCLAUpdatedEvent(msg)
}
