from discord import AudioSource
from librespot.core import Session
from librespot.metadata import TrackId
from librespot.audio.decoders import AudioQuality, VorbisOnlyAudioQuality

from settings import CONFIG


class SpotifySource(AudioSource):
    def __init__(self):
        self.session = (
            Session.Builder()
            .user_pass(CONFIG.get("spotify_user"), CONFIG.get("spotify_pass"))
            .create()
        )
        self.stream = None

    def load_track(self, track_id: str):
        track_id = TrackId.from_uri(f"spotify:track:{track_id}")  # anti-hero
        self.stream = self.session.content_feeder().load(
            track_id, VorbisOnlyAudioQuality(AudioQuality.VERY_HIGH), False, None
        )

    def read(self):
        return self.stream.input_stream.stream().read()
