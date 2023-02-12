from urllib.parse import urlparse, parse_qs
from collections.abc import Callable
from io import BytesIO
from pytube import Search, YouTube, Playlist

import discord
from discord import VoiceClient

from djyosof.audio_types.youtube import YoutubeTrack


class YoutubeSource:
    def load_track(self, track: YoutubeTrack):
        FFMPEG_OPTS = {
            "before_options": "-reconnect 1 -reconnect_streamed 1 -reconnect_delay_max 5",
            "options": "-vn",
        }
        stream = track.video.streams.get_audio_only()
        return discord.FFmpegPCMAudio(stream.url, **FFMPEG_OPTS)

    def open_link(self, link: str) -> list[YoutubeTrack]:
        parsed_url = urlparse(link)
        query = parse_qs(parsed_url.query)

        # Playlist
        if (
            parsed_url.path == "/watch"
            and "list" in query.keys()
            or parsed_url.path == "/playlist"
        ):
            playlist = Playlist(link)
            # TODO: fix this
            # return [YoutubeTrack(video) for video in playlist.videos]
            return []
        elif parsed_url.path == "/watch":
            track = YoutubeTrack(YouTube(link))
            return [track]
        else:
            # unrecognized link
            return []

    def search(self, query: str):
        return [YoutubeTrack(result) for result in Search(query).results[:5]]

    def play(
        self,
        track: YoutubeTrack,
        voice: VoiceClient,
        after: Callable | None = None,
    ):
        audio = self.load_track(track)
        voice.play(audio, after=after)
