from enum import Enum

import discord


class AudioType(Enum):
    spotify = "spotify"
    youtube = "youtube"


class PlayableAudio:
    def __init__(self):
        raise NotImplementedError()

    def get_embed(self):
        raise NotImplementedError()

    def get_type(self):
        raise NotImplementedError()
