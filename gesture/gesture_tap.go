package gesture

import (
	"math"

	evdev "github.com/gvalkov/golang-evdev"
	log "github.com/sirupsen/logrus"
)

const MAX_TAP_DETECTION_DISTANCE = 100
const MAX_TAP_DETECTION_DURATION = 200000000

const MAX_DOUBLETAP_DETECTION_DISTANCE = 300
const MAX_DOUBLETAP_DETECTION_DURATION = 400000000

type TapDetector struct {
	minX      int32
	minY      int32
	maxX      int32
	maxY      int32
	startTime int64
	endTime   int64
}

type TapEvent struct {
	x    int32
	y    int32
	time int64
}

func (d *GestureDetector) startTapDetection() {
	for i := 0; i < MAX_SUPPORTED_FINGERS; i++ {
		d.taps[i].maxX = -1
		d.taps[i].maxY = -1
		d.taps[i].minX = math.MaxInt32
		d.taps[i].minY = math.MaxInt32

		d.taps[i].startTime = -1
		d.taps[i].endTime = -1
	}

}

func (d *GestureDetector) stopTapDetection() []TapEvent {
	res := make([]TapEvent, 0)
	for i := 0; i < MAX_SUPPORTED_FINGERS; i++ {
		tap := d.taps[i]
		if tap.maxX == -1 || tap.maxY == -1 || tap.minX == math.MaxInt32 || tap.minY == math.MaxInt32 || tap.startTime == -1 || tap.endTime == -1 {
			log.Debugf("Tap detection: skip finger=%d minX=%d minY=%d maxX=%d maxY=%d start=%d end=%d", i, tap.minX, tap.minY, tap.maxX, tap.maxY, tap.startTime, tap.endTime)
			continue
		}

		dist := distance(tap.maxX, tap.maxY, tap.minX, tap.minY)
		dur := tap.endTime - tap.startTime

		log.Infof("Tap detection: finger=%d distance=%f duration=%d", i, dist, dur)

		if dist < MAX_TAP_DETECTION_DISTANCE && dur < MAX_TAP_DETECTION_DURATION {
			res = append(res, TapEvent{
				x:    tap.minX + ((tap.maxX - tap.minX) / 2),
				y:    tap.minY + ((tap.maxY - tap.minY) / 2),
				time: tap.startTime + ((tap.endTime - tap.startTime) / 2),
			})

		}

	}

	return res
}

func (d *GestureDetector) processTapX(evnt *evdev.InputEvent) {
	if d.trackingId == -1 {
		log.Warn("Received ABS_MT_POSITION_X before ABS_MT_TRACKING_ID")
		return
	}

	if d.trackingId >= MAX_SUPPORTED_FINGERS {
		return
	}

	if d.taps[d.trackingId].maxX < evnt.Value {
		d.taps[d.trackingId].maxX = evnt.Value
	}

	if d.taps[d.trackingId].minX > evnt.Value {
		d.taps[d.trackingId].minX = evnt.Value
	}

	if d.taps[d.trackingId].startTime == -1 {
		d.taps[d.trackingId].startTime = evnt.Time.Nano()
	}

	d.taps[d.trackingId].endTime = evnt.Time.Nano()

}

func (d *GestureDetector) processTapY(evnt *evdev.InputEvent) {
	if d.trackingId == -1 {
		log.Warn("Received ABS_MT_POSITION_Y before ABS_MT_TRACKING_ID")
		return
	}

	if d.trackingId >= MAX_SUPPORTED_FINGERS {
		return
	}

	if d.taps[d.trackingId].maxY < evnt.Value {
		d.taps[d.trackingId].maxY = evnt.Value
	}

	if d.taps[d.trackingId].minY > evnt.Value {
		d.taps[d.trackingId].minY = evnt.Value
	}

	if d.taps[d.trackingId].startTime == -1 {
		d.taps[d.trackingId].startTime = evnt.Time.Nano()
	}

	d.taps[d.trackingId].endTime = evnt.Time.Nano()

}
