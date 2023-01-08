from collections.abc import Callable

import discord
from discord import VoiceClient
from librespot.core import Session, SearchManager
from librespot.metadata import TrackId
from librespot.audio.decoders import AudioQuality, VorbisOnlyAudioQuality
import requests

from settings import CONFIG
from djyosof.audio_types.spotify_track import SpotifyTrack


class SpotifySource:
    def __init__(self):
        session_builder = Session.Builder().stored_file()
        if not session_builder.login_credentials:
            session_builder = Session.Builder().user_pass(
                CONFIG.get("spotify_user"), CONFIG.get("spotify_pass")
            )
        self.session = session_builder.create()
        self.stream = None

    def load_track(self, track: SpotifyTrack):
        track_id = TrackId.from_uri(f"spotify:track:{track.track_id}")  # anti-hero
        stream = self.session.content_feeder().load(
            track_id, VorbisOnlyAudioQuality(AudioQuality.VERY_HIGH), False, None
        )

        return discord.FFmpegOpusAudio(
            source=stream.input_stream.stream(),
            bitrate=320,
            pipe=True,
        )

    def search(self, query: str):
        token = self.session.tokens().get("user-read-email")
        resp = requests.get(
            "https://api.spotify.com/v1/search",
            {
                "limit": "5",
                "offset": "0",
                "q": query,
                "type": "track",
            },
            headers={"Authorization": "Bearer %s" % token},
        )
        tracks = [SpotifyTrack(item) for item in resp.json()["tracks"]["items"]]
        return tracks

    def play(
        self,
        track: SpotifyTrack,
        voice: VoiceClient,
        after: Callable | None = None,
    ):
        audio = self.load_track(track)
        voice.play(audio, after=after)
