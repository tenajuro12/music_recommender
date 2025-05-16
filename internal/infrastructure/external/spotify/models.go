package spotify

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

const (
	ScopeUserReadPrivate          = "user-read-private"
	ScopeUserReadEmail            = "user-read-email"
	ScopeUserReadPlaybackState    = "user-read-playback-state"
	ScopeUserModifyPlaybackState  = "user-modify-playback-state"
	ScopeUserReadCurrentlyPlaying = "user-read-currently-playing"
	ScopePlaylistReadPrivate      = "playlist-read-private"
	ScopePlaylistModifyPrivate    = "playlist-modify-private"
	ScopePlaylistModifyPublic     = "playlist-modify-public"
	ScopeUserLibraryRead          = "user-library-read"
	ScopeUserLibraryModify        = "user-library-modify"
	ScopeUserTopRead              = "user-top-read"
	ScopeUserReadRecentlyPlayed   = "user-read-recently-played"
)

func CommonScopes() []string {
	return []string{
		ScopeUserReadPrivate,
		ScopeUserReadEmail,
		ScopePlaylistReadPrivate,
		ScopePlaylistModifyPrivate,
		ScopePlaylistModifyPublic,
		ScopeUserLibraryRead,
		ScopeUserTopRead,
	}
}
