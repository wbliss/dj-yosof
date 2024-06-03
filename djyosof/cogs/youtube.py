import re

import discord
from discord import ApplicationContext, Option
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
        ctx: ApplicationContext,
        query: Option(str, "Query to search for", required=True),
    ):
        pattern = re.compile(r"https://(www.)?youtube.com/.+")
        matcher = pattern.search(query)

        # media doesn't exist
        if matcher:
            tracks = self.bot.players[AudioType.YOUTUBE].open_link(query)
            voice = await utilities.connect_or_move(ctx)
            if not voice:
                await ctx.respond("Unable to connect to a voice channel :(")
                return

            if not tracks:
                await ctx.respond(
                    "No video found. NOTE: Playlist functionality is very buggy."
                )
                return

            await ctx.respond(f"Added {len(tracks)} tracks to the queue")
            await self.bot.audio_players[ctx.guild_id].enqueue_and_play(
                tracks[0], voice, ctx
            )
            for track in tracks[1:]:
                await self.bot.audio_players[ctx.guild_id].enqueue(track, ctx)

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
            await ctx.respond("", embed=embed, view=view)
