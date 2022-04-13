package gesture

import (
	"math"

	evdev "github.com/gvalkov/golang-evdev"
	"github.com/jr64/omnisrv/eventids"
	log "github.com/sirupsen/logrus"
)

const (
	SWIPE_DIRECTION_LEFT = iota
	SWIPE_DIRECTION_UP
	SWIPE_DIRECTION_RIGHT
	SWIPE_DIRECTION_DOWN
	SWIPE_DIRECTION_UNDEFINED
	SWIPE_DIRECTION_INVALID
)

const JITTER_TOLERANCE_DISTANCE = 15
const ALLOWED_DEVIATION_DEGREES = 20
const MIN_DISTANCE_DIRECTION_SET = 100
const MIN_DISTANCE_EMIT_EVENT = 350

type SwipeDetector struct {
	startTime int64
	startX    int32
	startY    int32
	lastX     int32
	lastY     int32
	curX      int32
	curY      int32
	lastDist  float64
	direction int
	complete  bool
}

func (d *GestureDetector) startSwipeDetection() {
	for i := 0; i < MAX_SUPPORTED_FINGERS; i++ {
		d.swipe[i].startX = -1
		d.swipe[i].startY = -1
		d.swipe[i].lastX = -1
		d.swipe[i].lastY = -1
		d.swipe[i].curX = -1
		d.swipe[i].curY = -1
		d.swipe[i].lastDist = -1
		d.swipe[i].direction = SWIPE_DIRECTION_UNDEFINED
		d.swipe[i].complete = false
	}
}

func (d *GestureDetector) processSwipeX(evnt *evdev.InputEvent) {
	if d.trackingId == -1 {
		log.Warn("Received ABS_MT_POSITION_X before ABS_MT_TRACKING_ID")
		return
	}

	if d.trackingId >= MAX_SUPPORTED_FINGERS {
		return
	}

	sw := &d.swipe[d.trackingId]

	if sw.startX == -1 {
		sw.startX = evnt.Value
		sw.lastX = evnt.Value
		sw.startTime = evnt.Time.Nano()
		return
	}

	if sw.curX == -1 {
		sw.curX = evnt.Value
	}

	if sw.curX != -1 && sw.curY != -1 && sw.startX != -1 && sw.startY != -1 {
		d.processSwipeXY(sw)
	}

}

func (d *GestureDetector) processSwipeY(evnt *evdev.InputEvent) {
	if d.trackingId == -1 {
		log.Warn("Received ABS_MT_POSITION_X before ABS_MT_TRACKING_ID")
		return
	}

	if d.trackingId >= MAX_SUPPORTED_FINGERS {
		return
	}

	sw := &d.swipe[d.trackingId]

	if sw.startY == -1 {
		sw.startY = evnt.Value
		sw.lastY = evnt.Value
		sw.startTime = evnt.Time.Nano()
		return
	}

	if sw.curY == -1 {
		sw.curY = evnt.Value
	}

	if sw.curX != -1 && sw.curY != -1 && sw.startX != -1 && sw.startY != -1 {
		d.processSwipeXY(sw)
	}

}

