package valueObject

type Weather string

const (
	WeatherSunny  Weather = "sunny"
	WeatherCloudy Weather = "cloudy"
	WeatherRainy  Weather = "rainy"
	WeatherStormy Weather = "stormy"
	WeatherSnowy  Weather = "snowy"
	WeatherFoggy  Weather = "foggy"
	WeatherWindy  Weather = "windy"
	WeatherHot    Weather = "hot"
	WeatherCold   Weather = "cold"
)

func ValidWeather(weather Weather) bool {
	validWeather := []Weather{
		WeatherSunny, WeatherCloudy, WeatherRainy,
		WeatherStormy, WeatherSnowy, WeatherFoggy,
		WeatherWindy, WeatherHot, WeatherCold,
	}
	for _, w := range validWeather {
		if w == weather {
			return true
		}
	}
	return false
}

func MapFromOpenWeather(code int) Weather {
	switch {
	case code >= 200 && code < 300:
		return WeatherStormy // Грозы
	case code >= 300 && code < 400:
		return WeatherRainy // Моросящий дождь
	case code >= 500 && code < 600:
		return WeatherRainy // Дождь
	case code >= 600 && code < 700:
		return WeatherSnowy // Снег
	case code >= 700 && code < 800:
		return WeatherFoggy // Туман и другие явления
	case code == 800:
		return WeatherSunny // Ясно
	case code > 800:
		return WeatherCloudy // Облачно
	default:
		return WeatherSunny
	}
}
