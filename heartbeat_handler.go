package ngicluster

import (
	"time"

	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/handler"
)

type HeartbeatHandler struct {
	*core.DefaultInboundHandler
	allowedTimeoutTimes int
	periodSec           int
	timeoutTimes        int

	lastTimeout time.Time
}

func NewHeartbeatHandler(times int, periodSec int) *HeartbeatHandler {
	hh := &HeartbeatHandler{DefaultInboundHandler: core.NewDefaultInboundHandler()}
	hh.allowedTimeoutTimes = times
	hh.periodSec = periodSec

	return hh
}

func (hh *HeartbeatHandler) OnEvent(ctx *core.ChannelContext, event interface{}) {
	if _, ok := event.(*handler.IdleEvent); ok {
		log.Debugf("HeartbeatHandler timeout ")
		if time.Now().Sub(hh.lastTimeout) > time.Second*time.Duration(hh.periodSec) {
			hh.lastTimeout = time.Now()
			hh.timeoutTimes = 0
		}

		hh.timeoutTimes++
		if hh.timeoutTimes > hh.allowedTimeoutTimes {
			log.Debugf("HeartbeatHandler timeout close Channel")
			ctx.Close()
		}
	}
}
