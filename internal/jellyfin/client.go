package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	ServerURL  string
	Token      string
	UserID     string
	HTTPClient *http.Client
}

func NewClient(serverURL, token, userID string) *Client {
	return &Client{
		ServerURL:  serverURL,
		Token:      token,
		UserID:     userID,
		HTTPClient: &http.Client{},
	}
}

type AuthResponse struct {
	User struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	} `json:"User"`
	AccessToken string `json:"AccessToken"`
}

func (c *Client) Authenticate(username, password string) (*AuthResponse, error) {
	endpoint := fmt.Sprintf("%s/Users/AuthenticateByName", c.ServerURL)
	payload := map[string]string{
		"Username": username,
		"Pw":       password,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Emby-Authorization", "MediaBrowser Client=\"Jellyfin-TUI\", Device=\"Terminal\", DeviceId=\"TODO-ID\", Version=\"0.0.1\"")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed: %s - %s", resp.Status, string(bodyBytes))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}

	c.Token = authResp.AccessToken
	c.UserID = authResp.User.ID
	return &authResp, nil
}

type Item struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
	Type string `json:"Type"`
}

type ItemsResponse struct {
	Items []Item `json:"Items"`
}

func (c *Client) GetViews() ([]Item, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Users/%s/Views", c.ServerURL, c.UserID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get views: %s", resp.Status)
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetItems(parentID string) ([]Item, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", c.ServerURL, c.UserID, parentID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get items: %s", resp.Status)
	}

	var itemsResp ItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetImageURL(itemID string) string {
	return fmt.Sprintf("%s/Items/%s/Images/Primary", c.ServerURL, itemID)
}

type MusicItem struct {
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Type         string `json:"Type"`
	AlbumArtist  string `json:"AlbumArtist"`
	Album        string `json:"Album"`
	RunTimeTicks int64  `json:"RunTimeTicks"`
	IndexNumber  int    `json:"IndexNumber"`
}

type MusicItemsResponse struct {
	Items []MusicItem `json:"Items"`
}

func (c *Client) GetArtists() ([]MusicItem, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Artists?UserId=%s&SortBy=SortName&SortOrder=Ascending", c.ServerURL, c.UserID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get artists: %s", resp.Status)
	}

	var itemsResp MusicItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetAlbums(artistID string) ([]MusicItem, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Users/%s/Items?ArtistIds=%s&IncludeItemTypes=MusicAlbum&Recursive=true&SortBy=SortName",
		c.ServerURL, c.UserID, artistID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get albums: %s", resp.Status)
	}

	var itemsResp MusicItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetTracksByArtist(artistID string) ([]MusicItem, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Users/%s/Items?ArtistIds=%s&IncludeItemTypes=Audio&Recursive=true&SortBy=Album,IndexNumber",
		c.ServerURL, c.UserID, artistID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tracks: %s", resp.Status)
	}

	var itemsResp MusicItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetTracks(albumID string) ([]MusicItem, error) {
	if c.Token == "" || c.UserID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s&IncludeItemTypes=Audio&SortBy=IndexNumber",
		c.ServerURL, c.UserID, albumID)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tracks: %s", resp.Status)
	}

	var itemsResp MusicItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, err
	}

	return itemsResp.Items, nil
}

func (c *Client) GetAudioStreamURL(itemID string) string {
	return fmt.Sprintf("%s/Audio/%s/universal?UserId=%s&api_key=%s&Container=mp3&AudioCodec=mp3",
		c.ServerURL, itemID, c.UserID, c.Token)
}

func (c *Client) addHeaders(req *http.Request) {
	req.Header.Set("X-Emby-Token", c.Token)
	req.Header.Set("X-Emby-Authorization", "MediaBrowser Client=\"Jellyfin-TUI\", Device=\"Terminal\", DeviceId=\"TODO-ID\", Version=\"0.0.1\", Token=\""+c.Token+"\"")
}
