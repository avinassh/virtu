package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate)
	ch    = make(chan *oauth2.Token)
	state = getRandomString(20)
)

func main() {
	config := readConfig()
	if config.AccessToken == "" || config.RefreshToken == "" {
		fmt.Println("No Config")
		initOAuth()
		// read config again
		config = readConfig()
	} else {
		fmt.Println("From Config")
	}
	client := auth.NewClient(&oauth2.Token{
		AccessToken:  config.AccessToken,
		RefreshToken: config.RefreshToken,
		Expiry:       time.Unix(config.TokenExpiry, 0),
	})
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal("Shit :", err)
	}
	fmt.Println("Hi: ", user.ID)
}

func initOAuth() {
	config := readConfig()
	startServer()
	auth.SetAuthInfo(config.ClientID, config.ClientSecret)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	token := <-ch
	fmt.Println(token.Expiry.Unix())
	client := auth.NewClient(token)
	_, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	config.AccessToken = token.AccessToken
	config.RefreshToken = token.RefreshToken
	config.TokenExpiry = token.Expiry.Unix()
	writeConfig(config)
}

func startServer() *http.Server {
	srv := &http.Server{Addr: ":8080"}
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
	fmt.Fprintf(w, "Login Completed!")
	// pass the token to channel
	ch <- token
}

// from https://siongui.github.io/2015/04/13/go-generate-random-string/
func getRandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
