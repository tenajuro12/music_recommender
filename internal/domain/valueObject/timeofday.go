package valueObject

import "time"

type TimeOfDay string

const (
	TimeOfDayMorning   TimeOfDay = "morning"
	TimeOfDayAfternoon TimeOfDay = "afternoon"
	TimeOfDayEvening   TimeOfDay = "evening"
	TimeOfDayNight     TimeOfDay = "night"
)

func GetCurrentTimeOfToday() TimeOfDay {
	hour := time.Now().Hour()

	switch {
	case hour >= 5 && hour < 12:
		return TimeOfDayMorning
	case hour >= 12 && hour < 17:
		return TimeOfDayAfternoon
	case hour >= 17 && hour < 22:
		return TimeOfDayEvening
	default:
		return TimeOfDayNight
	}
}

func AllTimesOfDay() []TimeOfDay {
	return []TimeOfDay{
		TimeOfDayMorning, TimeOfDayAfternoon,
		TimeOfDayEvening, TimeOfDayNight,
	}
}
