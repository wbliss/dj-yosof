import discord
from discord import Interaction
from discord.ext import commands
from discord.commands import slash_command
from settings import CONFIG

from djyosof.cogs import utilities


class AudioPlayerCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def pause(self, interaction: Interaction):
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
            list(self.bot.audio_players[interaction.guild_id].queue._queue)[:10]
        ):
            queue_markdown += f"**{idx+1}**. {track.get_display_name()}\n"

        if queue_markdown == "":
            queue_markdown = "Queue is empty!"
        else:
            queue_length = len(
                list(self.bot.audio_players[interaction.guild_id].queue._queue)
            )
            queue_markdown += f"\nShowing {max(10, queue_length)} out of {queue_length} tracks in the queue."

        embed.add_field(name="Queue", value=queue_markdown)
        await interaction.response.send_message("Current Queue", embed=embed)

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def skip(self, interaction: Interaction):
        audio_player = self.bot.audio_players[interaction.guild_id]
        voice = await utilities.connect_or_move(interaction)
        if not voice:
            await interaction.response.send_message(
                "Unable to connect to a voice channel :("
            )
        audio_player.skip(voice)
        await interaction.response.send_message("Song skipped!")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def stop(self, interaction: Interaction):
        audio_player = self.bot.audio_players[interaction.guild_id]
        voice = await utilities.connect_or_move(interaction)
        if not voice:
            await interaction.response.send_message(
                "Unable to connect to a voice channel :("
            )
        audio_player.stop(voice)
        await interaction.response.send_message("Queue cleared and player stopped.")
