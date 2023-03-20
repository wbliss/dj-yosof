import logging
import traceback
from urllib.parse import urlparse, parse_qs
from collections.abc import Callable
from io import BytesIO
from pytube import Search, YouTube, Playlist

import discord
from discord import VoiceClient

from djyosof.audio_types.youtube import YoutubeTrack


class YoutubeSource:
    def load_track(self, track: YoutubeTrack):
        FFMPEG_OPTS = {
            "before_options": "-reconnect 1 -reconnect_streamed 1 -reconnect_delay_max 5",
            "options": "-vn",
        }

        yt = YouTube(track.watch_url)
        stream = yt.streams.get_audio_only()
        return discord.FFmpegPCMAudio(stream.url, **FFMPEG_OPTS)

    # TODO: burn this function to the ground
    def parse_playlist(self, playlist: Playlist):
        tracks = []

        videos = playlist.initial_data["contents"]["twoColumnBrowseResultsRenderer"][
            "tabs"
        ][0]["tabRenderer"]["content"]["sectionListRenderer"]["contents"][0][
            "itemSectionRenderer"
        ][
            "contents"
        ][
            0
        ][
            "playlistVideoListRenderer"
        ][
            "contents"
        ]
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
            except KeyError as e:
                logging.info(traceback.format_exc())
                return []
        elif parsed_url.path == "/watch":
            return [YoutubeTrack.from_pytube(YouTube(link))]
        else:
            # unrecognized link
            return []

    def search(self, query: str):
        return [
            YoutubeTrack.from_pytube(result) for result in Search(query).results[:5]
        ]

    def play(
        self,
        track: YoutubeTrack,
        voice: VoiceClient,
        after: Callable | None = None,
    ):
        audio = self.load_track(track)
        voice.play(audio, after=after)
