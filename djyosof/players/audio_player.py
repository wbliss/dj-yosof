from asyncio import Event, Queue

from discord import VoiceClient, Interaction

from djyosof.bot import DJYosof
from djyosof.audio_types.playable_audio import PlayableAudio
from djyosof.cogs import utilities


class AudioPlayer:
    def __init__(
        self,
        bot: DJYosof,
    ):
        self.queue = Queue()
        self.next: Event = Event()
        self.bot: DJYosof = bot
        self.is_playing: bool = False
        self.guild_id = -1

    async def enqueue_and_play(
        self,
        track: PlayableAudio,
        voice: VoiceClient,
        interaction: Interaction,
    ):
        await self.enqueue(track, interaction)
        if not self.is_playing:
            self.bot.loop.create_task(self.play_loop(voice, interaction))

    async def enqueue(self, track: PlayableAudio, interaction: Interaction):
        await self.queue.put(track)
        await interaction.response.send_message(
            f"Added {track.name} by {track.artist} to the queue"
        )

    async def play_loop(self, voice: VoiceClient, interaction: Interaction):
        # Grab latest track off the queue and play it
        await self.bot.wait_until_ready()
        self.guild_id = interaction.guild_id

        self.is_playing = True
        while not self.bot.is_closed():
            self.next.clear()

            if self.queue.empty():
                await utilities.leave(interaction)
                break

            track = await self.queue.get()
            player = self.bot.players[track.get_type()]

            player.play(
                track,
                voice,
                after=lambda _: self.bot.loop.call_soon_threadsafe(self.next.set),
            )

            print(f"Playing {track.name} by {track.artist}")
            await interaction.followup.send(
                content="Playing music!", embed=track.get_embed()
            )
            await self.next.wait()

        self.is_playing = False
