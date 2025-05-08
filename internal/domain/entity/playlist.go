package entity

import "time"

type Playlist struct {
	ID          string                `json:"id"`
	UserID      string                `json:"user_id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Mood        valueobject.Mood      `json:"mood"`
	Weather     valueobject.Weather   `json:"weather,omitempty"`
	TimeOfDay   valueobject.TimeOfDay `json:"time_of_day,omitempty"`
	Tracks      []string              `json:"track_ids"`
	IsPublic    bool                  `json:"is_public"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

func NewPlaylist(userID, name, description string, mood valueobject.Mood) *Playlist {
	return &Playlist{
		UserID:      userID,
		Name:        name,
		Description: description,
		Mood:        mood,
		Tracks:      []string{},
		IsPublic:    false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (p *Playlist) AddTrack(trackID string) {
	p.Tracks = append(p.Tracks, trackID)
	p.UpdatedAt = time.Now()
}

func (p *Playlist) RemoveTrack(trackID string) {
	for i, id := range p.Tracks {
		if id == trackID {
			p.Tracks = append(p.Tracks[:i], p.Tracks[i+1:]...)
		}
	}
	p.UpdatedAt = time.Now()
}
