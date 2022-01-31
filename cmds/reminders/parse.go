package reminders

import (
	"strings"
	"time"
	"unicode/utf8"
)

var formats = []string{
	"2006-01-02",
	"2006-01-02t15:04:05z07:00",
	"3pm",
	"3 pm",
	"3:04 pm",
	"3:04pm",
	"15:04",
	"January 02 at 3:04 pm",
	"January 2 at 3:04 pm",
	"Jan 2 at 3:04 pm",
	"Jan 02 at 3:04 pm",
	"January 02 3:04 pm",
	"January 2 3:04 pm",
	"Jan 2 3:04 pm",
	"January 02 at 3:04pm",
	"January 2 at 3:04pm",
	"Jan 2 at 3:04pm",
	"Jan 02 at 3:04pm",
	"January 02 3:04pm",
	"January 2 3:04pm",
	"January 2 3pm",
	"January 2 3 pm",
	"January 2 at 3 pm",
	"January 2 at 3pm",
	"Jan 2 3:04pm",
	"January 02 15:04",
	"January 02 at 15:04",
	"January 2 15:04",
	"January 2 at 15:04",
	"Jan 2 15:04",
	"Jan 02 15:04",
	"Jan 2 at 15:04",
	"Jan 02 at 15:04",
	"Jan 2",
	"Jan 02",
	"January 2",
	"January 02",
	"January 02 2006 at 3:04 pm",
	"January 2 2006 at 3:04 pm",
	"Jan 2 2006 at 3:04 pm",
	"Jan 02 2006 at 3:04 pm",
	"January 02 2006 3:04 pm",
	"January 2 2006 3:04 pm",
	"Jan 2 2006 3:04 pm",
	"January 02 2006 at 3:04pm",
	"January 2 2006 at 3:04pm",
	"Jan 2 2006 at 3:04pm",
	"Jan 02 2006 at 3:04pm",
	"January 02 2006 3:04pm",
	"January 2 2006 3:04pm",
	"Jan 2 2006 3:04pm",
	"January 02 2006 15:04",
	"January 02 2006 at 15:04",
	"January 2 2006 15:04",
	"January 2 2006 at 15:04",
	"Jan 2 2006 15:04",
	"Jan 02 2006 15:04",
	"Jan 2 2006 at 15:04",
	"Jan 02 2006 at 15:04",
	"Jan 2 2006",
	"Jan 02 2006",
	"January 2 2006",
	"January 02 2006",
}

// ParseTime parses a timestamp in a number of formats
func ParseTime(args []string, loc *time.Location) (t time.Time, i int, err error) {
	t, i, err = parseTime(args, loc)
	if err != nil {
		return
	}

	if t.Year() == 0 {
		now := time.Now().In(loc)

		year := now.Year()

		// also test for month and day
		day, month := t.Day(), t.Month()
		if day == 1 && month == time.January {
			day, month = now.Day(), now.Month()
		} else {
			if month < now.Month() {
				year++
			}
		}

		t = time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	}

	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
		now := time.Now().In(loc)

		t = time.Date(t.Year(), t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), loc)
	}

	if t.Before(time.Now().In(loc)) {
		t = t.Add(24 * time.Hour)
	}

	return
}

func parseTime(args []string, loc *time.Location) (t time.Time, i int, err error) {
	for i := len(args); i > 0; i-- {
		input := strings.Join(args[:i], " ")
		input = string(input[0]) + strings.ToLower(input[1:])

		r, size := utf8.DecodeRuneInString(input)
		input = strings.ToUpper(string(r)) + strings.ToLower(input[size:])

		if strings.EqualFold(input, "tomorrow") {
			return time.Now().UTC().Add(24 * time.Hour), i - 1, nil
		}

		for _, f := range formats {
			t, err = time.ParseInLocation(f, input, loc)
			if err == nil {
				return t, i - 1, nil
			}
		}
	}
	return t, -1, err
}
