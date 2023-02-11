from collections.abc import Callable
from io import BytesIO
from pytube import Search, YouTube

import discord
from discord import VoiceClient

from djyosof.audio_types.youtube import YoutubeTrack


class YoutubeSource:
    def load_track(self, track: YoutubeTrack):
        bytes_stream = BytesIO()
        stream = track.video.streams.get_audio_only()
        stream.stream_to_buffer(bytes_stream)

        return discord.FFmpegOpusAudio(
            source=bytes_stream,
            bitrate=stream.abr,
            pipe=True,
        )

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
