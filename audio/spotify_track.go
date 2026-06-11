package audio

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
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
func (t *SpotifyTrack) Embed() discord.Embed {
	return discord.NewEmbed().
		WithTitle("Now Playing").
		AddField("Track", t.Name, false).
		AddField("Artist", t.Artist, false).
		AddField("Album", t.Album, false).
		WithImage(t.AlbumArtURL)
}

// Type implements PlayableAudio.
func (t *SpotifyTrack) Type() Type { return TypeSpotify }

// DisplayName implements PlayableAudio.
func (t *SpotifyTrack) DisplayName() string {
	return fmt.Sprintf("%s by %s", t.Name, t.Artist)
}
