"""Contains class that controls playing audio"""

import logging
import traceback
from asyncio import Event, Queue, sleep
from typing import TYPE_CHECKING

from discord import ApplicationContext, VoiceClient, Message

from djyosof.audio_types.playable_audio import PlayableAudio
from djyosof.cogs import utilities

if TYPE_CHECKING:
    from ..bot import DJYosof


class AudioPlayer:
    """
    Handles the queue and play loop of an audio player.
    """

    def __init__(
        self,
        bot: "DJYosof",
    ):
        self.queue: Queue = Queue()
        self.next: Event = Event()
        self.bot: "DJYosof" = bot
        self.is_playing: bool = False
        self.now_playing: PlayableAudio | None = None

    async def enqueue_and_play(
        self,
        track: PlayableAudio,
        voice: VoiceClient,
        ctx: ApplicationContext,
    ):
        """Queues a track and begins the play loop it not currently running."""
        await self.enqueue(track, ctx)
        if not self.is_playing:
            self.bot.loop.create_task(self.play_loop(voice, ctx))

    async def enqueue(self, track: PlayableAudio, ctx: ApplicationContext):
        """Adds a track to the end of a queue."""
        await self.queue.put(track)

    async def play_loop(self, voice: VoiceClient, ctx: ApplicationContext):
        """Loop to play through any songs in the queue."""
        # Grab latest track off the queue and play it
        await self.bot.wait_until_ready()
        channel = self.bot.get_channel(ctx.channel_id)

        self.is_playing = True
        now_playing_message: Message = None
        while not self.bot.is_closed():
            self.next.clear()

            if self.queue.empty():
                await sleep(10)
                if self.queue.empty():
                    if now_playing_message:
                        await now_playing_message.delete()
                    await utilities.leave(ctx)
                    break

            self.now_playing = await self.queue.get()
            player = self.bot.players[self.now_playing.get_type()]

            try:
                player.play(
                    self.now_playing,
                    voice,
                    after=lambda _: self.bot.loop.call_soon_threadsafe(self.next.set),
                )
                logging.info("Playing %s", self.now_playing.get_display_name())

                if now_playing_message:
                    await now_playing_message.delete()

                embed = self.now_playing.get_embed()
                queue_markdown = self._get_queue_markdown(ctx)
                embed.add_field(name="Queue", value=queue_markdown)

                now_playing_message = await channel.send(content="", embed=embed)
            except Exception:
                logging.info(
                    f"Failed to play {self.now_playing.get_display_name()}, skipping"
                )
                traceback.print_exc()
                await channel.send(
                    content=f"Failed to play {self.now_playing.get_display_name()}, skipping"
                )
                self.bot.loop.call_soon_threadsafe(self.next.set)

            await self.next.wait()

        self.is_playing = False

    def skip(
        self,
        voice: VoiceClient,
    ):
        # Stops the current song so next will be picked up
        voice.stop()

    def stop(
        self,
        voice: VoiceClient,
    ):
        # Clear queue and stop playing
        for _ in range(self.queue.qsize()):
            self.queue.get_nowait()
            self.queue.task_done()

        voice.stop()

    def _get_queue_markdown(self, ctx: ApplicationContext):
        queue_markdown = ""
        for idx, track in enumerate(
            ([self.now_playing] + list(self.queue._queue))[:10]
        ):
            queue_markdown += f"**{idx + 1}**. {track.get_display_name()}"
            if idx == 0:
                queue_markdown += " - NOW PLAYING"
            queue_markdown += "\n"

        if queue_markdown == "":
            queue_markdown = "Queue is empty!"
        else:
            queue_length = (
                len(list(self.bot.audio_players[ctx.guild_id].queue._queue)) + 1
            )
            queue_markdown += f"\nShowing {min(10, queue_length)} out of {queue_length} tracks in the queue."
        return queue_markdown
