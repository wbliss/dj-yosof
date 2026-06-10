// Package bot wires the Discord session to the player, registers slash commands
// and routes interactions. It replaces the py-cord bot + cogs in the original.
package bot

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/config"
	"github.com/GusPrice/dj-yosof/player"
)

// Bot holds the Discord session and shared state.
type Bot struct {
	session *discordgo.Session
	cfg     *config.Config
	mgr     *player.Manager

	searchMu sync.Mutex
	searches map[string][]audio.PlayableAudio
}

// New constructs a Bot, registering the given sources with the player manager.
func New(cfg *config.Config, spotify, youtube player.Source) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("creating discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	mgr := player.NewManager(session)
	if spotify != nil {
		mgr.RegisterSource(audio.TypeSpotify, spotify)
	}
	if youtube != nil {
		mgr.RegisterSource(audio.TypeYoutube, youtube)
	}

	b := &Bot{
		session:  session,
		cfg:      cfg,
		mgr:      mgr,
		searches: make(map[string][]audio.PlayableAudio),
	}

	session.AddHandler(b.onReady)
	session.AddHandler(b.onInteraction)
	return b, nil
}

// Open connects to Discord and registers the slash commands in each configured
// guild.
func (b *Bot) Open() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening discord session: %w", err)
	}
	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("registering commands: %w", err)
	}
	return nil
}

// Close disconnects from Discord.
func (b *Bot) Close() error {
	return b.session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, _ *discordgo.Ready) {
	log.Printf("Logged in as %s#%s", s.State.User.Username, s.State.User.Discriminator)
}

// onInteraction routes slash commands and component (button) interactions.
func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	}
}
