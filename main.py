from djyosof.bot import DJYosof

from settings import CONFIG

tha_dj = DJYosof()
tha_dj.run(CONFIG.get("discord_token"))