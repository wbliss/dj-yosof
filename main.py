from djyosof.bot import DJYosof
from settings import CONFIG

tha_dj = DJYosof(command_prefix="/")
tha_dj.run(CONFIG.get("discord_token"))