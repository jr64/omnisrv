package gesture

import (
	evdev "github.com/gvalkov/golang-evdev"
	"github.com/jr64/omnisrv/eventids"
	log "github.com/sirupsen/logrus"
)

const MAX_SUPPORTED_FINGERS = 2

type GestureDetector struct {
	gestureStarted bool
	trackingId     int32
	taps           [MAX_SUPPORTED_FINGERS]TapDetector
	lastTap        []TapEvent
	swipe          [MAX_SUPPORTED_FINGERS]SwipeDetector
	eventC         chan int
}

func NewGestureDetector(c chan int) *GestureDetector {
	return &GestureDetector{
		gestureStarted: false,
		trackingId:     -1,
		eventC:         c,
	}
}

func (d *GestureDetector) ProcessEvent(evnt *evdev.InputEvent) {

	if evnt.Type == evdev.EV_SYN && evnt.Code == evdev.SYN_MT_REPORT {
		//Reset current event batch data
		d.trackingId = -1
		return
	}

	if evnt.Type == evdev.EV_ABS && evnt.Code == evdev.ABS_MT_TRACKING_ID {
		if d.trackingId != -1 {
			log.Warnf("Received ABS_MT_TRACKING_ID but ID is already %d", d.trackingId)
		}
		if d.trackingId >= MAX_SUPPORTED_FINGERS {
			log.Warnf("Received ABS_MT_TRACKING_ID %d but only %d fingers are supported", d.trackingId, MAX_SUPPORTED_FINGERS)
		}
		d.trackingId = evnt.Value
		return
	}

	if evnt.Type == evdev.EV_ABS && evnt.Code == evdev.ABS_MT_POSITION_X {
		log.Debugf("Received X for finger %d", d.trackingId)
		d.processTapX(evnt)
		d.processSwipeX(evnt)
		return
	}

	if evnt.Type == evdev.EV_ABS && evnt.Code == evdev.ABS_MT_POSITION_Y {
		log.Debugf("Received Y for finger %d", d.trackingId)
		d.processTapY(evnt)
		d.processSwipeY(evnt)
		return
	}

	if evnt.Type == evdev.EV_KEY && evnt.Code == evdev.BTN_TOUCH && evnt.Value == 1 {
		log.Infof("Gesture started")
		d.trackingId = -1

		d.gestureStarted = true
		d.startTapDetection()
		d.startSwipeDetection()
		return
	}

	if evnt.Type == evdev.EV_KEY && evnt.Code == evdev.BTN_TOUCH && evnt.Value == 0 {
		log.Infof("Gesture ended")

		d.gestureStarted = false
		taps := d.stopTapDetection()
		log.Infof("Detected %d finger tap", len(taps))
		if len(taps) > 0 && d.lastTap != nil && len(d.lastTap) > 0 {

			tapTime := taps[0].time - d.lastTap[0].time
			tapDist := distance(d.lastTap[0].x, d.lastTap[0].y, taps[0].x, taps[0].y)
			if tapTime < MAX_DOUBLETAP_DETECTION_DURATION && tapDist < MAX_DOUBLETAP_DETECTION_DISTANCE {
				log.Infof("Double tap detected: distance=%f duration=%d", tapDist, tapTime)
				if len(d.lastTap) == 1 && len(taps) == 1 {
					d.eventC <- eventids.EVENT_DOUBLE_TAP_1_FINGER
				} else if len(d.lastTap) == 2 && len(taps) == 2 {
					tapDist = distance(d.lastTap[1].x, d.lastTap[1].y, taps[1].x, taps[1].y)
					if tapDist < MAX_DOUBLETAP_DETECTION_DISTANCE {
						d.eventC <- eventids.EVENT_DOUBLE_TAP_2_FINGERS
					}
				}
				taps = nil
			} else {
				log.Debugf("No double tap detected: distance=%f duration=%d", tapDist, tapTime)
			}

		}
		d.lastTap = taps
		return
	}

}
