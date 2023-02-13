package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://localhost:8080/callback"
	mujiURI     = "spotify:artist:67J1KP70RqsL6uAIXkscKE"
)

var (
	auth  = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPlaybackState, spotifyauth.ScopeUserModifyPlaybackState))
	ch    = make(chan *oauth2.Token)
	state = "abc123"
)

func mujifyDir() (string, error) {
	confDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	mDir := filepath.Join(confDir, "mujify")
	if err := os.MkdirAll(mDir, 0755); err != nil {
		return "", err
	}
	return mDir, nil
}

func tokenPath() (string, error) {
	mDir, err := mujifyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(mDir, ".token_cache"), nil
}

func confPath() (string, error) {
	mDir, err := mujifyDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(mDir, ".config.yml"), nil
}

func fetchToken(ctx context.Context) (*oauth2.Token, error) {
	var token *oauth2.Token

	tPath, err := tokenPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(tPath); err != nil {
		if token, err = authenticate(ctx); err != nil {
			return nil, err
		}
	} else {
		tokenFile, err := os.Open(tPath)
		if err != nil {
			return nil, err
		}
		token = new(oauth2.Token)
		json.NewDecoder(tokenFile).Decode(token)
	}
	return token, nil
}

func updateToken(client *spotify.Client) error {
	token, err := client.Token()
	if err != nil {
		return fmt.Errorf("can't refresh: %w\n", err)
	}
	tPath, err := tokenPath()
	if err != nil {
		return err
	}
	tFile, err := os.Create(tPath)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(tFile).Encode(token); err != nil {
		return err
	}
	return nil
}

func newClient(ctx context.Context) (*spotify.Client, error) {
	token, err := fetchToken(ctx)
	if err != nil {
		return nil, err
	}
	client := spotify.New(auth.Client(ctx, token))
	if err := updateToken(client); err != nil {
		return nil, err
	}
	return client, nil
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	ch <- token
	fmt.Fprintf(w, "Login Completed!")
}

func authenticate(ctx context.Context) (*oauth2.Token, error) {
	http.HandleFunc("/callback", authHandler)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
	token := <-ch
	return token, nil
}

func play(ctx *cli.Context) error {
	client, err := newClient(ctx.Context)
	if err != nil {
		return err
	}
	u := spotify.URI(mujiURI)
	opt := &spotify.PlayOptions{
		PlaybackContext: &u,
	}
	if err = client.PlayOpt(ctx.Context, opt); err != nil {
		return err
	}
	return nil
}

func main() {
	app := &cli.App{
		Name:   "mujify",
		Usage:  "Play MUJI BGM at random",
		Action: play,
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
