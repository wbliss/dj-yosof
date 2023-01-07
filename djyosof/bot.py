from collections import defaultdict
from queue import Queue

import discord
from discord.ext import commands
import ffmpeg

from djyosof.audio_types.playable_audio import AudioType
from djyosof.players.spotify import SpotifySource
from settings import CONFIG


class DJYosof(commands.Bot):
    """
    The bot that will replace Yosof
    """

    def __init__(self, command_prefix):
        intents = discord.Intents.default()
        intents.message_content = True
        commands.Bot.__init__(self, intents=intents, command_prefix=command_prefix)
        self.players = {
            AudioType.spotify: SpotifySource(),
        }
        self.queues: defaultdict[int, Queue] = defaultdict(lambda: Queue())

        try:
            discord.opus.load_opus(CONFIG.get("opus_location"))
        except OSError:
            print("Opus library location invalid, voice commands will not work")

    async def on_ready(self):
        print(f"We have logged in as {self.user}")
