package virtu

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/zmb3/spotify"
)

const redirectURI = "http://localhost:8080/callback"

var (
	auth = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopePlaylistReadPrivate,
		spotify.ScopePlaylistModifyPrivate)
	ch    = make(chan *spotify.Client)
	state = getRandomString(20)
)

func main() {
	config := readConfig()
	if config.AccessToken == "" || config.RefreshToken == "" {
		initOAuth()
	} else {
		fmt.Println("Nope")
	}
}

func initOAuth() {
	config := readConfig()
	startServer()
	auth.SetAuthInfo(config.ClientID, config.ClientSecret)
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	token := <-ch
	client := auth.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.ID)
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
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
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
