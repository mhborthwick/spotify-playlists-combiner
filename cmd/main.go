package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/mhborthwick/spotify-playlist-compiler/config"
)

type Track struct {
	URI string `json:"uri"`
}

type GetPlaylistItemsResponseBody struct {
	Items []struct {
		Track Track `json:"track"`
	} `json:"items"`
}

type CreatePlaylistRequestBody struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

func GetSpotifyURIs(body []byte) ([]string, error) {
	var parsed GetPlaylistItemsResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	uris := make([]string, len(parsed.Items))
	for i, item := range parsed.Items {
		uris[i] = item.Track.URI
	}
	return uris, nil
}

func GetSpotifyId(playlist string) (string, error) {
	re := regexp.MustCompile(`[a-zA-Z0-9]{22}`)
	id := re.FindString(playlist)
	if id == "" {
		return "", errors.New("Invalid playlist")
	}
	return id, nil
}

func GetSpotifyPlaylistItems(cfg *config.Config, id string, client *http.Client) ([]byte, error) {
	url := "https://api.spotify.com/v1/playlists/" + id + "/tracks"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	token := "Bearer " + cfg.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func CreateSpotifyPlaylist(cfg *config.Config, client *http.Client) ([]byte, error) {
	url := "https://api.spotify.com/v1/users/" + cfg.UserID + "/playlists"
	requestData := CreatePlaylistRequestBody{
		Name:        "New Playlist",
		Description: "Created by Playlist Compiler",
		Public:      false,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	token := "Bearer " + cfg.Token
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func main() {
	cfg, err := config.LoadFromPath(context.Background(), "pkl/local/config.pkl")
	if err != nil {
		panic(err)
	}

	playlist := "https://open.spotify.com/playlist/4KMuVswhHsgHusA1hSdZQ5?si=a4b8123f214d470c"

	id, err := GetSpotifyId(playlist)

	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	client := &http.Client{}

	body, err := GetSpotifyPlaylistItems(cfg, id, client)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	uris, err := GetSpotifyURIs(body)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(uris)

	// now that we have uris, create a new blank playlist
	created, err := CreateSpotifyPlaylist(cfg, client)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(created)
}
