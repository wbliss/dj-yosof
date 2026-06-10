# dj-yosof

A Discord music bot that plays audio from Spotify and YouTube in voice channels.
This is a Go rewrite of the original Python project.

## Requirements

- **Go** 1.26+
- **ffmpeg** — audio transcoding
  - macOS: `brew install ffmpeg`
  - Linux: `apt install ffmpeg`
- **opus** — Opus codec (linked via cgo by the `gopus` encoder)
  - macOS: `brew install opus`
  - Linux: `apt install libopus-dev`
- A **Discord bot token** with the *message content* and *voice* privileged intents,
  and permissions to read/send messages and connect/speak in voice channels.
- A **Spotify Premium** account (required to stream Spotify audio).

## Setup

1. Copy `config.yaml.example` to `config.yaml` and fill it out.
2. Build: `go build .`

## Run

```sh
go run .            # or ./dj-yosof after `go build`
```

Pass a different config path with `-config /path/to/config.yaml`.

### First-time Spotify login

Spotify streaming uses [`go-librespot`](https://github.com/devgianlu/go-librespot),
which authenticates with **OAuth2** (Spotify removed username/password login).
On the first run the bot prints an authorization URL to the console — open it,
log in, and authorize. The resulting credentials are cached in
`spotify_credentials_file` and reused on subsequent runs, so you only do this
once. (If you run headless, complete the login on a machine with a browser and
copy the credentials file over.)

## Commands

| Command | Description |
|---|---|
| `/hello` | Say hello |
| `/join` | Join your voice channel |
| `/leave` | Leave the voice channel |
| `/spotify <query>` | Play a Spotify track/album/playlist link, or search Spotify |
| `/yt <query>` | Play a YouTube link, or search YouTube |
| `/queue` | Show the queue |
| `/skip` | Skip the current track |
| `/stop` | Stop playback and clear the queue |
| `/pause` | (not implemented) |

## Architecture

- `config/` — YAML config loading
- `audio/` — `PlayableAudio` track abstraction (Spotify/YouTube) and their embeds
- `player/` — audio sources and the per-guild queue + playback loop
  - `spotify.go` — go-librespot session, Web API metadata, decoded-PCM streaming
  - `youtube.go` — kkdai/youtube streams + scraped search
  - `audioplayer.go` — queue, play loop, now-playing message
- `voice/` — voice connect/move/leave and the ffmpeg → Opus → Discord pipeline
- `bot/` — Discord session, slash commands, interaction routing
- `views/` — search-result buttons
