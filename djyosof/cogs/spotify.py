import discord
from discord import Interaction, VoiceClient, Option
from discord.ext import commands
from discord.commands import slash_command

from djyosof.audio_types.playable_audio import AudioType
from djyosof.players.spotify import SpotifySource
from djyosof.cogs import utilities
from djyosof.views.search_view import SearchView
from settings import CONFIG


class SpotifyCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot
        self.bot.players[AudioType.spotify] = SpotifySource()

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def play(
        self,
        interaction: Interaction,
        query: Option(str, "Query to search for", required=True),
    ):
        tracks = self.bot.players[AudioType.spotify].search(query)

        embed = discord.Embed(
            title="",
            color=discord.Colour.blurple(),
        )

        tracklist_markdown = ""
        for idx, track in enumerate(tracks):
            tracklist_markdown += f"**{idx+1}**. {track.name} - {track.artist}\n"

        embed.add_field(name="Search Results", value=tracklist_markdown)

        view = SearchView(tracks, self.bot)
        await interaction.response.send_message("", embed=embed, view=view)
