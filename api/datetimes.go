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
