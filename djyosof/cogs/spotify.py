import re

import discord
from discord import Interaction, Option
from discord.ext import commands
from discord.commands import slash_command

from djyosof.audio_types.playable_audio import AudioType
from djyosof.cogs import utilities
from djyosof.players.spotify import SpotifySource
from djyosof.views.search_view import SearchView
from settings import CONFIG


class SpotifyCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot
        self.bot.players[AudioType.SPOTIFY] = SpotifySource()

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def spotify(
        self,
        interaction: Interaction,
        query: Option(str, "Query to search for", required=True),
    ):
        pattern = re.compile(
            r"https://open.spotify.com/(track|album|playlist)/(.{22}).*"
        )
        matcher = pattern.search(query)

        # media doesn't exist
        if matcher:
            tracks = self.bot.players[AudioType.SPOTIFY].open_link(query)
            voice = await utilities.connect_or_move(interaction)
            if not voice:
                await interaction.response.send_message(
                    "Unable to connect to a voice channel :("
                )
                return
            await self.bot.audio_players[interaction.guild_id].enqueue_and_play(
                tracks[0], voice, interaction
            )
            for track in tracks[1:]:
                await self.bot.audio_players[interaction.guild_id].enqueue(
                    track, interaction
                )

            await interaction.response.send_message(
                f"Added {len(tracks)} tracks to the queue"
            )

        else:
            tracks = self.bot.players[AudioType.SPOTIFY].search(query)

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
