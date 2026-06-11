# dj-yosof

A Discord music bot that plays audio from Spotify and YouTube in voice channels.
This is a Go rewrite of the original Python project.

## Requirements

- **Go** 1.26+
- A **C toolchain** + **ffmpeg** + codec libraries (libogg, libvorbis, libFLAC,
  plus ALSA on Linux and libopus on non-amd64) — this is a cgo program. Discord's
  DAVE end-to-end-encrypted voice (required since March 2026) is handled by a
  **pure-Go** implementation, so no extra native library is needed for it. See
  [BUILDING.md](BUILDING.md) for the exact per-platform packages and server steps.
- A **Discord bot token** with the *message content* and *voice* privileged intents,
  and permissions to read/send messages and connect/speak in voice channels.
- A **Spotify Premium** account (required to stream Spotify audio).

Dependencies are vendored under `vendor/`, so builds need no module downloads.

## Setup

1. Install the build dependencies — see [BUILDING.md](BUILDING.md).
2. Copy `config.yaml.example` to `config.yaml` and fill it out.
3. Build: `go build .`

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
once.

The authorization redirects to a one-off callback server at
`http://127.0.0.1:<port>/login` **on the machine running the bot**. On a
headless/remote server that address won't be reachable from your laptop's
browser, so use one of:

- **Authorize locally, copy the file (simplest):** run the bot on a machine with
  a browser, complete the login (the callback is local, so it works), then copy
  the generated `spotify_credentials.json` to the server next to its
  `config.yaml` (`scp spotify_credentials.json user@server:/path/to/dj-yosof/`).
  The credentials include the device id, so they're portable.
- **SSH port-forward:** set a fixed `spotify_oauth_callback_port` (e.g. `8888`),
  connect with `ssh -L 8888:127.0.0.1:8888 user@server`, then open the printed
  URL in your laptop browser — the redirect tunnels back to the server.

### Spotify search & links (developer app)

`/spotify` **search** and **track/album/playlist links** use Spotify's Web API.
Spotify now rejects Web API requests made with go-librespot's session token with
persistent `429 Too Many Requests` errors
([go-librespot#282](https://github.com/devgianlu/go-librespot/issues/282)), so
you must supply your own **Spotify developer app** credentials:

1. Create an app at <https://developer.spotify.com/dashboard>. The dashboard
   requires a **Redirect URI** even though the Client Credentials flow never
   uses one — enter any placeholder, e.g. `http://127.0.0.1:8888/callback`
   (Spotify requires `127.0.0.1` rather than `localhost`). Tick **Web API**
   when asked which APIs you'll use.
2. Put its **Client ID** and **Client Secret** into `spotify_client_id` /
   `spotify_client_secret` in `config.yaml`.

Audio **streaming** does not use the Web API, so it works without this. If you
leave the credentials empty, search/links fall back to the session token and
will likely hit 429.

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
