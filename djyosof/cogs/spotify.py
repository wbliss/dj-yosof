import discord
from discord import Interaction, VoiceChannel, Option
from discord.ext import commands
from discord.commands import slash_command

from settings import CONFIG
from djyosof.players.spotify import SpotifySource


class SpotifyCog(commands.Cog):
    def __init__(self, bot) -> None:
        self.bot = bot

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def hello(self, interaction: Interaction):
        """Says hello!"""
        await interaction.response.send_message(f"Hi, {interaction.user.mention}")

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def join(self, interaction: Interaction):
        await self._connect_or_move(interaction)
        await interaction.response.send_message(
            f"Joining: {interaction.user.voice.channel}"
        )

    @slash_command(guild_ids=CONFIG.get("guild_ids"))
    async def leave(self, interaction: Interaction):
        await self._leave(interaction)
        await interaction.response.send_message(f"Left voice channel.")

    @slash_command(
        guild_ids=CONFIG.get("guild_ids"),
    )
    async def play(
        self,
        interaction: Interaction,
        query: Option(str, "Query to search for", required=True),
    ):
        voice = await self._connect_or_move(interaction)
        if not voice:
            await interaction.response.send_message(
                "Unable to connect to a voice channel :("
            )

        spotify = SpotifySource()
        tracks = spotify.search(query)
        spotify.load_track(tracks[0].track_id)

        voice.play(spotify.get_audio())
        await interaction.response.send_message(
            "Playing music!", embed=tracks[0].get_embed()
        )

    # Helpers
    async def _connect_or_move(
        self, interaction: Interaction, *args, **kwargs
    ) -> VoiceChannel | None:
        author_voice = interaction.user.voice
        # yeah this won't work
        if not author_voice:
            await interaction.response.send_message("Join a voice channel first.")

        author_voice_channel = author_voice.channel

        # Not connected anywhere, connect
        current_voice_client = interaction.guild.voice_client
        if not current_voice_client:
            print(f"Joining: {author_voice_channel}")
            return await author_voice_channel.connect(*args, **kwargs)

        # If we're already in a channel for that guild check to see
        # if we need to move channels or do nothing
        current_voice_channel = current_voice_client.channel
        if author_voice_channel == current_voice_channel:
            print(f"Already in {author_voice_channel}, not joining")
            return

        print(f"Joining: {author_voice_channel}")
        return await current_voice_client.move_to(author_voice_channel)

    async def _leave(self, interaction: Interaction):
        current_voice_client = interaction.guild.voice_client
        await current_voice_client.disconnect()
