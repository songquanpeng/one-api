package sigv4

import (
	gotime "time"
)

// Time wraps time.Time to cache its string format result.
type Time struct {
	gotime.Time
	short string
	long  string
}

// NewTime creates a new signingTime with the specified time.Time.
func NewTime(t gotime.Time) Time {
	return Time{Time: t.UTC()}
}

// TimeFormat provides a time formatted in the X-Amz-Date format.
func (m *Time) TimeFormat() string {
	return m.readOrFormat(&m.long, TimeFormat)
}

// ShortTimeFormat provides a time formatted in short time format.
func (m *Time) ShortTimeFormat() string {
	return m.readOrFormat(&m.short, ShortTimeFormat)
}

func (m *Time) readOrFormat(target *string, format string) string {
	if len(*target) > 0 {
		return *target
	}
	v := m.Time.Format(format)
	*target = v
	return v
}
