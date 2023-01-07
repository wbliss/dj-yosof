import discord
from discord import Interaction
from discord.ext import commands

from djyosof.audio_types.playable_audio import AudioType, PlayableAudio
from djyosof.audio_types.spotify_track import SpotifyTrack
from djyosof.cogs import utilities


class SearchView(discord.ui.View):
    def __init__(
        self,
        tracks: list[SpotifyTrack],
        bot: commands.Bot,
    ):
        super().__init__(timeout=30, disable_on_timeout=True)
        for idx, track in enumerate(tracks):
            self.add_item(SearchResultButton(idx + 1, track, bot))


class SearchResultButton(discord.ui.Button):
    def __init__(
        self,
        index: int,
        track: PlayableAudio,
        bot: commands.Bot,
    ):
        self.track = track
        self.bot = bot
        super().__init__(label=index, style=discord.ButtonStyle.primary)

    async def callback(self, interaction: discord.Interaction):
        voice = await utilities.connect_or_move(interaction)
        if not voice:
            await interaction.response.send_message(
                "Unable to connect to a voice channel :("
            )

        # Add to queue or just play if nothing in queue/playing
        await utilities.queue_or_play(self.bot, self.track, voice, interaction)
