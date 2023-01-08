import discord

from djyosof.audio_types.playable_audio import AudioType, PlayableAudio


class SpotifyTrack(PlayableAudio):
    def __init__(self, search_response: dict):
        self.name = search_response["name"]
        self.artist = search_response["artists"][0]["name"]
        self.album = search_response["album"]["name"]
        self.album_art_url = search_response["album"]["images"][1]["url"]
        self.track_id = search_response["id"]
        self.audio_stream = None

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

    def get_type(self):
        return AudioType.spotify
