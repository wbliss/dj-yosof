import discord
from discord.ext import commands
import ffmpeg

class DJYosof(commands.Bot):
    """
    The bot that will replace Yosof
    """
    def __init__(self, command_prefix):
        intents = discord.Intents.default()
        intents.message_content = True
        commands.Bot.__init__(self, intents=intents, command_prefix=command_prefix)
        self.register_commands()

    async def _connect_or_move(self, message, *args, **kwargs):

        author_voice_channel = message.author.voice.channel

        # yeah this won't work
        if not author_voice_channel:
            return

        # Not connected anywhere, connect
        current_voice_client = message.guild.voice_client
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

    async def _leave(self, message):
        current_voice_client = message.guild.voice_client
        await current_voice_client.disconnect()

    async def _play(self, voice):
        """
        you know what's coming
        """
        voice.play(discord.FFmpegPCMAudio(source="/Users/willbliss/code/dj-yosof/static/testsound.mp3"))

    async def on_ready(self):
        print(f"We have logged in as {self.user}")

    def register_commands(self):
        @self.command(guild_ids=[460571766901964801], pass_context=True)
        async def join(ctx):
            await self._connect_or_move(ctx.message)

        @self.command(guild_ids=[460571766901964801], pass_context=True)
        async def hello(ctx):
            await ctx.send("Hello!")

        @self.command(guild_ids=[460571766901964801], pass_context=True)
        async def leave(ctx):
            await self._leave(ctx.message)

        @self.command(name="fuck off", guild_ids=[460571766901964801], pass_context=True)
        async def fuck_off(ctx):
            await self._leave(ctx.message)

        @self.command(guild_ids=[460571766901964801], pass_context=True)
        async def play(ctx):
            voice = await self._connect_or_move(ctx.message)
            if not voice:
                return await ctx.send("Unable to connect to a voice channel :(")
            await self._play(voice)
