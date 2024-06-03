from datetime import timedelta

import discord
from pytube import YouTube

from djyosof.audio_types.playable_audio import AudioType, PlayableAudio


class YoutubeTrack(PlayableAudio):
    def __init__(self, video: dict | None = None):
        if not video:
            return

        self.title = video["title"]
        self.thumbnail_url = video["thumbnail_url"]
        self.video_length = video["video_length"]
        self.watch_url = video["watch_url"]

    @staticmethod
    def from_pytube(video: YouTube):
        yt = YoutubeTrack()
        yt.title = video.title
        yt.thumbnail_url = video.thumbnail_url
        yt.video_length = video.length
        yt.watch_url = video.watch_url

        return yt

    def get_embed(self):
        embed = discord.Embed(
            title="Now Playing",
            color=discord.Colour.blurple(),
        )
        embed.add_field(name="Title", value=f"{self.get_display_name()}", inline=True)
        embed.set_image(url=self.thumbnail_url)
        return embed

    def get_display_name(self):
        length = timedelta(seconds=self.video_length)
        return f"[{self.title} ({str(length)})]({self.watch_url})"

    def get_type(self):
        return AudioType.YOUTUBE
