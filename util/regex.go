// Package util holds small shared helpers, including the URL regexes ported
// verbatim from the original Python cogs.
package util

import "regexp"

// SpotifyLinkRegexp matches Spotify track/album/playlist links and captures the
// media type and the 22-char base62 id. Ported from djyosof/cogs/spotify.py.
var SpotifyLinkRegexp = regexp.MustCompile(`https://open\.spotify\.com/(track|album|playlist)/(.{22})`)

// YoutubeLinkRegexp matches YouTube watch/playlist URLs. Ported from
// djyosof/cogs/youtube.py.
var YoutubeLinkRegexp = regexp.MustCompile(`https://(www\.)?youtube\.com/.+`)

// IsSpotifyLink reports whether s looks like a Spotify link.
func IsSpotifyLink(s string) bool {
	return SpotifyLinkRegexp.MatchString(s)
}

// IsYoutubeLink reports whether s looks like a YouTube link.
func IsYoutubeLink(s string) bool {
	return YoutubeLinkRegexp.MatchString(s)
}
