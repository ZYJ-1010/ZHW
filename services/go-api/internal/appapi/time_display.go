package appapi

import "time"

var appDisplayLocation = loadAppDisplayLocation()

func loadAppDisplayLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	return loc
}

func appDisplayTime(value time.Time) time.Time {
	if value.IsZero() {
		return value
	}
	return value.In(appDisplayLocation)
}

func formatAppDisplayTime(value time.Time, layout string) string {
	if value.IsZero() {
		return ""
	}
	return appDisplayTime(value).Format(layout)
}
