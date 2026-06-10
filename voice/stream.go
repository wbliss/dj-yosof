package voice

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

const (
	// Discord voice expects 48kHz stereo Opus.
	sampleRate = 48000
	channels   = 2
	// 20ms frame: 960 samples per channel.
	frameSize  = 960
	maxOpusLen = 4000
)

// Stream decodes audio from src using ffmpeg, encodes it to Opus and sends it to
// the voice connection until the source ends or ctx is cancelled. inputArgs are
// the ffmpeg input flags placed before "-i pipe:0" (e.g. raw-PCM format flags);
// pass nil to let ffmpeg auto-detect a container/codec.
//
// This replaces discord.py's FFmpegOpusAudio / FFmpegPCMAudio + voice.play.
func Stream(ctx context.Context, vc *discordgo.VoiceConnection, inputArgs []string, src io.Reader) error {
	args := append([]string{"-hide_banner", "-loglevel", "error"}, inputArgs...)
	args = append(args,
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdin = src
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting ffmpeg: %w", err)
	}
	// Ensure the process is reaped and the pipeline torn down on return.
	defer func() { _ = cmd.Wait() }()

	enc, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return fmt.Errorf("creating opus encoder: %w", err)
	}

	_ = vc.Speaking(true)
	defer func() { _ = vc.Speaking(false) }()

	pcm := make([]int16, frameSize*channels)
	buf := make([]byte, frameSize*channels*2)

	for {
		// Stop promptly on skip/stop.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := io.ReadFull(stdout, buf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}

		for i := range pcm {
			pcm[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
		}

		opus, err := enc.Encode(pcm, frameSize, maxOpusLen)
		if err != nil {
			return fmt.Errorf("opus encode: %w", err)
		}

		select {
		case vc.OpusSend <- opus:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
