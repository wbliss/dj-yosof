package player

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kkdai/youtube/v2"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/voice"
)

const youtubeWatchPrefix = "https://www.youtube.com/watch?v="

// YoutubeSource streams audio and resolves metadata from YouTube. It ports
// djyosof/players/youtube.py. Search is implemented by scraping the results
// page (matching the no-API-key behaviour of the original pytubefix Search).
type YoutubeSource struct {
	client *youtube.Client
	http   *http.Client
}

// NewYoutubeSource creates a YouTube source.
func NewYoutubeSource() *YoutubeSource {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	return &YoutubeSource{
		client: &youtube.Client{HTTPClient: httpClient},
		http:   httpClient,
	}
}

// Play implements Source. Ports YoutubeSource.play + load_track.
func (y *YoutubeSource) Play(ctx context.Context, track audio.PlayableAudio, vc *discordgo.VoiceConnection) error {
	yt, ok := track.(*audio.YoutubeTrack)
	if !ok {
		return fmt.Errorf("youtube: unexpected track type %T", track)
	}

	video, err := y.client.GetVideoContext(ctx, yt.WatchURL)
	if err != nil {
		return fmt.Errorf("youtube: fetching video: %w", err)
	}

	format := selectAudioFormat(video.Formats)
	if format == nil {
		return fmt.Errorf("youtube: no audio formats for %s", yt.WatchURL)
	}

	stream, _, err := y.client.GetStreamContext(ctx, video, format)
	if err != nil {
		return fmt.Errorf("youtube: opening stream: %w", err)
	}
	defer func() { _ = stream.Close() }()

	// Let ffmpeg auto-detect the container/codec from the stream.
	return voice.Stream(ctx, vc, nil, stream)
}

// selectAudioFormat picks the best audio-only format, falling back to any
// format carrying audio.
func selectAudioFormat(formats youtube.FormatList) *youtube.Format {
	withAudio := formats.WithAudioChannels()
	var best *youtube.Format
	for i := range withAudio {
		f := &withAudio[i]
		if !strings.HasPrefix(f.MimeType, "audio/") {
			continue
		}
		if best == nil || f.Bitrate > best.Bitrate {
			best = f
		}
	}
	if best != nil {
		return best
	}
	if len(withAudio) > 0 {
		return &withAudio[0]
	}
	return nil
}

// Search implements Source. Scrapes the YouTube results page for the top 5
// videos. Replaces pytubefix Search.
func (y *YoutubeSource) Search(ctx context.Context, query string) ([]audio.PlayableAudio, error) {
	u := "https://www.youtube.com/results?search_query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	// A desktop UA and English locale keep the response shape stable.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := y.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube: search request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: reading search page: %w", err)
	}

	renderers, err := parseSearchResults(string(body))
	if err != nil {
		return nil, err
	}

	tracks := make([]audio.PlayableAudio, 0, 5)
	for _, vr := range renderers {
		if vr.VideoID == "" {
			continue
		}
		tracks = append(tracks, videoRendererToTrack(vr))
		if len(tracks) >= 5 {
			break
		}
	}
	return tracks, nil
}

// OpenLink implements Source. Ports YoutubeSource.open_link.
func (y *YoutubeSource) OpenLink(ctx context.Context, link string) ([]audio.PlayableAudio, error) {
	if strings.Contains(link, "list=") {
		playlist, err := y.client.GetPlaylistContext(ctx, link)
		if err != nil {
			return nil, fmt.Errorf("youtube: fetching playlist: %w", err)
		}
		tracks := make([]audio.PlayableAudio, 0, len(playlist.Videos))
		for _, entry := range playlist.Videos {
			tracks = append(tracks, &audio.YoutubeTrack{
				Title:        entry.Title,
				ThumbnailURL: lastThumbnail(entry.Thumbnails),
				Length:       entry.Duration,
				WatchURL:     youtubeWatchPrefix + entry.ID,
			})
		}
		return tracks, nil
	}

	video, err := y.client.GetVideoContext(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("youtube: fetching video: %w", err)
	}
	return []audio.PlayableAudio{&audio.YoutubeTrack{
		Title:        video.Title,
		ThumbnailURL: lastThumbnail(video.Thumbnails),
		Length:       video.Duration,
		WatchURL:     youtubeWatchPrefix + video.ID,
	}}, nil
}

func videoRendererToTrack(vr videoRenderer) *audio.YoutubeTrack {
	title := ""
	if len(vr.Title.Runs) > 0 {
		title = vr.Title.Runs[0].Text
	}
	thumb := ""
	if len(vr.Thumbnail.Thumbnails) > 0 {
		thumb = vr.Thumbnail.Thumbnails[len(vr.Thumbnail.Thumbnails)-1].URL
	}
	return &audio.YoutubeTrack{
		Title:        title,
		ThumbnailURL: thumb,
		Length:       parseLengthText(vr.LengthText.SimpleText),
		WatchURL:     youtubeWatchPrefix + vr.VideoID,
	}
}

func lastThumbnail(thumbs youtube.Thumbnails) string {
	if len(thumbs) == 0 {
		return ""
	}
	return thumbs[len(thumbs)-1].URL
}

// parseLengthText converts "h:mm:ss" or "m:ss" into a Duration.
func parseLengthText(s string) time.Duration {
	parts := strings.Split(strings.TrimSpace(s), ":")
	var total time.Duration
	for _, p := range parts {
		n, _ := strconv.Atoi(strings.TrimSpace(p))
		total = total*60 + time.Duration(n)
	}
	return total * time.Second
}
