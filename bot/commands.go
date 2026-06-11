package bot

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/util"
	"github.com/GusPrice/dj-yosof/views"
	"github.com/GusPrice/dj-yosof/voice"
)

// commands defines the slash commands, mirroring the Python cogs.
var commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{Name: "hello", Description: "Say hello"},
	discord.SlashCommandCreate{Name: "join", Description: "Join your voice channel"},
	discord.SlashCommandCreate{Name: "leave", Description: "Leave the voice channel"},
	discord.SlashCommandCreate{
		Name:        "spotify",
		Description: "Play a Spotify track/album/playlist link, or search Spotify",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "query", Description: "Link or search terms", Required: true},
		},
	},
	discord.SlashCommandCreate{
		Name:        "yt",
		Description: "Play a YouTube link, or search YouTube",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "query", Description: "Link or search terms", Required: true},
		},
	},
	discord.SlashCommandCreate{Name: "pause", Description: "Pause playback"},
	discord.SlashCommandCreate{Name: "queue", Description: "Show the queue"},
	discord.SlashCommandCreate{Name: "skip", Description: "Skip the current track"},
	discord.SlashCommandCreate{Name: "stop", Description: "Stop playback and clear the queue"},
}

func (b *Bot) registerCommands() error {
	for _, gid := range b.cfg.GuildIDs {
		guildID, err := snowflake.Parse(gid)
		if err != nil {
			return fmt.Errorf("invalid guild id %q: %w", gid, err)
		}
		if _, err := b.client.Rest.SetGuildCommands(b.client.ApplicationID, guildID, commands); err != nil {
			return fmt.Errorf("registering commands in guild %s: %w", gid, err)
		}
	}
	return nil
}

func (b *Bot) handleCommand(e *events.ApplicationCommandInteractionCreate) {
	switch e.SlashCommandInteractionData().CommandName() {
	case "hello":
		reply(e, fmt.Sprintf("Hi, %s", e.User().Mention()))
	case "join":
		b.cmdJoin(e)
	case "leave":
		b.cmdLeave(e)
	case "spotify":
		b.cmdPlay(e, audio.TypeSpotify, util.IsSpotifyLink)
	case "yt":
		b.cmdPlay(e, audio.TypeYoutube, util.IsYoutubeLink)
	case "pause":
		reply(e, "Pause is not implemented yet")
	case "queue":
		b.cmdQueue(e)
	case "skip":
		b.mgr.Player(*e.GuildID()).Skip()
		reply(e, "Skipped")
	case "stop":
		b.mgr.Player(*e.GuildID()).Stop()
		reply(e, "Stopped")
	}
}

func (b *Bot) cmdJoin(e *events.ApplicationCommandInteractionCreate) {
	// Connecting now includes the DAVE handshake and can exceed the 3s reply
	// window, so defer first and follow up.
	_ = e.DeferCreateMessage(false)
	if _, err := voice.ConnectOrMove(b.client, *e.GuildID(), e.User().ID); err != nil {
		followup(e, err.Error())
		return
	}
	followup(e, "Joined your voice channel")
}

func (b *Bot) cmdLeave(e *events.ApplicationCommandInteractionCreate) {
	voice.Leave(b.client, *e.GuildID())
	reply(e, "Left the voice channel")
}

// cmdPlay handles both /spotify and /yt: link → enqueue immediately; otherwise
// search and present selection buttons. Ports SpotifyCog.spotify / YoutubeCog.yt.
func (b *Bot) cmdPlay(e *events.ApplicationCommandInteractionCreate, srcType audio.Type, isLink func(string) bool) {
	src := b.mgr.Source(srcType)
	if src == nil {
		reply(e, fmt.Sprintf("%s is not available", srcType))
		return
	}
	query := e.SlashCommandInteractionData().String("query")

	_ = e.DeferCreateMessage(false)
	ctx := context.Background()

	if isLink(query) {
		tracks, err := src.OpenLink(ctx, query)
		if err != nil {
			followup(e, fmt.Sprintf("Couldn't load that link: %v", err))
			return
		}
		if len(tracks) == 0 {
			followup(e, "That link had no playable tracks")
			return
		}
		if !b.enqueueAll(e, tracks) {
			return
		}
		followup(e, fmt.Sprintf("Added %d track(s) to the queue", len(tracks)))
		return
	}

	tracks, err := src.Search(ctx, query)
	if err != nil {
		followup(e, fmt.Sprintf("Search failed: %v", err))
		return
	}
	if len(tracks) == 0 {
		followup(e, "No results")
		return
	}

	key := e.ID().String()
	b.cacheSearch(key, tracks)
	followupComplex(e, discord.NewMessageCreate().
		WithEmbeds(views.SearchEmbed(tracks)).
		WithComponents(views.SearchComponents(key, len(tracks))...))
}

// enqueueAll connects to voice and enqueues every track, playing the first.
// Returns false (after sending an error followup) if joining voice fails.
func (b *Bot) enqueueAll(e *events.ApplicationCommandInteractionCreate, tracks []audio.PlayableAudio) bool {
	if _, err := voice.ConnectOrMove(b.client, *e.GuildID(), e.User().ID); err != nil {
		followup(e, err.Error())
		return false
	}
	ap := b.mgr.Player(*e.GuildID())
	ap.EnqueueAndPlay(tracks[0], e.Channel().ID())
	for _, t := range tracks[1:] {
		ap.Enqueue(t)
	}
	return true
}

func (b *Bot) cmdQueue(e *events.ApplicationCommandInteractionCreate) {
	ap := b.mgr.Player(*e.GuildID())
	nowPlaying, queue := ap.Snapshot()
	if nowPlaying == nil && len(queue) == 0 {
		reply(e, "The queue is empty")
		return
	}
	reply(e, ap.QueueText(true, true))
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
func (b *Bot) handleComponent(e *events.ComponentInteractionCreate) {
	key, index, ok := views.ParseComponentID(e.ButtonInteractionData().CustomID())
	if !ok {
		return
	}
	tracks, found := b.takeSearch(key)
	if !found || index < 0 || index >= len(tracks) {
		_ = e.UpdateMessage(discord.NewMessageUpdate().
			WithContent("That selection has expired").
			ClearEmbeds().
			ClearComponents())
		return
	}

	track := tracks[index]
	// Ack now (connecting may exceed the 3s window), then edit the message.
	_ = e.DeferUpdateMessage()
	if _, err := voice.ConnectOrMove(b.client, *e.GuildID(), e.User().ID); err != nil {
		editComponent(e, err.Error())
		return
	}
	b.mgr.Player(*e.GuildID()).EnqueueAndPlay(track, e.Channel().ID())
	editComponent(e, fmt.Sprintf("Added %s", track.DisplayName()))
}
