package internal

import "time"

const (
	formatDDMMYYYY = "02.01.2006"
)

func Format(date time.Time) string {
	return date.Format(formatDDMMYYYY)
}
