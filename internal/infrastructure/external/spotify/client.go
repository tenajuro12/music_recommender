package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"strings"
	"sync"
	"time"
)

const (
	tokenURL          = "https://accounts.spotify.com/api/token"
	apiURL            = "https://api.spotify.com/v1"
	recommendationURL = apiURL + "/recommendations"
	trackURL          = apiURL + "/tracks"
	audioFeaturesURL  = apiURL + "/audio-features"
	searchURL         = apiURL + "/search"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Client struct {
	config       Config
	httpClient   *http.Client
	accessToken  string
	tokenExpiry  time.Time
	refreshToken string
	mutex        sync.Mutex
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SetRefreshToken(refreshToken string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.refreshToken = refreshToken
}

func (c *Client) GetAuthURL(state string, scopes []string) string {
	params := url.Values{}
	params.Add("client_id", c.config.ClientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", c.config.RedirectURI)
	params.Add("state", state)
	params.Add("scope", strings.Join(scopes, " "))

	return "https://accounts.spotify.com/authorize?" + params.Encode()
}

func (c *Client) ExchangeCodeForToken(ctx context.Context, code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.config.RedirectURI)

	return c.requestToken(ctx, data)
}

func (c *Client) refreshAccessToken(ctx context.Context) error {
	if c.refreshToken == "" {
		return errors.New("refresh token is not set")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.refreshToken)

	return c.requestToken(ctx, data)
}

func (c *Client) requestToken(ctx context.Context, data url.Values) error {

	auth := base64.StdEncoding.EncodeToString([]byte(c.config.ClientID + ":" + c.config.ClientSecret))

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.accessToken = tokenResponse.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	if tokenResponse.RefreshToken != "" {
		c.refreshToken = tokenResponse.RefreshToken
	}

	return nil
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	c.mutex.Lock()
	isValid := c.accessToken != "" && time.Now().Before(c.tokenExpiry)
	c.mutex.Unlock()

	if !isValid {
		return c.refreshAccessToken(ctx)
	}

	return nil
}

func (c *Client) makeRequest(ctx context.Context, method, url string, body io.Reader, result interface{}) error {
	if err := c.ensureValidToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}

	c.mutex.Lock()
	req.Header.Add("Authorization", "Bearer "+c.accessToken)
	c.mutex.Unlock()

	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Spotify API error: %d %s", resp.StatusCode, string(errorBody))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

func (c *Client) GetTrack(ctx context.Context, trackID string) (*entity.Track, error) {
	var response struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  string `json:"preview_url"`
		ExternalURL struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Album struct {
			Name   string `json:"name"`
			Images []struct {
				URL    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
			ReleaseDate string `json:"release_date"`
		} `json:"album"`
		Artists []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artists"`
	}

	url := fmt.Sprintf("%s/%s", trackURL, trackID)
	if err := c.makeRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, err
	}

	releaseDate, err := time.Parse("2006-01-02", response.Album.ReleaseDate)
	if err != nil {

		releaseDate = time.Now()
	}

	audioFeatures, err := c.GetAudioFeatures(ctx, trackID)
	if err != nil {

		audioFeatures = valueObject.AudioFeatures{}
	}

	var artistName string
	if len(response.Artists) > 0 {
		artistName = response.Artists[0].Name
	}

	var imageURL string
	if len(response.Album.Images) > 0 {
		imageURL = response.Album.Images[0].URL
	}
	track := entity.NewTrack(
		response.ID,
		response.Name,
		artistName,
		response.Album.Name,
		releaseDate,
		response.Popularity,
		audioFeatures,
		response.PreviewURL,
		imageURL,
	)

	return track, nil
}

func (c *Client) GetAudioFeatures(ctx context.Context, trackID string) (valueObject.AudioFeatures, error) {
	var response struct {
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

	url := fmt.Sprintf("%s/%s", audioFeaturesURL, trackID)
	if err := c.makeRequest(ctx, "GET", url, nil, &response); err != nil {
		return valueObject.AudioFeatures{}, err
	}

	return valueObject.AudioFeatures{
		Danceability:     response.Danceability,
		Energy:           response.Energy,
		Key:              response.Key,
		Loudness:         response.Loudness,
		Mode:             response.Mode,
		Speechiness:      response.Speechiness,
		Acousticness:     response.Acousticness,
		Instrumentalness: response.Instrumentalness,
		Liveness:         response.Liveness,
		Valence:          response.Valence,
		Tempo:            response.Tempo,
		Duration:         response.Duration,
		TimeSignature:    response.TimeSignature,
	}, nil
}

func (c *Client) SearchTracks(ctx context.Context, query string, limit int) ([]*entity.Track, error) {
	var response struct {
		Tracks struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Popularity  int    `json:"popularity"`
				PreviewURL  string `json:"preview_url"`
				ExternalURL struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Album struct {
					Name   string `json:"name"`
					Images []struct {
						URL    string `json:"url"`
						Height int    `json:"height"`
						Width  int    `json:"width"`
					} `json:"images"`
					ReleaseDate string `json:"release_date"`
				} `json:"album"`
				Artists []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"items"`
		} `json:"tracks"`
	}

	params := url.Values{}
	params.Add("q", query)
	params.Add("type", "track")
	params.Add("limit", fmt.Sprintf("%d", limit))

	url := fmt.Sprintf("%s?%s", searchURL, params.Encode())
	if err := c.makeRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(response.Tracks.Items))

	for _, item := range response.Tracks.Items {
		releaseDate, err := time.Parse("2006-01-02", item.Album.ReleaseDate)
		if err != nil {
			releaseDate = time.Now()
		}

		var artistName string
		if len(item.Artists) > 0 {
			artistName = item.Artists[0].Name
		}

		var imageURL string
		if len(item.Album.Images) > 0 {
			imageURL = item.Album.Images[0].URL
		}

		audioFeatures, err := c.GetAudioFeatures(ctx, item.ID)
		if err != nil {
			audioFeatures = valueObject.AudioFeatures{}
		}

		track := entity.NewTrack(
			item.ID,
			item.Name,
			artistName,
			item.Album.Name,
			releaseDate,
			item.Popularity,
			audioFeatures,
			item.PreviewURL,
			imageURL,
		)

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (c *Client) GetRecommendations(ctx context.Context, params map[string]string, limit int) ([]*entity.Track, error) {
	var response struct {
		Tracks []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Popularity  int    `json:"popularity"`
			PreviewURL  string `json:"preview_url"`
			ExternalURL struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Album struct {
				Name   string `json:"name"`
				Images []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"images"`
				ReleaseDate string `json:"release_date"`
			} `json:"album"`
			Artists []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"tracks"`
	}

	queryParams := url.Values{}
	queryParams.Add("limit", fmt.Sprintf("%d", limit))

	for key, value := range params {
		queryParams.Add(key, value)
	}

	url := fmt.Sprintf("%s?%s", recommendationURL, queryParams.Encode())
	if err := c.makeRequest(ctx, "GET", url, nil, &response); err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(response.Tracks))

	for _, item := range response.Tracks {
		releaseDate, err := time.Parse("2006-01-02", item.Album.ReleaseDate)
		if err != nil {
			releaseDate = time.Now()
		}

		var artistName string
		if len(item.Artists) > 0 {
			artistName = item.Artists[0].Name
		}

		var imageURL string
		if len(item.Album.Images) > 0 {
			imageURL = item.Album.Images[0].URL
		}

		audioFeatures, err := c.GetAudioFeatures(ctx, item.ID)
		if err != nil {
			audioFeatures = valueObject.AudioFeatures{}
		}

		track := entity.NewTrack(
			item.ID,
			item.Name,
			artistName,
			item.Album.Name,
			releaseDate,
			item.Popularity,
			audioFeatures,
			item.PreviewURL,
			imageURL,
		)

		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (c *Client) GetRecommendationsByMood(ctx context.Context, mood valueObject.Mood, limit int) ([]*entity.Track, error) {
	params := make(map[string]string)

	switch mood {
	case valueObject.MoodHappy:
		params["target_valence"] = "0.8"
		params["target_energy"] = "0.7"
		params["min_valence"] = "0.7"
	case valueObject.MoodSad:
		params["target_valence"] = "0.3"
		params["target_energy"] = "0.4"
		params["max_valence"] = "0.4"
	case valueObject.MoodEnergetic:
		params["target_energy"] = "0.9"
		params["min_tempo"] = "120"
		params["min_energy"] = "0.8"
	case valueObject.MoodCalm:
		params["target_energy"] = "0.3"
		params["target_acousticness"] = "0.7"
		params["max_energy"] = "0.4"
	case valueObject.MoodFocused:
		params["target_instrumentalness"] = "0.7"
		params["target_energy"] = "0.5"
		params["min_instrumentalness"] = "0.5"
	case valueObject.MoodRomantic:
		params["target_valence"] = "0.6"
		params["target_energy"] = "0.5"
		params["target_acousticness"] = "0.5"
	case valueObject.MoodNostalgic:
		params["target_valence"] = "0.5"
		params["target_acousticness"] = "0.6"
	case valueObject.MoodParty:
		params["target_danceability"] = "0.8"
		params["target_energy"] = "0.8"
		params["min_danceability"] = "0.7"
	case valueObject.MoodMelancholy:
		params["target_valence"] = "0.3"
		params["target_energy"] = "0.4"
		params["target_acousticness"] = "0.6"
	default:
		params["target_valence"] = "0.5"
		params["target_energy"] = "0.5"
	}
	params["seed_genres"] = "pop,rock,electronic,classical,hip-hop"

	return c.GetRecommendations(ctx, params, limit)
}
