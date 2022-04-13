package main

import (
	"fmt"

	evdev "github.com/gvalkov/golang-evdev"
	"github.com/jr64/omnisrv/events"
	"github.com/jr64/omnisrv/gesture"
	"github.com/jr64/omnisrv/tolinoutil"
	log "github.com/sirupsen/logrus"
)

const INPUT_EVENT_PATH = "/dev/input/event1"

func main() {

	eventC := make(chan int, 10)

	go func() {
		err := tolinoutil.MonitorFb(eventC)
		if err != nil {
			log.Errorf("failed to monitor screen state: %v", err)
		}
	}()

	go events.HandleEventsFromChan(eventC)

	device, _ := evdev.Open(INPUT_EVENT_PATH)
	fmt.Println(device)

	gestureD := gesture.NewGestureDetector(eventC)

	for {
		events, err := device.Read()

		if err != nil {
			fmt.Printf("Failed to read events: %v", err)
		} else {
			for _, evnt := range events {
				gestureD.ProcessEvent(&evnt)
			}
		}
	}
}
