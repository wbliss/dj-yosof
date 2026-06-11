package player

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/ap"
	splayer "github.com/devgianlu/go-librespot/player"
	devicespb "github.com/devgianlu/go-librespot/proto/spotify/connectstate/devices"
	"github.com/devgianlu/go-librespot/session"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/util"
	"github.com/GusPrice/dj-yosof/voice"
)

// spotifyBitrate selects the 320kbps Ogg Vorbis stream (Premium), matching the
// Python VorbisOnlyAudioQuality.VERY_HIGH.
const spotifyBitrate = 320

// SpotifySource streams audio and resolves metadata via go-librespot. It ports
// djyosof/players/spotify.py.
//
// Unlike the Python version (which used username/password), go-librespot only
// supports OAuth2: on first run an interactive browser login is performed and
// the resulting credentials are cached in the credentials file for reuse.
type SpotifySource struct {
	sess         *session.Session
	player       *splayer.Player
	httpClient   *http.Client
	countryCode  *string
	countryReady chan struct{}
}

// spotifyCredentials is the JSON persisted to the credentials cache file.
type spotifyCredentials struct {
	DeviceID string `json:"device_id"`
	Username string `json:"username"`
	Data     string `json:"data"` // base64-encoded stored-credentials blob
}

// NewSpotifySource authenticates with Spotify and prepares the player. credFile
// is the credentials cache path; callbackPort is the OAuth2 redirect port used
// only on first-time interactive login.
func NewSpotifySource(ctx context.Context, credFile string, callbackPort int) (*SpotifySource, error) {
	logger := newSpotifyLogger()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	stored, _ := loadSpotifyCredentials(credFile)

	deviceID := ""
	var creds any
	if stored != nil {
		deviceID = stored.DeviceID
		data, err := base64.StdEncoding.DecodeString(stored.Data)
		if err != nil {
			return nil, fmt.Errorf("decoding cached spotify credentials: %w", err)
		}
		creds = session.StoredCredentials{Username: stored.Username, Data: data}
	} else {
		var err error
		if deviceID, err = generateDeviceID(); err != nil {
			return nil, err
		}
		creds = session.InteractiveCredentials{CallbackPort: callbackPort}
	}

	sess, err := session.NewSessionFromOptions(ctx, &session.Options{
		Log:         logger,
		DeviceType:  devicespb.DeviceType_COMPUTER,
		DeviceId:    deviceID,
		Credentials: creds,
		Client:      httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("creating spotify session: %w", err)
	}

	// Persist freshly-obtained credentials for next time.
	if stored == nil {
		if err := saveSpotifyCredentials(credFile, &spotifyCredentials{
			DeviceID: deviceID,
			Username: sess.Username(),
			Data:     base64.StdEncoding.EncodeToString(sess.StoredCredentials()),
		}); err != nil {
			logger.Warnf("failed persisting spotify credentials: %v", err)
		}
	}

	countryCode := new(string)
	pl, err := splayer.NewPlayer(&splayer.Options{
		Spclient:    sess.Spclient(),
		AudioKey:    sess.AudioKey(),
		Events:      sess.Events(),
		Log:         logger,
		CountryCode: countryCode,
	})
	if err != nil {
		sess.Close()
		return nil, fmt.Errorf("creating spotify player: %w", err)
	}

	s := &SpotifySource{
		sess:         sess,
		player:       pl,
		httpClient:   httpClient,
		countryCode:  countryCode,
		countryReady: make(chan struct{}),
	}

	// Start receiving the country-code packet (this also starts the AP receive
	// loop) and wait briefly for it so the first stream isn't rejected as
	// restricted. We proceed the instant it arrives rather than polling.
	go s.receiveCountryCode()
	select {
	case <-s.countryReady:
	case <-time.After(5 * time.Second):
	}

	return s, nil
}

func (s *SpotifySource) receiveCountryCode() {
	var once sync.Once
	for pkt := range s.sess.Accesspoint().Receive(ap.PacketTypeCountryCode) {
		if pkt.Type == ap.PacketTypeCountryCode {
			*s.countryCode = string(pkt.Payload)
			once.Do(func() { close(s.countryReady) })
		}
	}
}

// Close shuts down the underlying session and player.
func (s *SpotifySource) Close() {
	s.player.Close()
	s.sess.Close()
}

// Play implements Source. Ports SpotifySource.play + load_track.
func (s *SpotifySource) Play(ctx context.Context, track audio.PlayableAudio, vc *discordgo.VoiceConnection) error {
	st, ok := track.(*audio.SpotifyTrack)
	if !ok {
		return fmt.Errorf("spotify: unexpected track type %T", track)
	}

	spotID, err := librespot.SpotifyIdFromBase62(librespot.SpotifyIdTypeTrack, st.TrackID)
	if err != nil {
		return fmt.Errorf("spotify: invalid track id %q: %w", st.TrackID, err)
	}

	stream, err := s.player.NewStream(ctx, s.httpClient, *spotID, spotifyBitrate, 0)
	if err != nil {
		return fmt.Errorf("spotify: opening stream: %w", err)
	}

	// go-librespot decodes to 44100Hz stereo float32 PCM; tell ffmpeg the raw
	// input format so it can resample to Discord's 48kHz.
	inputArgs := []string{"-f", "s16le", "-ar", strconv.Itoa(splayer.SampleRate), "-ac", strconv.Itoa(splayer.Channels)}
	return voice.Stream(ctx, vc, inputArgs, newPCMReader(stream.Source))
}

// Search implements Source. Ports SpotifySource.search.
func (s *SpotifySource) Search(ctx context.Context, query string) ([]audio.PlayableAudio, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("type", "track")
	q.Set("limit", "5")

	var resp struct {
		Tracks struct {
			Items []apiTrack `json:"items"`
		} `json:"tracks"`
	}
	if err := s.webAPI(ctx, "/v1/search", q, &resp); err != nil {
		return nil, err
	}

	tracks := make([]audio.PlayableAudio, 0, len(resp.Tracks.Items))
	for i := range resp.Tracks.Items {
		tracks = append(tracks, resp.Tracks.Items[i].toTrack("", ""))
	}
	return tracks, nil
}

// OpenLink implements Source. Ports SpotifySource.open_link.
func (s *SpotifySource) OpenLink(ctx context.Context, link string) ([]audio.PlayableAudio, error) {
	m := util.SpotifyLinkRegexp.FindStringSubmatch(link)
	if m == nil {
		return nil, fmt.Errorf("spotify: not a recognised link: %s", link)
	}
	mediaType, mediaID := m[1], m[2]

	switch mediaType {
	case "track":
		var t apiTrack
		if err := s.webAPI(ctx, "/v1/tracks/"+mediaID, nil, &t); err != nil {
			return nil, err
		}
		return []audio.PlayableAudio{t.toTrack("", "")}, nil

	case "album":
		var resp struct {
			Name   string     `json:"name"`
			Images []apiImage `json:"images"`
			Tracks struct {
				Items []apiTrack `json:"items"`
			} `json:"tracks"`
		}
		if err := s.webAPI(ctx, "/v1/albums/"+mediaID, nil, &resp); err != nil {
			return nil, err
		}
		art := firstImageURL(resp.Images)
		tracks := make([]audio.PlayableAudio, 0, len(resp.Tracks.Items))
		for i := range resp.Tracks.Items {
			tracks = append(tracks, resp.Tracks.Items[i].toTrack(resp.Name, art))
		}
		return tracks, nil

	case "playlist":
		var resp struct {
			Tracks struct {
				Items []struct {
					Track apiTrack `json:"track"`
				} `json:"items"`
			} `json:"tracks"`
		}
		if err := s.webAPI(ctx, "/v1/playlists/"+mediaID, nil, &resp); err != nil {
			return nil, err
		}
		tracks := make([]audio.PlayableAudio, 0, len(resp.Tracks.Items))
		for i := range resp.Tracks.Items {
			tracks = append(tracks, resp.Tracks.Items[i].Track.toTrack("", ""))
		}
		return tracks, nil

	default:
		return nil, fmt.Errorf("spotify: unsupported media type %q", mediaType)
	}
}

// webAPI retry tuning for HTTP 429 (rate limit).
const (
	spotifyMaxRetries   = 3
	spotifyMaxRetryWait = 15 * time.Second
)

// webAPI performs an authenticated Spotify Web API GET and decodes JSON into out.
// It transparently retries on HTTP 429, honoring the Retry-After header, up to
// spotifyMaxRetries times. If the server asks to wait longer than
// spotifyMaxRetryWait, it returns a clear "try again" error instead of blocking.
func (s *SpotifySource) webAPI(ctx context.Context, path string, query url.Values, out any) error {
	for attempt := 0; ; attempt++ {
		resp, err := s.sess.WebApi(ctx, http.MethodGet, path, query, nil, nil)
		if err != nil {
			return fmt.Errorf("spotify web api %s: %w", path, err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("spotify web api %s: reading body: %w", path, err)
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait := retryAfter(resp.Header)
			if attempt >= spotifyMaxRetries || wait > spotifyMaxRetryWait {
				return fmt.Errorf("Spotify is rate limiting requests; try again in %d second(s)", int(wait.Seconds()))
			}
			log.Printf("[spotify] rate limited on %s, retrying in %s", path, wait)
			select {
			case <-time.After(wait):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("spotify web api %s: status %d: %s", path, resp.StatusCode, string(body))
		}
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("spotify web api %s: decoding: %w", path, err)
		}
		return nil
	}
}

// retryAfter returns how long to wait before retrying a 429, taken from the
// Retry-After header (delta-seconds), defaulting to 1s when absent/unparseable.
func retryAfter(h http.Header) time.Duration {
	if v := h.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return time.Second
}

// --- Web API JSON shapes ---

type apiImage struct {
	URL string `json:"url"`
}

type apiArtist struct {
	Name string `json:"name"`
}

type apiAlbum struct {
	Name   string     `json:"name"`
	Images []apiImage `json:"images"`
}

type apiTrack struct {
	ID      string      `json:"id"`
	Name    string      `json:"name"`
	Artists []apiArtist `json:"artists"`
	Album   apiAlbum    `json:"album"`
}

// toTrack converts an API track into a SpotifyTrack. albumName/albumArt are used
// as fallbacks when the track payload omits them (e.g. album-listing items).
func (t apiTrack) toTrack(albumName, albumArt string) *audio.SpotifyTrack {
	artist := ""
	if len(t.Artists) > 0 {
		artist = t.Artists[0].Name
	}
	album := t.Album.Name
	if album == "" {
		album = albumName
	}
	art := firstImageURL(t.Album.Images)
	if art == "" {
		art = albumArt
	}
	return &audio.SpotifyTrack{
		Name:        t.Name,
		Artist:      artist,
		Album:       album,
		AlbumArtURL: art,
		TrackID:     t.ID,
	}
}

func firstImageURL(images []apiImage) string {
	if len(images) > 0 {
		return images[0].URL
	}
	return ""
}

// --- credentials persistence ---

func loadSpotifyCredentials(path string) (*spotifyCredentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c spotifyCredentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Data == "" || c.DeviceID == "" {
		return nil, fmt.Errorf("incomplete credentials")
	}
	return &c, nil
}

func saveSpotifyCredentials(path string, c *spotifyCredentials) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func generateDeviceID() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating device id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// --- float32 PCM -> s16le byte stream adapter ---

// pcmReader adapts a librespot AudioSource (44100Hz stereo float32) into an
// io.Reader of little-endian signed-16 PCM suitable for piping into ffmpeg.
type pcmReader struct {
	src     librespot.AudioSource
	floats  []float32
	pending []byte
	err     error
}

func newPCMReader(src librespot.AudioSource) *pcmReader {
	return &pcmReader{src: src, floats: make([]float32, 4096)}
}

func (r *pcmReader) Read(p []byte) (int, error) {
	if len(r.pending) == 0 {
		if r.err != nil {
			return 0, r.err
		}
		n, err := r.src.Read(r.floats)
		if n > 0 {
			r.pending = r.pending[:0]
			if cap(r.pending) < n*2 {
				r.pending = make([]byte, 0, n*2)
			}
			for i := 0; i < n; i++ {
				v := r.floats[i]
				if v > 1 {
					v = 1
				} else if v < -1 {
					v = -1
				}
				s := uint16(int16(v * 32767))
				r.pending = append(r.pending, byte(s), byte(s>>8))
			}
		}
		if err != nil {
			r.err = err
			if len(r.pending) == 0 {
				return 0, err
			}
		}
	}

	nc := copy(p, r.pending)
	r.pending = r.pending[nc:]
	return nc, nil
}
