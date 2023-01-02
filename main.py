from djyosof.bot import DJYosof
from djyosof.cogs.spotify import SpotifyCog
from settings import CONFIG

bot = DJYosof(command_prefix="/")
bot.add_cog(SpotifyCog(bot))
bot.run(CONFIG.get("discord_token"))
