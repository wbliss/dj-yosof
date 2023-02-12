from datetime import timedelta

import discord

from djyosof.audio_types.playable_audio import AudioType, PlayableAudio
from pytube import YouTube


class YoutubeTrack(PlayableAudio):
    def __init__(self, video: YouTube):
        self.title = video.title
        self.thumbnail_url = video.thumbnail_url
        self.video_length = video.length
        self.video = video

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
        return f"[{self.title} ({str(length)})]({self.video.watch_url})"

    def get_type(self):
        return AudioType.YOUTUBE
