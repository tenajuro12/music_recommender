package valueObject

type AudioFeatures struct {
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Key              int     `json:"key"`
	Loudness         float64 `json:"loudness"`
	Mode             int     `json:"mode"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:"tempo"`
	Duration         int     `json:"duration_ms"`
	TimeSignature    int     `json:"time_signature"`
}

func (af *AudioFeatures) MatchesMood(mood Mood) bool {
	switch mood {
	case MoodHappy:
		return af.Valence > 0.7 && af.Energy > 0.5
	case MoodSad:
		return af.Valence < 0.4 && af.Energy < 0.5
	case MoodEnergetic:
		return af.Energy > 0.8 && af.Tempo > 120
	case MoodCalm:
		return af.Energy < 0.4 && af.Acousticness > 0.5
	case MoodFocused:
		return af.Instrumentalness > 0.5 && af.Energy < 0.7
	case MoodRomantic:
		return af.Valence > 0.5 && af.Energy < 0.6 && af.Acousticness > 0.4
	case MoodNostalgic:
		return af.Valence > 0.3 && af.Valence < 0.7 && af.Acousticness > 0.4
	case MoodParty:
		return af.Danceability > 0.7 && af.Energy > 0.7
	case MoodMelancholy:
		return af.Valence < 0.4 && af.Energy < 0.5 && af.Acousticness > 0.5
	default:
		return true
	}
}

func (af *AudioFeatures) MatchesWeather(weather Weather) bool {
	switch weather {
	case WeatherSunny:
		return af.Valence > 0.6 && af.Energy > 0.5
	case WeatherRainy:
		return af.Valence < 0.5 && af.Acousticness > 0.5
	case WeatherStormy:
		return af.Energy > 0.7 && af.Loudness > -8.0
	case WeatherSnowy:
		return af.Acousticness > 0.6 && af.Energy < 0.5
	case WeatherCloudy:
		return af.Valence > 0.3 && af.Valence < 0.7
	case WeatherFoggy:
		return af.Acousticness > 0.5 && af.Instrumentalness > 0.3
	case WeatherWindy:
		return af.Energy > 0.6 && af.Acousticness < 0.4
	case WeatherHot:
		return af.Energy > 0.5 && af.Danceability > 0.6
	case WeatherCold:
		return af.Energy < 0.6 && af.Acousticness > 0.4
	default:
		return true
	}
}

func (af *AudioFeatures) MatchesTimeOfDay(time TimeOfDay) bool {
	switch time {
	case TimeOfDayMorning:
		return af.Valence > 0.5 && af.Energy > 0.5 && af.Energy < 0.8
	case TimeOfDayAfternoon:
		return af.Energy > 0.5 && af.Danceability > 0.5
	case TimeOfDayEvening:
		return af.Energy > 0.3 && af.Energy < 0.8
	case TimeOfDayNight:
		return (af.Energy < 0.5 && af.Acousticness > 0.5) ||
			(af.Energy > 0.8 && af.Danceability > 0.7) // либо спокойная, либо для вечеринок
	default:
		return true
	}
}
