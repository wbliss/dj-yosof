// Command dj-yosof is a Discord music bot that plays audio from Spotify and
// YouTube. It is a Go rewrite of the original Python project.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GusPrice/dj-yosof/bot"
	"github.com/GusPrice/dj-yosof/config"
	"github.com/GusPrice/dj-yosof/player"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()

	// Spotify is optional: if authentication fails (e.g. no credentials yet), the
	// bot still runs with YouTube. On first run an interactive OAuth2 login link
	// is printed to the console.
	var spotify player.Source
	if sp, err := player.NewSpotifySource(ctx, cfg.SpotifyCredentialsFile, cfg.SpotifyOAuthCallbackPort); err != nil {
		log.Printf("spotify disabled: %v", err)
	} else {
		spotify = sp
		defer sp.Close()
	}

	youtube := player.NewYoutubeSource()

	b, err := bot.New(cfg, spotify, youtube)
	if err != nil {
		log.Fatalf("bot: %v", err)
	}

	if err := b.Open(); err != nil {
		log.Fatalf("bot: %v", err)
	}
	defer func() { _ = b.Close() }()

	log.Println("dj-yosof is running. Press Ctrl+C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
}
