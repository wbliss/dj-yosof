import discord
from discord import AudioSource
from librespot.core import Session, SearchManager
from librespot.metadata import TrackId
from librespot.audio.decoders import AudioQuality, VorbisOnlyAudioQuality
import requests

from settings import CONFIG


class SpotifySource:
    def __init__(self):
        session_builder = Session.Builder().stored_file()
        if not session_builder.login_credentials:
            session_builder = Session.Builder().user_pass(
                CONFIG.get("spotify_user"), CONFIG.get("spotify_pass")
            )
        self.session = session_builder.create()

        self.stream = None

    def load_track(self, track_id: str):
        track_id = TrackId.from_uri(f"spotify:track:{track_id}")  # anti-hero
        self.stream = self.session.content_feeder().load(
            track_id, VorbisOnlyAudioQuality(AudioQuality.VERY_HIGH), False, None
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

    def get_audio(self):
        # No track loaded
        if not self.stream:
            return None

        return discord.FFmpegOpusAudio(
            source=self.stream.input_stream.stream(),
            bitrate=320,
            pipe=True,
        )


class SpotifyTrack:
    def __init__(self, search_response: dict):
        self.name = search_response["name"]
        self.artist = search_response["artists"][0]["name"]
        self.album = search_response["album"]["name"]
        self.album_art_url = search_response["album"]["images"][1]["url"]
        self.track_id = search_response["id"]

    def get_embed(self):
        embed = discord.Embed(
            title="Now Playing",
            color=discord.Colour.blurple(),
        )
        embed.add_field(name="Track", value=self.name, inline=True)
        embed.add_field(name="Artist", value=self.artist, inline=True)
        embed.add_field(name="Album", value=self.album, inline=True)
        embed.set_image(url=self.album_art_url)
        return embed
