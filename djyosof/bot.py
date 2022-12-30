import discord

class DJYosof(discord.Client):
    """
    The bot that will replace Yosof
    """
    def __init__(self):
        intents = discord.Intents.default()
        intents.message_content = True
        super().__init__(intents=intents)

    async def on_ready(self):
        print(f"We have logged in as {self.user}")

    async def on_message(self, message):
        if message.author == self.user:
            return

        if message.content.startswith("$hello"):
            await message.channel.send("Hello!")
