package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://localhost:8080/callback"
	port        = 8080
	defaultLimit = 50
	defaultOffset = 0
)

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate)
	ch    = make(chan *oauth2.Token)
	state = getRandomString(20)
)

func main() {
	client := getSpotifyClient()
	disoverWeekly := currentDiscoverWeekly(client)
	fmt.Println(disoverWeekly)
}

func getSpotifyClient() *spotify.Client {
	config := readConfig()
	if config.AccessToken == "" || config.RefreshToken == "" {
		initOAuth()
		// read config again
		config = readConfig()
	}
	token := &oauth2.Token{
		Expiry:       time.Unix(config.TokenExpiry, 0),
		TokenType:    config.TokenType,
		AccessToken:  config.AccessToken,
		RefreshToken: config.RefreshToken,
	}
	auth.SetAuthInfo(config.ClientID, config.ClientSecret)
	client := auth.NewClient(token)
	_, err := client.CurrentUser()
	if err != nil {
		log.Fatal("Failed :", err)
	}
	return &client
}

func initOAuth() {
	startServer()
	config := readConfig()
	auth.SetAuthInfo(config.ClientID, config.ClientSecret)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	token := <-ch
	updateConfig(token)
}

func startServer() *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("Error with the server", err)
		}
	}()
	return srv
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	fmt.Fprint(w, "Login Completed!")
	// pass the token to channel
	ch <- token
}

// Returns current Discover Weekly playlist of the user
//
// A user may not have a Discover Weekly playlist at all, user can 'unfollow'
// the playlist and the playlist will be gone
//
// Similarly a user may have more than two such playlists, as user can follow
// someone else's playlist
//
// Also, the playlist returned may not be the user's at all. A user may unfollow
// his own playlist and follow someone else's. And I give a damn.
func currentDiscoverWeekly(c *spotify.Client) *spotify.SimplePlaylist {
	allPlaylists, err := getAllPlaylists(c)
	if err != nil {
		log.Fatal("Failed to fetch playlists from Spotify", err)
	}
	var discoverWeekly *spotify.SimplePlaylist
	count := 0
	for _, item := range allPlaylists.Playlists {
		if item.Name == "Discover Weekly" && item.Owner.ID == "spotify" {
			discoverWeekly = &item
			count += 1
		}
	}
	if count == 0 {
		log.Fatal("User does not have a Discover Weekly playlist")
	}
	if count > 1 {
		log.Fatal("User has more than one Discover Weekly playlist")
	}
	return discoverWeekly
}

// Returns all the playlists of the current user
func getAllPlaylists(c *spotify.Client) (*spotify.SimplePlaylistPage, error){
	var allPlaylists *spotify.SimplePlaylistPage
	var total int
	limit := defaultLimit
	offset := defaultOffset
	opt := spotify.Options{
		Limit:  &limit,
		Offset: &offset,
	}
	for {
		playlists, err := c.CurrentUsersPlaylistsOpt(&opt)
		if err != nil {
			return nil, err
		}
		total = playlists.Total
		if allPlaylists == nil {
			allPlaylists = playlists
		} else {
			allPlaylists.Playlists = append(allPlaylists.Playlists, playlists.Playlists...)
		}
		offset = offset + limit
		if total < offset {
			break
		}
	}
	return allPlaylists, nil
}