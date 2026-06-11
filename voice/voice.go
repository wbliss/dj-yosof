// Package voice handles Discord voice connections and the audio streaming
// pipeline (ffmpeg -> PCM -> Opus -> Discord). It ports djyosof/cogs/utilities.py
// and the playback half of the Python audio sources.
package voice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/bot"
	dvoice "github.com/disgoorg/disgo/voice"
	"github.com/disgoorg/snowflake/v2"
)

// ErrNotInVoice is returned when the requesting user is not in a voice channel.
var ErrNotInVoice = errors.New("you must be in a voice channel")

// connectTimeout bounds the voice gateway handshake (which now includes the
// DAVE/E2EE key exchange).
const connectTimeout = 30 * time.Second

// ConnectOrMove connects the bot to the user's voice channel, moving if it is
// already connected to a different one. Ports utilities.connect_or_move.
func ConnectOrMove(client *bot.Client, guildID, userID snowflake.ID) (dvoice.Conn, error) {
	vs, ok := client.Caches.VoiceState(guildID, userID)
	if !ok || vs.ChannelID == nil {
		return nil, ErrNotInVoice
	}

	conn := client.VoiceManager.CreateConn(guildID)
	if cid := conn.ChannelID(); cid != nil && *cid == *vs.ChannelID {
		return conn, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	// Open connects, or moves an existing connection to the new channel.
	if err := conn.Open(ctx, *vs.ChannelID, false, true); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("timed out connecting to voice — check the bot has Connect/Speak permissions and the host allows outbound UDP to Discord's voice servers (run with debug: true for details)")
		}
		return nil, fmt.Errorf("connecting to voice: %w", err)
	}
	return conn, nil
}

// Leave disconnects the bot from the guild's voice channel. Ports
// utilities.leave.
func Leave(client *bot.Client, guildID snowflake.ID) {
	conn := client.VoiceManager.GetConn(guildID)
	if conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	conn.Close(ctx)
	client.VoiceManager.RemoveConn(guildID)
}
