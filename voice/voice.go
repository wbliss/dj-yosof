// Package voice handles Discord voice connections and the audio streaming
// pipeline (ffmpeg -> PCM -> Opus -> Discord). It ports djyosof/cogs/utilities.py
// and the playback half of the Python audio sources.
package voice

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

// ErrNotInVoice is returned when the requesting user is not in a voice channel.
var ErrNotInVoice = errors.New("you must be in a voice channel")

// userVoiceChannel finds the voice channel the given user is currently in.
func userVoiceChannel(s *discordgo.Session, guildID, userID string) (string, error) {
	g, err := s.State.Guild(guildID)
	if err != nil {
		return "", err
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			return vs.ChannelID, nil
		}
	}
	return "", ErrNotInVoice
}

// ConnectOrMove connects the bot to the user's voice channel, moving if it is
// already connected to a different one. Ports utilities.connect_or_move.
func ConnectOrMove(s *discordgo.Session, guildID, userID string) (*discordgo.VoiceConnection, error) {
	channelID, err := userVoiceChannel(s, guildID, userID)
	if err != nil {
		return nil, err
	}

	if vc, ok := s.VoiceConnections[guildID]; ok && vc.ChannelID == channelID && vc.Ready {
		return vc, nil
	}

	// ChannelVoiceJoin connects, or moves an existing connection.
	return s.ChannelVoiceJoin(guildID, channelID, false, true)
}

// Leave disconnects the bot from the guild's voice channel. Ports
// utilities.leave.
func Leave(s *discordgo.Session, guildID string) error {
	if vc, ok := s.VoiceConnections[guildID]; ok {
		return vc.Disconnect()
	}
	return nil
}
