import discord
from discord.ext import commands

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
        current_voice_client = message.guild.voice_client

        # Not connected anywhere, connect
        if not current_voice_client:
            print(f"Joining: {author_voice_channel}")
            return await author_voice_channel.connect(*args, **kwargs)

        current_voice_channel = current_voice_client.channel
        # If we're already in a channel for that guild check to see
        # if we need to move channels or do nothing
        if author_voice_channel == current_voice_channel:
            print(f"Already in {author_voice_channel}, not joining")
            return

        print(f"Joining: {author_voice_channel}")
        return await current_voice_client.move_to(author_voice_channel)

    async def _leave(self, message):
        current_voice_client = message.guild.voice_client
        await current_voice_client.disconnect()

    async def _heman(self):
        """
        you know what's coming
        """
        pass

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
