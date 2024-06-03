from discord import ApplicationContext
from discord.ext import commands
from discord.commands import slash_command

from djyosof.cogs import utilities
from settings import CONFIG


class SystemCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def hello(self, ctx: ApplicationContext):
        await ctx.interaction.response.send_message(
            f"Hi, {ctx.interaction.user.mention}"
        )

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def join(self, ctx: ApplicationContext):
        voice = await utilities.connect_or_move(interaction)
        if not voice:
            await ctx.interaction.response.send_message(
                f"Unable to join: {ctx.interaction.user.voice.channel}"
            )
            return

        await ctx.interaction.response.send_message(
            f"Joining: {ctx.interaction.user.voice.channel}"
        )

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def leave(self, ctx: ApplicationContext):
        await utilities.leave(ctx.interaction)
        await ctx.interaction.response.send_message("Left voice channel.")
