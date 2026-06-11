# Building dj-yosof

dj-yosof is a **cgo** program: it links/compiles C code for the audio path
(Opus via `layeh.com/gopus`, Ogg/Vorbis via `github.com/xlab/vorbis-go`, plus
go-librespot's FLAC/output backends). So a build host needs a C toolchain and a
few `-dev` libraries in addition to Go.

> Discord requires the DAVE end-to-end-encryption protocol on all voice
> connections (enforced since March 2026). The bot uses
> [disgo](https://github.com/disgoorg/disgo) with the **pure-Go**
> [dave-go](https://github.com/thomas-vilte/dave-go) DAVE implementation, so
> **no extra native library is needed** for E2EE voice (no libdave/C++).

Dependencies are **vendored** (committed under `vendor/`), so no module
downloads are required at build time — only the system libraries below.

## Requirements

- **Go 1.26+** (see the `go` directive in `go.mod`). Ubuntu's `apt` Go is
  usually too old — install from <https://go.dev/dl/>.
- A **C compiler** (`gcc`/`clang`) and **pkg-config** — for cgo.
- **libogg** and **libvorbis** development headers — `vorbis-go` links them via
  `pkg-config: ogg vorbis vorbisenc`.
- **libFLAC** development headers — go-librespot's `player` package links it
  (`pkg-config: flac`) even though dj-yosof only streams Vorbis.
- **ALSA** development headers on **Linux** (`pkg-config: alsa`) — same reason
  (go-librespot's audio `output` backend). On macOS this is CoreAudio, built in.
- **libopus** development headers — **only on non-amd64** (e.g. arm64). On
  amd64/386, `gopus` compiles its own bundled libopus and does not need a
  system copy.
- **ffmpeg** on `PATH` — used at runtime to transcode audio.

> Note: libFLAC and ALSA/CoreAudio are required only because go-librespot's
> `player` package links them unconditionally; dj-yosof does not use FLAC
> playback or the local audio output backends.

## Ubuntu / Debian (amd64)

```sh
# System dependencies (libopus NOT needed on amd64 — gopus bundles it)
sudo apt-get update
sudo apt-get install -y build-essential pkg-config ffmpeg git \
  libogg-dev libvorbis-dev libflac-dev libasound2-dev

# Go 1.26 (apt's Go is too old; go.mod requires 1.26)
curl -fsSL https://go.dev/dl/go1.26.0.linux-amd64.tar.gz -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH=$PATH:/usr/local/go/bin   # add to ~/.profile to persist

# Clone and build (vendored deps — no network needed for modules)
git clone https://github.com/GusPrice/dj-yosof.git
cd dj-yosof
go build -mod=vendor -o dj-yosof .

# Configure and run
cp config.yaml.example config.yaml   # edit: discord_token, guild_ids
./dj-yosof
```

### Ubuntu on arm64 (e.g. AWS Graviton)

Same as above, plus the system Opus library (arm64 links it instead of
compiling the bundle), and use the arm64 Go tarball:

```sh
sudo apt-get install -y libopus-dev
curl -fsSL https://go.dev/dl/go1.26.0.linux-arm64.tar.gz -o /tmp/go.tgz
```

## macOS

```sh
brew install go ffmpeg pkg-config libogg libvorbis flac opus
go build -o dj-yosof .
```

## Notes

- **CGO must be enabled.** It is by default once a C compiler is present. If you
  have globally set `CGO_ENABLED=0`, build with `CGO_ENABLED=1 go build ...`.
- **Headless Spotify login.** On first run the bot prints a Spotify OAuth2 URL.
  On a server with no browser, open that URL on another machine, authorize, and
  copy the generated credentials file (`spotify_credentials.json` by default) to
  the server — it is reused afterward. YouTube works without any of this.
- **Regenerating the vendor tree.** `go mod vendor` omits nested directories
  that contain no Go files, which strips the bundled C sources from `gopus`
  (`vendor/layeh.com/gopus/opus-1.1.2/`) and the headers from `vorbis-go`
  (`vendor/github.com/xlab/vorbis-go/vorbis/{ogg,vorbis}/`). If you re-run
  `go mod vendor`, re-copy those directories from the module cache
  (`$(go env GOMODCACHE)/...`) or the build will fail with
  `opus-1.1.2/config.h: No such file or directory`.
