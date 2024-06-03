from discord import ApplicationContext
from discord.commands import slash_command
from discord.ext import commands

from djyosof.cogs import utilities
from settings import CONFIG


class SystemCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def hello(self, ctx: ApplicationContext):
        await ctx.respond(f"Hi, {ctx.user.mention}")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def join(self, ctx: ApplicationContext):
        voice = await utilities.connect_or_move(ctx)
        if not voice:
            await ctx.respond(f"Unable to join: {ctx.user.voice.channel}")
            return

        await ctx.respond(f"Joining: {ctx.user.voice.channel}")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def leave(self, ctx: ApplicationContext):
        await utilities.leave(ctx)
        await ctx.respond("Left voice channel.")
