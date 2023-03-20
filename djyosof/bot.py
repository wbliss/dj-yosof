import logging
from collections import defaultdict

import discord
from discord.ext import commands

from djyosof.players.audio_player import AudioPlayer
from settings import CONFIG


class DJYosof(commands.Bot):
    """
    The bot that will replace Yosof
    """

    def __init__(self, command_prefix):
        intents = discord.Intents.default()
        intents.message_content = True
        commands.Bot.__init__(self, intents=intents, command_prefix=command_prefix)
        self.players = {}
        self.audio_players: defaultdict[int, AudioPlayer] = defaultdict(
            lambda: AudioPlayer(bot=self)
        )

        try:
            discord.opus.load_opus(CONFIG.get("opus_location"))
        except OSError:
            logging.warning(
                "Opus library location invalid, voice commands will not work"
            )

    async def on_ready(self):
        logging.info(f"We have logged in as {self.user}")
