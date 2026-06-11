// Package bot wires the Discord client to the player, registers slash commands
// and routes interactions. It replaces the py-cord bot + cogs in the original.
package bot

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/disgoorg/disgo"
	dbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	dvoice "github.com/disgoorg/disgo/voice"
	davesession "github.com/thomas-vilte/dave-go/session"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/config"
	"github.com/GusPrice/dj-yosof/player"
)

// Bot holds the Discord client and shared state.
type Bot struct {
	client *dbot.Client
	cfg    *config.Config
	mgr    *player.Manager

	searchMu sync.Mutex
	searches map[string][]audio.PlayableAudio
}

// New constructs a Bot, registering the given sources with the player manager.
func New(cfg *config.Config, spotify, youtube player.Source) (*Bot, error) {
	b := &Bot{
		cfg:      cfg,
		searches: make(map[string][]audio.PlayableAudio),
	}

	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	client, err := disgo.New(cfg.DiscordToken,
		dbot.WithLogger(logger),
		dbot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildVoiceStates,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),
		dbot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagGuilds|cache.FlagVoiceStates|cache.FlagChannels),
		),
		// Enable Discord's DAVE end-to-end encryption for voice (required since
		// 2026-03). Uses the pure-Go dave-go implementation (no C++ libdave).
		dbot.WithVoiceManagerConfigOpts(
			dvoice.WithDaveSessionCreateFunc(davesession.New),
			dvoice.WithDaveSessionLogger(logger),
		),
		dbot.WithEventListenerFunc(b.onReady),
		dbot.WithEventListenerFunc(b.onSlash),
		dbot.WithEventListenerFunc(b.onComponent),
	)
	if err != nil {
		return nil, fmt.Errorf("creating discord client: %w", err)
	}

	mgr := player.NewManager(client)
	if spotify != nil {
		mgr.RegisterSource(audio.TypeSpotify, spotify)
	}
	if youtube != nil {
		mgr.RegisterSource(audio.TypeYoutube, youtube)
	}

	b.client = client
	b.mgr = mgr
	return b, nil
}

// Open connects to Discord and registers the slash commands in each configured
// guild.
func (b *Bot) Open() error {
	if err := b.client.OpenGateway(context.Background()); err != nil {
		return fmt.Errorf("opening gateway: %w", err)
	}
	return b.registerCommands()
}

// Close disconnects from Discord.
func (b *Bot) Close() error {
	b.client.Close(context.Background())
	return nil
}

func (b *Bot) onReady(_ *events.Ready) {
	log.Println("dj-yosof connected to Discord")
}

func (b *Bot) onSlash(e *events.ApplicationCommandInteractionCreate) {
	b.handleCommand(e)
}

func (b *Bot) onComponent(e *events.ComponentInteractionCreate) {
	b.handleComponent(e)
}
