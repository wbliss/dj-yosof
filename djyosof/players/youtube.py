import logging
import traceback
from collections.abc import Callable
from urllib.parse import parse_qs, urlparse

import discord
from discord import VoiceClient
from pytubefix import Playlist, Search, YouTube

from djyosof.audio_types.youtube import YoutubeTrack


class YoutubeSource:
    def load_track(self, track: YoutubeTrack):
        FFMPEG_OPTS = {
            "before_options": "-reconnect 1 -reconnect_streamed 1 -reconnect_delay_max 5",
            "options": "-vn",
        }

        yt = YouTube(track.watch_url, "WEB_EMBED")
        stream = yt.streams.get_audio_only()

        return discord.FFmpegPCMAudio(stream.url, **FFMPEG_OPTS)

    # TODO: burn this function to the ground
    def parse_playlist(self, playlist: Playlist):
        tracks = []

        videos = playlist.initial_data["contents"]["twoColumnBrowseResultsRenderer"][
            "tabs"
        ][0]["tabRenderer"]["content"]["sectionListRenderer"]["contents"][0][
            "itemSectionRenderer"
        ]["contents"][0]["playlistVideoListRenderer"]["contents"]
        for video in videos:
            if "playlistVideoRenderer" not in video:
                continue

            _video = video["playlistVideoRenderer"]
            track_data = {
                "title": " ".join([run["text"] for run in _video["title"]["runs"]]),
                "thumbnail_url": _video["thumbnail"]["thumbnails"][-1]["url"],
                "video_length": int(_video["lengthSeconds"]),
                "watch_url": f"https://www.youtube.com/watch?v={_video['videoId']}",
            }
            tracks.append(YoutubeTrack(track_data))
        return tracks

    def open_link(self, link: str) -> list[YoutubeTrack]:
        parsed_url = urlparse(link)
        query = parse_qs(parsed_url.query)

        # Playlist
        if (
            parsed_url.path == "/watch"
            and "list" in query.keys()
            or parsed_url.path == "/playlist"
        ):
            try:
                return self.parse_playlist(Playlist(link))
            except KeyError:
                logging.info(traceback.format_exc())
                return []
        elif parsed_url.path == "/watch":
            return [YoutubeTrack.from_pytube(YouTube(link, "WEB_EMBED"))]
        else:
            # unrecognized link
            return []

    def search(self, query: str) -> list[YoutubeTrack]:
        try:
            return [
                YoutubeTrack.from_pytube(result) for result in Search(query).videos[:5]
            ]

        except TypeError as e:
            logging.error("Error getting pytube results", exc_info=e)

        return []

    def play(
        self,
        track: YoutubeTrack,
        voice: VoiceClient,
        after: Callable | None = None,
    ):
        audio = self.load_track(track)
        voice.play(audio, after=after)
