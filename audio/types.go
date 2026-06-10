// Package audio defines the playable-track abstraction shared by every audio
// source, mirroring djyosof/audio_types in the original Python project.
package audio

import "github.com/bwmarrin/discordgo"

// Type identifies which source a track came from.
type Type string

const (
	// TypeSpotify is a track streamed from Spotify.
	TypeSpotify Type = "spotify"
	// TypeYoutube is a track streamed from YouTube.
	TypeYoutube Type = "youtube"
)

// PlayableAudio is a track that can be queued and played. Both SpotifyTrack and
// YoutubeTrack implement it, allowing the queue to be source-agnostic.
type PlayableAudio interface {
	// Embed renders the "Now Playing" embed for this track.
	Embed() *discordgo.MessageEmbed
	// Type returns the source type of this track.
	Type() Type
	// DisplayName returns a short human-readable label, used in the queue list.
	DisplayName() string
}
