import logging
import sys

from djyosof.bot import DJYosof
from djyosof.cogs.system import SystemCog
from djyosof.cogs.spotify import SpotifyCog
from djyosof.cogs.youtube import YoutubeCog
from djyosof.cogs.audio_player import AudioPlayerCog
from settings import CONFIG

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
handler = logging.StreamHandler(stream=sys.stdout)
logger.addHandler(handler)


def handle_exception(exc_type, exc_value, exc_traceback):
    if issubclass(exc_type, KeyboardInterrupt):
        sys.__excepthook__(exc_type, exc_value, exc_traceback)
        return

    logger.error("Uncaught exception", exc_info=(exc_type, exc_value, exc_traceback))


sys.excepthook = handle_exception


def main():
    bot = DJYosof(command_prefix="/")
    bot.add_cog(SpotifyCog(bot))
    bot.add_cog(YoutubeCog(bot))
    bot.add_cog(AudioPlayerCog(bot))
    bot.add_cog(SystemCog(bot))
    bot.run(CONFIG.get("discord_token"))


if __name__ == "__main__":
    main()
