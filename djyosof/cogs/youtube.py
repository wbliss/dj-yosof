import discord
from discord import Interaction, Option
from discord.ext import commands
from discord.commands import slash_command

from djyosof.audio_types.playable_audio import AudioType
from djyosof.players.youtube import YoutubeSource
from djyosof.views.search_view import SearchView
from settings import CONFIG


class YoutubeCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot
        self.bot.players[AudioType.YOUTUBE] = YoutubeSource()

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def yt(
        self,
        interaction: Interaction,
        query: Option(str, "Query to search for", required=True),
    ):
        tracks = self.bot.players[AudioType.YOUTUBE].search(query)

        embed = discord.Embed(
            title="",
            color=discord.Colour.blurple(),
        )

        tracklist_markdown = ""
        for idx, track in enumerate(tracks):
            tracklist_markdown += f"**{idx+1}**. {track.title} ({track.video_length})\n"

        embed.add_field(name="Search Results", value=tracklist_markdown)

        view = SearchView(tracks, self.bot)
        await interaction.response.send_message("", embed=embed, view=view)
