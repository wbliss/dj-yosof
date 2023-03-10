import re

import discord
from discord import Interaction, Option
from discord.ext import commands
from discord.commands import slash_command

from djyosof.audio_types.playable_audio import AudioType
from djyosof.cogs import utilities
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
        pattern = re.compile(r"https://(www.)?youtube.com/.+")
        matcher = pattern.search(query)

        # media doesn't exist
        if matcher:
            tracks = self.bot.players[AudioType.YOUTUBE].open_link(query)
            voice = await utilities.connect_or_move(interaction)
            if not voice:
                await interaction.response.send_message(
                    "Unable to connect to a voice channel :("
                )
                return

            if not tracks:
                await interaction.response.send_message(
                    "No video found. NOTE: Playlist functionality is very buggy."
                )
                return

            await interaction.response.send_message(
                f"Added {len(tracks)} tracks to the queue"
            )
            await self.bot.audio_players[interaction.guild_id].enqueue_and_play(
                tracks[0], voice, interaction
            )
            for track in tracks[1:]:
                await self.bot.audio_players[interaction.guild_id].enqueue(
                    track, interaction
                )

        else:
            tracks = self.bot.players[AudioType.YOUTUBE].search(query)

            embed = discord.Embed(
                title="",
                color=discord.Colour.blurple(),
            )

            tracklist_markdown = ""
            for idx, track in enumerate(tracks):
                tracklist_markdown += f"**{idx+1}**. {track.get_display_name()}\n"

            embed.add_field(name="Search Results", value=tracklist_markdown)

            view = SearchView(tracks, self.bot)
            await interaction.response.send_message("", embed=embed, view=view)
