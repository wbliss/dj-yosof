"""Contains class for view that displays search results"""
import discord

from djyosof.bot import DJYosof
from djyosof.audio_types.playable_audio import PlayableAudio
from djyosof.audio_types.spotify_track import SpotifyTrack
from djyosof.cogs import utilities


class SearchView(discord.ui.View):
    """
    Discord view that shows search results and
    allows user to choose one to add to queue
    """

    def __init__(
        self,
        tracks: list[SpotifyTrack],
        bot: DJYosof,
    ):
        super().__init__(timeout=30, disable_on_timeout=True)
        for idx, track in enumerate(tracks):
            self.add_item(SearchResultButton(idx + 1, track, bot))


class SearchResultButton(discord.ui.Button):
    def __init__(
        self,
        index: int,
        track: PlayableAudio,
        bot: DJYosof,
    ):
        self.track = track
        self.bot = bot
        super().__init__(label=str(index), style=discord.ButtonStyle.primary)

    async def callback(self, interaction: discord.Interaction):
        voice = await utilities.connect_or_move(interaction)
        if not voice:
            await interaction.response.send_message(
                "Unable to connect to a voice channel :("
            )
            return

        # Add to queue or just play if nothing in queue/playing
        await self.bot.audio_players[interaction.guild_id].enqueue_and_play(
            self.track, voice, interaction
        )
