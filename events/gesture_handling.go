package events

import (
	"github.com/jr64/omnisrv/tolinoutil"
	log "github.com/sirupsen/logrus"
)

func togglePowerIfOff() {
	screenOn, err := tolinoutil.IsScreenOn()

	if err != nil {
		log.Errorf("Failed to get screen state: %v", err)
	}

	if screenOn {
		return
	}

	err = tolinoutil.SendPowerButtonPress()
	if err != nil {
		log.Errorf("Failed to turn on device on: %v", err)
	}
}

func togglePowerUnconditional() {
	err := tolinoutil.SendPowerButtonPress()
	if err != nil {
		log.Errorf("Failed to turn device on/off: %v", err)
	}
}
