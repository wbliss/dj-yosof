import re

import discord
from discord import ApplicationContext, Option
from discord.commands import slash_command
from discord.ext import commands

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
        ctx: ApplicationContext,
        query: Option(str, "Query to search for", required=True),
    ):
        pattern = re.compile(
            r"https://open.spotify.com/(track|album|playlist)/(.{22}).*"
        )
        matcher = pattern.search(query)

        # media doesn't exist
        if matcher:
            tracks = self.bot.players[AudioType.SPOTIFY].open_link(query)
            voice = await utilities.connect_or_move(ctx)
            if not voice:
                await ctx.respond("Unable to connect to a voice channel :(")
                return
            await self.bot.audio_players[ctx.guild_id].enqueue_and_play(
                tracks[0], voice, ctx
            )
            for track in tracks[1:]:
                await self.bot.audio_players[ctx.guild_id].enqueue(track, ctx)

            await ctx.respond(f"Added {len(tracks)} tracks to the queue")

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
            await ctx.respond("", embed=embed, view=view)
