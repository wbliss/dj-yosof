package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/util"
	"github.com/GusPrice/dj-yosof/views"
	"github.com/GusPrice/dj-yosof/voice"
)

// commands defines the slash commands, mirroring the Python cogs.
var commands = []*discordgo.ApplicationCommand{
	{Name: "hello", Description: "Say hello"},
	{Name: "join", Description: "Join your voice channel"},
	{Name: "leave", Description: "Leave the voice channel"},
	{
		Name:        "spotify",
		Description: "Play a Spotify track/album/playlist link, or search Spotify",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "Link or search terms", Required: true},
		},
	},
	{
		Name:        "yt",
		Description: "Play a YouTube link, or search YouTube",
		Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "Link or search terms", Required: true},
		},
	},
	{Name: "pause", Description: "Pause playback"},
	{Name: "queue", Description: "Show the queue"},
	{Name: "skip", Description: "Skip the current track"},
	{Name: "stop", Description: "Stop playback and clear the queue"},
}

func (b *Bot) registerCommands() error {
	appID := b.session.State.User.ID
	for _, guildID := range b.cfg.GuildIDs {
		for _, cmd := range commands {
			if _, err := b.session.ApplicationCommandCreate(appID, guildID, cmd); err != nil {
				return fmt.Errorf("creating %q in guild %s: %w", cmd.Name, guildID, err)
			}
		}
	}
	return nil
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "hello":
		respond(s, i, fmt.Sprintf("Hi, <@%s>", i.Member.User.ID))
	case "join":
		b.cmdJoin(s, i)
	case "leave":
		b.cmdLeave(s, i)
	case "spotify":
		b.cmdPlay(s, i, audio.TypeSpotify, util.IsSpotifyLink)
	case "yt":
		b.cmdPlay(s, i, audio.TypeYoutube, util.IsYoutubeLink)
	case "pause":
		respond(s, i, "Pause is not implemented yet")
	case "queue":
		b.cmdQueue(s, i)
	case "skip":
		b.mgr.Player(i.GuildID).Skip()
		respond(s, i, "Skipped")
	case "stop":
		b.mgr.Player(i.GuildID).Stop()
		respond(s, i, "Stopped")
	}
}

func (b *Bot) cmdJoin(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if _, err := voice.ConnectOrMove(s, i.GuildID, i.Member.User.ID); err != nil {
		respond(s, i, err.Error())
		return
	}
	respond(s, i, "Joined your voice channel")
}

func (b *Bot) cmdLeave(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if err := voice.Leave(s, i.GuildID); err != nil {
		respond(s, i, fmt.Sprintf("Failed to leave: %v", err))
		return
	}
	respond(s, i, "Left the voice channel")
}

// cmdPlay handles both /spotify and /yt: link → enqueue immediately; otherwise
// search and present selection buttons. Ports SpotifyCog.spotify / YoutubeCog.yt.
func (b *Bot) cmdPlay(s *discordgo.Session, i *discordgo.InteractionCreate, srcType audio.Type, isLink func(string) bool) {
	query := i.ApplicationCommandData().Options[0].StringValue()
	src := b.mgr.Source(srcType)
	if src == nil {
		respond(s, i, fmt.Sprintf("%s is not available", srcType))
		return
	}

	deferResponse(s, i)
	ctx := context.Background()

	if isLink(query) {
		tracks, err := src.OpenLink(ctx, query)
		if err != nil {
			followup(s, i, fmt.Sprintf("Couldn't load that link: %v", err))
			return
		}
		if len(tracks) == 0 {
			followup(s, i, "That link had no playable tracks")
			return
		}
		if !b.enqueueAll(s, i, tracks) {
			return
		}
		followup(s, i, fmt.Sprintf("Added %d track(s) to the queue", len(tracks)))
		return
	}

	tracks, err := src.Search(ctx, query)
	if err != nil {
		followup(s, i, fmt.Sprintf("Search failed: %v", err))
		return
	}
	if len(tracks) == 0 {
		followup(s, i, "No results")
		return
	}

	b.cacheSearch(i.ID, tracks)
	followupComplex(s, i, &discordgo.WebhookParams{
		Embeds:     []*discordgo.MessageEmbed{views.SearchEmbed(tracks)},
		Components: views.SearchComponents(i.ID, len(tracks)),
	})
}

// enqueueAll connects to voice and enqueues every track, playing the first.
// Returns false (after sending an error followup) if joining voice fails.
func (b *Bot) enqueueAll(s *discordgo.Session, i *discordgo.InteractionCreate, tracks []audio.PlayableAudio) bool {
	if _, err := voice.ConnectOrMove(s, i.GuildID, i.Member.User.ID); err != nil {
		followup(s, i, err.Error())
		return false
	}
	ap := b.mgr.Player(i.GuildID)
	ap.EnqueueAndPlay(tracks[0], i.ChannelID)
	for _, t := range tracks[1:] {
		ap.Enqueue(t)
	}
	return true
}

func (b *Bot) cmdQueue(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ap := b.mgr.Player(i.GuildID)
	nowPlaying, queue := ap.Snapshot()
	if nowPlaying == nil && len(queue) == 0 {
		respond(s, i, "The queue is empty")
		return
	}
	respond(s, i, ap.QueueText(true, true))
}

// --- search result cache ---

func (b *Bot) cacheSearch(key string, tracks []audio.PlayableAudio) {
	b.searchMu.Lock()
	b.searches[key] = tracks
	b.searchMu.Unlock()
}

func (b *Bot) takeSearch(key string) ([]audio.PlayableAudio, bool) {
	b.searchMu.Lock()
	defer b.searchMu.Unlock()
	tracks, ok := b.searches[key]
	if ok {
		delete(b.searches, key)
	}
	return tracks, ok
}

// handleComponent handles a search-result button press. Ports
// SearchResultButton.callback.
func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	key, index, ok := views.ParseComponentID(i.MessageComponentData().CustomID)
	if !ok {
		return
	}
	tracks, found := b.takeSearch(key)
	if !found || index < 0 || index >= len(tracks) {
		updateMessage(s, i, "That selection has expired")
		return
	}

	track := tracks[index]
	if _, err := voice.ConnectOrMove(s, i.GuildID, i.Member.User.ID); err != nil {
		updateMessage(s, i, err.Error())
		return
	}
	b.mgr.Player(i.GuildID).EnqueueAndPlay(track, i.ChannelID)
	updateMessage(s, i, fmt.Sprintf("Added %s", track.DisplayName()))
}
