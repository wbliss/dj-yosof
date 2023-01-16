from enum import Enum


class AudioType(Enum):
    SPOTIFY = "spotify"
    YOUTUBE = "youtube"


class PlayableAudio:
    def __init__(self):
        raise NotImplementedError()

    def get_embed(self):
        raise NotImplementedError()

    def get_type(self):
        raise NotImplementedError()
