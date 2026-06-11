package audio

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
)

// YoutubeTrack holds the metadata needed to display and stream a YouTube video.
// Mirrors djyosof/audio_types/youtube.py.
type YoutubeTrack struct {
	Title        string
	ThumbnailURL string
	Length       time.Duration
	WatchURL     string
}

// Embed implements PlayableAudio.
func (t *YoutubeTrack) Embed() discord.Embed {
	return discord.NewEmbed().
		WithTitle("Now Playing").
		AddField("Title", fmt.Sprintf("[%s](%s)", t.Title, t.WatchURL), false).
		WithImage(t.ThumbnailURL)
}

// Type implements PlayableAudio.
func (t *YoutubeTrack) Type() Type { return TypeYoutube }

// DisplayName implements PlayableAudio.
func (t *YoutubeTrack) DisplayName() string {
	return fmt.Sprintf("[%s (%s)](<%s>)", t.Title, formatDuration(t.Length), t.WatchURL)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
