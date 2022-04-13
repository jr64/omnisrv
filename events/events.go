package events

import (
	"github.com/jr64/omnisrv/androidutil"
	"github.com/jr64/omnisrv/eventids"
	"github.com/jr64/omnisrv/tolinoutil"
	log "github.com/sirupsen/logrus"
)

func HandleEventsFromChan(c chan int) {
	for evnt := range c {
		HandleEvent(evnt)
	}
}

func HandleEvent(evnt int) {
	switch evnt {
	case eventids.EVENT_DOUBLE_TAP_1_FINGER:
		log.Info("EVENT_DOUBLE_TAP_1_FINGER")
		togglePowerIfOff()

	case eventids.EVENT_SWIPE_UP_2_FINGERS:
		log.Info("EVENT_SWIPE_UP_2_FINGERS")
		togglePowerUnconditional()

	case eventids.EVENT_SWIPE_DOWN_2_FINGERS:
		log.Info("EVENT_SWIPE_DOWN_2_FINGERS")
		screenOn, err := tolinoutil.IsScreenOn()
		if err != nil {
			log.Errorf("failed to read screen state: %v", err)
		} else if screenOn {
			androidutil.StartActivity("com.gacode.relaunchx/.Home")
		}

	case eventids.EVENT_FB_SLEEP:
		log.Info("EVENT_FB_SLEEP")

	case eventids.EVENT_FB_WAKE:
		log.Info("EVENT_FB_WAKE")
	}
}
