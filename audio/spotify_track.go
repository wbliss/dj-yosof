package audio

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// SpotifyTrack holds the metadata needed to display and stream a Spotify track.
// Mirrors djyosof/audio_types/spotify.py.
type SpotifyTrack struct {
	Name        string
	Artist      string
	Album       string
	AlbumArtURL string
	// TrackID is the base62 Spotify track id (the "{22}" portion of a link).
	TrackID string
}

// Embed implements PlayableAudio.
func (t *SpotifyTrack) Embed() *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title: "Now Playing",
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Track", Value: t.Name},
			{Name: "Artist", Value: t.Artist},
			{Name: "Album", Value: t.Album},
		},
		Image: &discordgo.MessageEmbedImage{URL: t.AlbumArtURL},
	}
}

// Type implements PlayableAudio.
func (t *SpotifyTrack) Type() Type { return TypeSpotify }

// DisplayName implements PlayableAudio.
func (t *SpotifyTrack) DisplayName() string {
	return fmt.Sprintf("%s by %s", t.Name, t.Artist)
}
