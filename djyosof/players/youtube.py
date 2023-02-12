from collections.abc import Callable
from io import BytesIO
from pytube import Search, YouTube

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
