// Package player implements audio sources (Spotify, YouTube) and the per-guild
// playback queue. It ports djyosof/players and djyosof/cogs/audio_player.py.
package player

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/GusPrice/dj-yosof/audio"
)

// Source resolves and plays tracks for a single backend. Mirrors
// djyosof/players/base_source.py (plus the link/search helpers the cogs use).
type Source interface {
	// Search returns up to a handful of tracks matching a free-text query.
	Search(ctx context.Context, query string) ([]audio.PlayableAudio, error)
	// OpenLink resolves a track/album/playlist URL into its tracks.
	OpenLink(ctx context.Context, link string) ([]audio.PlayableAudio, error)
	// Play streams the track to the voice connection, blocking until the track
	// finishes or ctx is cancelled (used for skip/stop).
	Play(ctx context.Context, track audio.PlayableAudio, vc *discordgo.VoiceConnection) error
}

// Manager owns the registered sources and the per-guild players. It is the
// rough equivalent of the cog/bot wiring that held the Python `players` and
// `audio_players` maps.
type Manager struct {
	session *discordgo.Session

	sources map[audio.Type]Source

	mu      sync.Mutex
	players map[string]*AudioPlayer
}

// NewManager creates a Manager bound to a Discord session.
func NewManager(session *discordgo.Session) *Manager {
	return &Manager{
		session: session,
		sources: make(map[audio.Type]Source),
		players: make(map[string]*AudioPlayer),
	}
}

// RegisterSource associates a source implementation with a track type.
func (m *Manager) RegisterSource(t audio.Type, src Source) {
	m.sources[t] = src
}

// Source returns the source for a track type, or nil if unregistered.
func (m *Manager) Source(t audio.Type) Source {
	return m.sources[t]
}

// Player returns the AudioPlayer for a guild, lazily creating it. This mirrors
// the Python defaultdict[guild_id -> AudioPlayer].
func (m *Manager) Player(guildID string) *AudioPlayer {
	m.mu.Lock()
	defer m.mu.Unlock()
	ap, ok := m.players[guildID]
	if !ok {
		ap = newAudioPlayer(m, guildID)
		m.players[guildID] = ap
	}
	return ap
}
