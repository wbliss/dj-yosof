from djyosof.bot import DJYosof
from djyosof.cogs.system import SystemCog
from djyosof.cogs.spotify import SpotifyCog
from djyosof.cogs.audio_player import AudioPlayerCog
from settings import CONFIG

bot = DJYosof(command_prefix="/")
bot.add_cog(SpotifyCog(bot))
bot.add_cog(AudioPlayerCog(bot))
bot.add_cog(SystemCog(bot))
bot.run(CONFIG.get("discord_token"))
