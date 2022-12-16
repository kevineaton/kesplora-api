package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateParsingOLD(t *testing.T) {
	tests := []struct {
		inputString          string
		inputFormat          string
		expectedOutputString string
		shouldError          bool
	}{
		{
			inputString:          "2005-11-06T01:54:01Z",
			inputFormat:          "2006-01-02 15:04:05",
			expectedOutputString: "2005-11-06 01:54:01",
			shouldError:          false,
		},
		{
			inputString:          "2005-11-06 01:54:01",
			inputFormat:          "2006-01-02T15:04:05Z",
			expectedOutputString: "2005-11-06T01:54:01Z",
			shouldError:          false,
		},
		{
			inputString:          "2008-01-01 17:08:31",
			inputFormat:          "2006-01-02 15:04:05",
			expectedOutputString: "2008-01-01 17:08:31",
			shouldError:          false,
		},
		{
			inputString:          "2008-01-01",
			inputFormat:          "2006-01-02",
			expectedOutputString: "2008-01-01",
			shouldError:          false,
		},
		{
			inputString:          "2008-01-01",
			inputFormat:          "2006-01-02 15:04:05",
			expectedOutputString: "2008-01-01 00:00:00",
			shouldError:          false,
		},
		{
			inputString:          "2008-01-01T17:08:31",
			inputFormat:          "2006-01-02T15:04:05Z",
			expectedOutputString: "2008-01-01T17:08:31Z",
			shouldError:          false,
		},
		{
			inputString:          "2008-01-01T17:08:31+05:00",
			inputFormat:          "2006-01-02T15:04:05Z",
			expectedOutputString: "2008-01-01T17:08:31Z",
			shouldError:          false,
		},
		{
			inputString:          "1985-01-13T15:23:00Z",
			inputFormat:          "01/02/2006 3:04:05 PM",
			expectedOutputString: "01/13/1985 3:23:00 PM",
			shouldError:          false,
		},
		{
			inputString:          ";alsdfh",
			inputFormat:          "01/02/2006 3:04:05 PM",
			expectedOutputString: ";alsdfh",
			shouldError:          true,
		},
	}

	count := 0
	for _, tt := range tests {
		converted, err := parseTimeToTimeFormat(tt.inputString, tt.inputFormat)
		assert.Equal(t, tt.expectedOutputString, converted, fmt.Sprintf("passed in %s with format of %s, got %s", tt.inputString, tt.inputFormat, converted))
		if tt.shouldError {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		count++
	}
	assert.Equal(t, len(tests), count)
}

func TestDateDurationCalculations(t *testing.T) {
	tests := []struct {
		inputTime     time.Time
		minimumYears  int
		minimumMonths int
	}{
		{
			inputTime:     time.Now().AddDate(-22, -1, 0),
			minimumYears:  22,
			minimumMonths: 1,
		},
		{
			inputTime:     time.Now().AddDate(0, -1, 0),
			minimumYears:  0,
			minimumMonths: 1,
		},
	}

	count := 0
	for _, tt := range tests {
		years, months, _, _, _, _ := CalculateDuration(tt.inputTime)
		assert.True(t, years >= tt.minimumYears && months >= tt.minimumMonths, fmt.Sprintf("passed in %s and had %d years and %d months old", tt.inputTime.Format("2006-01-02"), years, months))
		count++
	}
	assert.Equal(t, len(tests), count)
}
