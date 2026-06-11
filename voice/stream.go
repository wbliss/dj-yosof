package voice

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"

	dvoice "github.com/disgoorg/disgo/voice"
	"layeh.com/gopus"
)

const (
	// Discord voice expects 48kHz stereo Opus.
	sampleRate = 48000
	channels   = 2
	// 20ms frame: 960 samples per channel.
	frameSize  = 960
	maxOpusLen = 4000
	// frameBuffer is how many encoded frames the pump may run ahead of playback.
	frameBuffer = 10
)

// Stream decodes audio from src with ffmpeg, encodes it to Opus, and feeds it to
// the voice connection via an OpusFrameProvider until the source ends or ctx is
// cancelled (skip/stop). inputArgs are the ffmpeg input flags placed before
// "-i pipe:0" (e.g. raw-PCM format flags); pass nil to let ffmpeg auto-detect.
//
// This replaces discord.py's FFmpegOpusAudio / FFmpegPCMAudio + voice.play.
func Stream(ctx context.Context, conn dvoice.Conn, inputArgs []string, src io.Reader) error {
	pctx, cancel := context.WithCancel(ctx)

	args := append([]string{"-hide_banner", "-loglevel", "error"}, inputArgs...)
	args = append(args,
		"-i", "pipe:0",
		"-f", "s16le",
		"-ar", strconv.Itoa(sampleRate),
		"-ac", strconv.Itoa(channels),
		"pipe:1",
	)

	cmd := exec.CommandContext(pctx, "ffmpeg", args...)
	cmd.Stdin = src
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("ffmpeg stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("starting ffmpeg: %w", err)
	}

	provider := &opusProvider{
		frames:  make(chan []byte, frameBuffer),
		cancel:  cancel,
		drained: make(chan struct{}),
	}
	go pump(pctx, stdout, cmd, provider.frames)

	conn.SetOpusFrameProvider(provider)
	defer cancel()

	select {
	case <-provider.drained:
		// The sender consumed every frame: the track played to completion.
		return nil
	case <-ctx.Done():
		// Skip/stop: cancel ffmpeg and stop streaming this track.
		return ctx.Err()
	}
}

// pump reads PCM from ffmpeg, encodes Opus frames, and sends them to frames
// until EOF/error or ctx cancellation. It owns reaping the ffmpeg process, and
// reads stdout then Waits sequentially (never concurrently).
func pump(ctx context.Context, stdout io.Reader, cmd *exec.Cmd, frames chan<- []byte) {
	defer close(frames)
	defer func() { _ = cmd.Wait() }()

	enc, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return
	}

	pcm := make([]int16, frameSize*channels)
	buf := make([]byte, frameSize*channels*2)
	for {
		if _, err := io.ReadFull(stdout, buf); err != nil {
			return // EOF, short read, or ffmpeg killed
		}
		for i := range pcm {
			pcm[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
		}
		opus, err := enc.Encode(pcm, frameSize, maxOpusLen)
		if err != nil {
			return
		}
		select {
		case frames <- opus:
		case <-ctx.Done():
			return
		}
	}
}

// opusProvider implements voice.OpusFrameProvider, pulling pre-encoded Opus
// frames produced by pump. disgo's audio sender calls ProvideOpusFrame every
// 20ms and auto-manages the speaking state.
type opusProvider struct {
	frames    chan []byte
	cancel    context.CancelFunc
	drained   chan struct{}
	drainOnce sync.Once
}

var _ dvoice.OpusFrameProvider = (*opusProvider)(nil)

func (p *opusProvider) ProvideOpusFrame() ([]byte, error) {
	frame, ok := <-p.frames
	if !ok {
		p.drainOnce.Do(func() { close(p.drained) })
		return nil, io.EOF
	}
	return frame, nil
}

func (p *opusProvider) Close() {
	p.cancel()
}
