import discord
from discord import Interaction, VoiceClient, Option
from discord.ext import commands
from discord.commands import slash_command

from djyosof.cogs import utilities
from settings import CONFIG


class SystemCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def hello(self, interaction: Interaction):
        """Says hello!"""
        await interaction.response.send_message(f"Hi, {interaction.user.mention}")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def join(self, interaction: Interaction):
        self.bot.voice_client = await utilities.connect_or_move(interaction)

        await interaction.response.send_message(
            f"Joining: {interaction.user.voice.channel}"
        )

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def leave(self, interaction: Interaction):
        await utilities.leave(interaction)
        await interaction.response.send_message(f"Left voice channel.")
