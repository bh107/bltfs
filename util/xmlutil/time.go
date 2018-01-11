package xmlutil

import "time"

const FormatISO8601 = "2006-01-02T15:04:05.999999999Z"

type Time time.Time

func TimeNow() Time {
	now := time.Now()
	return Time(now.UTC())
}

func (t Time) UnixNano() int64 {
	t2 := time.Time(t)
	return t2.UnixNano()
}

func Unix(sec int64, usec int64) Time {
	return Time(time.Unix(sec, usec))
}
