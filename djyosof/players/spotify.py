import re
from collections.abc import Callable

import discord
from discord import VoiceClient
from librespot.core import Session
from librespot.metadata import TrackId
from librespot.audio.decoders import AudioQuality, VorbisOnlyAudioQuality
import requests

from settings import CONFIG
from djyosof.audio_types.spotify import SpotifyTrack


class SpotifySource:
    def __init__(self):
        self.session_builder = Session.Builder().stored_file()
        if not self.session_builder.login_credentials:
            self.session_builder = Session.Builder().user_pass(
                CONFIG.get("spotify_user"), CONFIG.get("spotify_pass")
            )
        self.session = self.session_builder.create()

    def load_track(self, track: SpotifyTrack):
        track_id = TrackId.from_uri(f"spotify:track:{track.track_id}")  # anti-hero
        try:
            stream = self.session.content_feeder().load(
                track_id, VorbisOnlyAudioQuality(AudioQuality.VERY_HIGH), False, None
            )
        except:  # TODO: catch specific exceptions, possible decorator for this functionality
            # retry after creating a new session
            self.session = self.session_builder.create()
            stream = self.session.content_feeder().load(
                track_id, VorbisOnlyAudioQuality(AudioQuality.VERY_HIGH), False, None
            )

        return discord.FFmpegOpusAudio(
            source=stream.input_stream.stream(),
            bitrate=320,
            pipe=True,
        )

    def open_link(self, link: str) -> list[SpotifyTrack]:
        pattern = re.compile(
            r"https://open.spotify.com/(track|album|playlist)/(.{22}).*"
        )
        matcher = pattern.search(link)

        # media doesn't exist
        if matcher is None:
            return []

        media_type = matcher.group(1)
        media_id = matcher.group(2)

        try:
            token = self.session.tokens().get("user-read-email")
        except:
            self.session = self.session_builder.create()
            token = self.session.tokens().get("user-read-email")

        resp = requests.get(
            f"https://api.spotify.com/v1/{media_type}s/{media_id}",
            headers={"Authorization": f"Bearer {token}"},
        )
        if media_type == "track":
            tracks = [SpotifyTrack(resp.json())]
        elif media_type == "album":
            album_json = resp.json()
            del album_json["tracks"]
            tracks_json = resp.json()["tracks"]["items"]
            tracks = []
            for item in tracks_json:
                item["album"] = album_json
                tracks.append(SpotifyTrack(item))
        elif media_type == "playlist":
            tracks = [
                SpotifyTrack(item["track"]) for item in resp.json()["tracks"]["items"]
            ]

        return tracks

    def search(self, query: str) -> list[SpotifyTrack]:
        try:
            token = self.session.tokens().get("user-read-email")
        except:
            # retry after creating new session
            self.session = self.session_builder.create()
            token = self.session.tokens().get("user-read-email")

        resp = requests.get(
            "https://api.spotify.com/v1/search",
            {
                "limit": "5",
                "offset": "0",
                "q": query,
                "type": "track",
            },
            headers={"Authorization": f"Bearer {token}"},
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
