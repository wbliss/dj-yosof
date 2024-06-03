import abc
from collections.abc import Callable

from discord import VoiceClient

from djyosof.audio_types.playable_audio import PlayableAudio


class BaseSource:
    __metaclass__ = abc.ABCMeta

    @abc.abstractmethod
    def play(
        self,
        track: PlayableAudio,
        voice: VoiceClient,
        after: Callable | None = None,
    ):
        return

    @abc.abstractmethod
    def search(self, query: str):
        return
