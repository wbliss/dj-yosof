package player

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"

	"github.com/GusPrice/dj-yosof/audio"
	"github.com/GusPrice/dj-yosof/voice"
)

// idleTimeout is how long the play loop waits on an empty queue before
// disconnecting, matching the 10s behaviour in the Python play_loop.
const idleTimeout = 10 * time.Second

// queuePreviewLen is how many tracks the now-playing / queue views show.
const queuePreviewLen = 10

// AudioPlayer manages the queue and playback loop for a single guild. Ports
// djyosof/players/audio_player.py.
type AudioPlayer struct {
	mgr     *Manager
	guildID snowflake.ID

	mu              sync.Mutex
	queue           []audio.PlayableAudio
	nowPlaying      audio.PlayableAudio
	playing         bool
	textChannelID   snowflake.ID
	nowPlayingMsgID snowflake.ID
	cancel          context.CancelFunc

	notify chan struct{}
}

func newAudioPlayer(mgr *Manager, guildID snowflake.ID) *AudioPlayer {
	return &AudioPlayer{
		mgr:     mgr,
		guildID: guildID,
		notify:  make(chan struct{}, 1),
	}
}

func (ap *AudioPlayer) signal() {
	select {
	case ap.notify <- struct{}{}:
	default:
	}
}

// EnqueueAndPlay adds a track and starts the play loop if it is not running.
// Ports AudioPlayer.enqueue_and_play.
func (ap *AudioPlayer) EnqueueAndPlay(track audio.PlayableAudio, textChannelID snowflake.ID) {
	ap.mu.Lock()
	ap.textChannelID = textChannelID
	ap.queue = append(ap.queue, track)
	start := !ap.playing
	if start {
		ap.playing = true
	}
	ap.mu.Unlock()

	ap.signal()
	if start {
		go ap.playLoop()
	}
}

// Enqueue adds a track to the queue without (re)starting the loop. Ports
// AudioPlayer.enqueue.
func (ap *AudioPlayer) Enqueue(track audio.PlayableAudio) {
	ap.mu.Lock()
	ap.queue = append(ap.queue, track)
	ap.mu.Unlock()
	ap.signal()
}

// Skip stops the current track, causing the loop to advance. Ports
// AudioPlayer.skip.
func (ap *AudioPlayer) Skip() {
	ap.mu.Lock()
	cancel := ap.cancel
	ap.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Stop clears the queue and stops the current track. Ports AudioPlayer.stop.
func (ap *AudioPlayer) Stop() {
	ap.mu.Lock()
	ap.queue = nil
	cancel := ap.cancel
	ap.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// Snapshot returns the now-playing track and a copy of the pending queue, used
// by the /queue command.
func (ap *AudioPlayer) Snapshot() (audio.PlayableAudio, []audio.PlayableAudio) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	q := make([]audio.PlayableAudio, len(ap.queue))
	copy(q, ap.queue)
	return ap.nowPlaying, q
}

func (ap *AudioPlayer) playLoop() {
	defer func() {
		ap.mu.Lock()
		ap.playing = false
		ap.nowPlaying = nil
		ap.mu.Unlock()
	}()

	for {
		track, ok := ap.waitForTrack()
		if !ok {
			voice.Leave(ap.mgr.client, ap.guildID)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		ap.mu.Lock()
		ap.nowPlaying = track
		ap.cancel = cancel
		textChannel := ap.textChannelID
		ap.mu.Unlock()

		ap.updateNowPlaying(track, textChannel)
		log.Printf("Playing %s", track.DisplayName())

		conn := ap.mgr.client.VoiceManager.GetConn(ap.guildID)
		if conn == nil {
			cancel()
			return
		}

		src := ap.mgr.Source(track.Type())
		if src == nil {
			ap.sendError(textChannel, fmt.Sprintf("No source registered for %s", track.Type()))
		} else if err := src.Play(ctx, track, conn); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("error playing %s: %v", track.DisplayName(), err)
			ap.sendError(textChannel, fmt.Sprintf("Error playing %s", track.DisplayName()))
		}
		cancel()
	}
}

// waitForTrack pops the next track, blocking until one is enqueued or the idle
// timeout elapses. ok is false when the loop should disconnect and exit.
func (ap *AudioPlayer) waitForTrack() (audio.PlayableAudio, bool) {
	for {
		ap.mu.Lock()
		if len(ap.queue) > 0 {
			track := ap.queue[0]
			ap.queue = ap.queue[1:]
			ap.mu.Unlock()
			return track, true
		}
		ap.mu.Unlock()

		select {
		case <-ap.notify:
		case <-time.After(idleTimeout):
			ap.mu.Lock()
			empty := len(ap.queue) == 0
			ap.mu.Unlock()
			if empty {
				return nil, false
			}
		}
	}
}

func (ap *AudioPlayer) updateNowPlaying(track audio.PlayableAudio, textChannel snowflake.ID) {
	if textChannel == 0 {
		return
	}
	rest := ap.mgr.client.Rest

	ap.mu.Lock()
	prev := ap.nowPlayingMsgID
	ap.nowPlayingMsgID = 0
	ap.mu.Unlock()
	if prev != 0 {
		_ = rest.DeleteMessage(textChannel, prev)
	}

	embed := track.Embed()
	if md := ap.QueueText(false, false); md != "" {
		embed = embed.AddField("Up Next", md, false)
	}

	msg, err := rest.CreateMessage(textChannel, discord.NewMessageCreate().WithEmbeds(embed))
	if err != nil {
		log.Printf("failed sending now-playing message: %v", err)
		return
	}
	ap.mu.Lock()
	ap.nowPlayingMsgID = msg.ID
	ap.mu.Unlock()
}

func (ap *AudioPlayer) sendError(textChannel snowflake.ID, msg string) {
	if textChannel == 0 {
		return
	}
	_, _ = ap.mgr.client.Rest.CreateMessage(textChannel, discord.NewMessageCreate().WithContent(msg))
}

// QueueText renders the queue for display, mirroring the Python
// _get_queue_markdown. withNowPlaying prefixes the current track; withTotals
// appends an "N out of M tracks" footer.
func (ap *AudioPlayer) QueueText(withNowPlaying, withTotals bool) string {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	var b strings.Builder
	if withNowPlaying && ap.nowPlaying != nil {
		fmt.Fprintf(&b, "**NOW PLAYING:** %s\n", ap.nowPlaying.DisplayName())
	}
	shown := 0
	for i, t := range ap.queue {
		if i >= queuePreviewLen {
			break
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, t.DisplayName())
		shown++
	}
	if withTotals {
		fmt.Fprintf(&b, "\n%d out of %d tracks", shown, len(ap.queue))
	}
	return b.String()
}
