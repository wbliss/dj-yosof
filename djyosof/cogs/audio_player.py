import discord
from discord import ApplicationContext
from discord.commands import slash_command
from discord.ext import commands

from djyosof.cogs import utilities
from settings import CONFIG


class AudioPlayerCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def pause(self, ctx: ApplicationContext):
        pass

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def queue(self, ctx: ApplicationContext):
        # TODO pagination
        embed = discord.Embed(
            title="",
            color=discord.Colour.blurple(),
        )

        queue_markdown = ""
        for idx, track in enumerate(
            (
                [self.bot.audio_players[ctx.guild_id].now_playing]
                + list(self.bot.audio_players[ctx.guild_id].queue._queue)
            )[:10]
        ):
            queue_markdown += f"**{idx+1}**. {track.get_display_name()}"
            if idx == 0:
                queue_markdown += " - NOW PLAYING"
            queue_markdown += "\n"

        if queue_markdown == "":
            queue_markdown = "Queue is empty!"
        else:
            queue_length = (
                len(list(self.bot.audio_players[ctx.guild_id].queue._queue)) + 1
            )
            queue_markdown += f"\nShowing {min(10, queue_length)} out of {queue_length} tracks in the queue."

        embed.add_field(name="Queue", value=queue_markdown)
        await ctx.respond("", embed=embed)

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def skip(self, ctx: ApplicationContext):
        audio_player = self.bot.audio_players[ctx.guild_id]
        voice = await utilities.connect_or_move(ctx)
        if not voice:
            await ctx.respond("Unable to connect to a voice channel :(")
        audio_player.skip(voice)
        await ctx.respond("Song skipped!")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def stop(self, ctx: ApplicationContext):
        audio_player = self.bot.audio_players[ctx.guild_id]
        voice = await utilities.connect_or_move(ctx)
        if not voice:
            await ctx.respond("Unable to connect to a voice channel :(")
        audio_player.stop(voice)
        await ctx.respond("Queue cleared and player stopped.")
