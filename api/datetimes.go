package api

import "time"

var dateFormats = []string{
	// most common first
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02",
	"01-02-2006",
	"01-02-06",
	"01-2-06",
	"1-2-06",
	"01-2-2006",
	"1-2-2006",
	"01/02/2006",
	"1/02/2006",
	"01/02/06",
	"01/2/06",
	"1/2/06",
	"1/2/2006",
	"01/2/2006",
	"1/2/2006",
	"January 2, 2006",
	"January 2, 06",
}

const (
	timeFormatDB  = "2006-01-02 15:04:05"
	timeFormatAPI = "2006-01-02T15:04:05Z"
	dateFormat    = "2006-01-02"
)

// parseTime parses a time against the supported formats
func parseTime(input string) (time.Time, error) {
	//run through all of the supported formats and try to convert
	var err error
	var parsed = time.Time{}
	for i := range dateFormats {
		parsed, err = time.Parse(dateFormats[i], input)
		if err == nil {
			break
		}
	}
	return parsed, err
}

// parseTimeToTimeFormat is the preferred mechanism for time parsing a timestamp
func parseTimeToTimeFormat(input, format string) (string, error) {
	t, err := parseTime(input)
	if err != nil {
		return input, err
	}
	return t.Format(format), nil
}

// CalculateDuration gets the duration from a certain date time to now.
// Timezone is not important for this one and we always calculate from now.
// Shamelessly adapted from icza at https://stackoverflow.com/questions/36530251/golang-time-since-with-months-and-years. Support
// him through his profile: https://stackoverflow.com/users/1705598/icza
func CalculateDuration(input time.Time) (year, month, day, hour, min, sec int) {
	now := time.Now()
	y1, M1, d1 := input.Date()
	y2, M2, d2 := now.Date()

	h1, m1, s1 := input.Clock()
	h2, m2, s2 := now.Clock()

	year = int(y2 - y1)
	month = int(M2 - M1)
	day = int(d2 - d1)
	hour = int(h2 - h1)
	min = int(m2 - m1)
	sec = int(s2 - s1)

	// Normalize negative values
	if sec < 0 {
		sec += 60
		min--
	}
	if min < 0 {
		min += 60
		hour--
	}
	if hour < 0 {
		hour += 24
		day--
	}
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
