package valueobject

type Mood string

const (
	MoodHappy      Mood = "happy"
	MoodSad        Mood = "sad"
	MoodEnergetic  Mood = "energetic"
	MoodCalm       Mood = "calm"
	MoodFocused    Mood = "focused"
	MoodRomantic   Mood = "romantic"
	MoodNostalgic  Mood = "nostalgic"
	MoodParty      Mood = "party"
	MoodMelancholy Mood = "melancholy"
)

func ValidMood(mood Mood) bool {
	validMoods := []Mood{
		MoodHappy, MoodSad, MoodEnergetic, MoodCalm,
		MoodFocused, MoodRomantic, MoodNostalgic, MoodParty,
		MoodMelancholy,
	}
	for _, m := range validMoods {
		if mood == m {
			return true
		}
	}
	return false
}

func AllMoods() []Mood {
	return []Mood{
		MoodHappy, MoodSad, MoodEnergetic, MoodCalm,
		MoodFocused, MoodRomantic, MoodNostalgic, MoodParty,
		MoodMelancholy,
	}
}