func (d *GestureDetector) processSwipeXY(sw *SwipeDetector) {

	if !inJitterFilter(sw) {
		dist := distance(sw.startX, sw.startY, sw.curX, sw.curY)

		if sw.direction == SWIPE_DIRECTION_UNDEFINED && dist >= MIN_DISTANCE_DIRECTION_SET {
			ang := calcAngle(sw)
			switch {
			case ang >= 90-ALLOWED_DEVIATION_DEGREES && ang <= 90+ALLOWED_DEVIATION_DEGREES:
				sw.direction = SWIPE_DIRECTION_LEFT
			case ang >= 180-ALLOWED_DEVIATION_DEGREES && ang <= 180+ALLOWED_DEVIATION_DEGREES:
				sw.direction = SWIPE_DIRECTION_UP
			case ang >= 270-ALLOWED_DEVIATION_DEGREES && ang <= 270+ALLOWED_DEVIATION_DEGREES:
				sw.direction = SWIPE_DIRECTION_RIGHT
			case ang >= 360-ALLOWED_DEVIATION_DEGREES || ang <= 0+ALLOWED_DEVIATION_DEGREES:
				sw.direction = SWIPE_DIRECTION_DOWN
			default:
				sw.direction = SWIPE_DIRECTION_INVALID
			}
			log.Infof("Angle=%f direction=%d", ang, sw.direction)
		} else if dist < sw.lastDist-JITTER_TOLERANCE_DISTANCE {
			log.Infof("Swipe invalid due to distance loss. distance=%f lastdistance=%f", dist, sw.lastDist)
			sw.direction = SWIPE_DIRECTION_INVALID
		} else {
			ang := calcAngle(sw)
			switch sw.direction {
			case SWIPE_DIRECTION_LEFT:
				if ang < 90-ALLOWED_DEVIATION_DEGREES || ang > 90+ALLOWED_DEVIATION_DEGREES {
					log.Infof("SWIPE_LEFT invalid due to angle deviation %f", ang)
					sw.direction = SWIPE_DIRECTION_INVALID
				}
			case SWIPE_DIRECTION_UP:
				if ang < 180-ALLOWED_DEVIATION_DEGREES || ang > 180+ALLOWED_DEVIATION_DEGREES {
					log.Infof("SWIPE_UP invalid due to angle deviation %f", ang)
					sw.direction = SWIPE_DIRECTION_INVALID
				}
			case SWIPE_DIRECTION_RIGHT:
				if ang < 270-ALLOWED_DEVIATION_DEGREES || ang > 270+ALLOWED_DEVIATION_DEGREES {
					log.Infof("SWIPE_RIGHT invalid due to angle deviation %f", ang)
					sw.direction = SWIPE_DIRECTION_INVALID
				}
			case SWIPE_DIRECTION_DOWN:
				if ang < 360-ALLOWED_DEVIATION_DEGREES && ang > 0+ALLOWED_DEVIATION_DEGREES {
					log.Infof("SWIPE_DOWN invalid due to angle deviation %f", ang)
					sw.direction = SWIPE_DIRECTION_INVALID
				}
			}

		}

		if !sw.complete && dist >= MIN_DISTANCE_EMIT_EVENT && (sw.direction == SWIPE_DIRECTION_LEFT || sw.direction == SWIPE_DIRECTION_UP || sw.direction == SWIPE_DIRECTION_RIGHT || sw.direction == SWIPE_DIRECTION_DOWN) {
			emit := true
			fingerCnt := 0
			sw.complete = true
			dir := sw.direction

			for i := 0; i < MAX_SUPPORTED_FINGERS; i++ {
				if d.swipe[i].direction == SWIPE_DIRECTION_UNDEFINED {
					continue
				}
				if d.swipe[i].complete && d.swipe[i].direction == dir {
					fingerCnt += 1
				} else if d.swipe[i].direction != SWIPE_DIRECTION_UNDEFINED {
					emit = false
				}
			}
			if emit {
				log.Infof("EMIT %d finger swipe event direction %d", fingerCnt, dir)
				switch {
				case fingerCnt == 1 && dir == SWIPE_DIRECTION_RIGHT:
					d.eventC <- eventids.EVENT_SWIPE_RIGHT_1_FINGER

				case fingerCnt == 1 && dir == SWIPE_DIRECTION_DOWN:
					d.eventC <- eventids.EVENT_SWIPE_DOWN_1_FINGER

				case fingerCnt == 1 && dir == SWIPE_DIRECTION_LEFT:
					d.eventC <- eventids.EVENT_SWIPE_LEFT_1_FINGER

				case fingerCnt == 1 && dir == SWIPE_DIRECTION_UP:
					d.eventC <- eventids.EVENT_SWIPE_UP_1_FINGER

				case fingerCnt == 2 && dir == SWIPE_DIRECTION_RIGHT:
					d.eventC <- eventids.EVENT_SWIPE_RIGHT_2_FINGERS

				case fingerCnt == 2 && dir == SWIPE_DIRECTION_DOWN:
					d.eventC <- eventids.EVENT_SWIPE_DOWN_2_FINGERS

				case fingerCnt == 2 && dir == SWIPE_DIRECTION_LEFT:
					d.eventC <- eventids.EVENT_SWIPE_LEFT_2_FINGERS

				case fingerCnt == 2 && dir == SWIPE_DIRECTION_UP:
					d.eventC <- eventids.EVENT_SWIPE_UP_2_FINGERS
				}
			} else {
				log.Infof("DO NOT EMIT %d finger swipe event", fingerCnt)
			}
		}
		sw.lastX = sw.curX
		sw.lastY = sw.curY
		sw.lastDist = dist
	}

	sw.curX = -1
	sw.curY = -1
}

func inJitterFilter(sw *SwipeDetector) bool {
	if math.Abs(float64(sw.curX)-float64(sw.lastX)) < JITTER_TOLERANCE_DISTANCE && math.Abs(float64(sw.curY)-float64(sw.lastY)) < JITTER_TOLERANCE_DISTANCE {
		return true
	}
	return false
}

func calcAngle(sw *SwipeDetector) float64 {
	ang := (math.Atan2(float64(sw.curY-sw.startY), float64(sw.curX-sw.startX)) * 180) / math.Pi
	if ang < 0 {
		ang += 360
	}
	return ang
}
