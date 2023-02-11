import discord
from discord import Interaction
from discord.ext import commands
from discord.commands import slash_command
from settings import CONFIG


class AudioPlayerCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def play(self, interaction: Interaction):
        pass

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def queue(self, interaction: Interaction):
        # TODO pagination
        embed = discord.Embed(
            title="",
            color=discord.Colour.blurple(),
        )

        queue_markdown = ""
        for idx, track in enumerate(
            list(self.bot.audio_players[interaction.guild_id].queue._queue)
        ):
            queue_markdown += f"**{idx+1}**. {track.get_display_name()}\n"

        if queue_markdown == "":
            queue_markdown = "Queue is empty!"

        embed.add_field(name="Queue", value=queue_markdown)
        await interaction.response.send_message("Current Queue", embed=embed)
