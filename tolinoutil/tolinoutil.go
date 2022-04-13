package tolinoutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"

	evdev "github.com/gvalkov/golang-evdev"
	"github.com/jr64/omnisrv/eventids"
)

const FB_WAKE_PATH = "/sys/power/wait_for_fb_wake"
const FB_SLEEP_PATH = "/sys/power/wait_for_fb_sleep"
const POWER_STATE_PATH = "/sys/power/state-extended"

const INPUT_EVENT_PATH_KEYS = "/dev/input/event0"

func MonitorFb(c chan int) error {
	screenOn, err := IsScreenOn()

	if err != nil {
		return err
	}

	for {
		if screenOn {
			err = MonitorFbSleep(c)
			if err != nil {
				return err
			}
			screenOn = false
		} else {
			err = MonitorFbWake(c)
			if err != nil {
				return err
			}
			screenOn = true
		}
	}

}
func MonitorFbSleep(c chan int) error {
	return monitorPath(c, FB_SLEEP_PATH, eventids.EVENT_FB_SLEEP)
}

func MonitorFbWake(c chan int) error {
	return monitorPath(c, FB_WAKE_PATH, eventids.EVENT_FB_WAKE)
}

func monitorPath(c chan int, path string, emit_event int) error {
	f, err := os.Open(path)
	if err != nil {
		return err

	}
	defer f.Close()

	buf := make([]byte, 1)
	_, err = f.Read(buf)
	if err != nil {
		return err
	}
	c <- emit_event

	return nil
}

func IsScreenOn() (bool, error) {
	state, err := ioutil.ReadFile(POWER_STATE_PATH)

	if err != nil {
		return false, err
	}

	stateS := string(state[0])

	if stateS == "1" {
		return false, nil
	} else if stateS == "0" {
		return true, nil
	} else {
		return false, fmt.Errorf("failed to detect screen state: read %s", stateS)
	}

}

func SendPowerButtonPress() error {
	return SendButtonPress(evdev.KEY_POWER)
}

func SendButtonPress(keycode uint16) error {
	events := []evdev.InputEvent{
		{
			Type:  evdev.EV_KEY,
			Code:  keycode,
			Value: 1,
		},
		{
			Type:  evdev.EV_SYN,
			Code:  evdev.SYN_REPORT,
			Value: 0,
		},
		{
			Type:  evdev.EV_KEY,
			Code:  keycode,
			Value: 0,
		},
		{
			Type:  evdev.EV_SYN,
			Code:  evdev.SYN_REPORT,
			Value: 0,
		},
	}

	f, err := os.OpenFile(INPUT_EVENT_PATH_KEYS, os.O_APPEND|os.O_WRONLY, 0440)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := new(bytes.Buffer)

	for i := 0; i < len(events); i++ {

		e := events[i]
		err = binary.Write(buf, binary.LittleEndian, e.Time.Sec)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, e.Time.Usec)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, e.Type)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, e.Code)
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, e.Value)
		if err != nil {
			return err
		}
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
