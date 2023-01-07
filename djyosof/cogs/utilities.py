from queue import Queue
from discord.ext import commands
from discord import Interaction, VoiceClient

from djyosof.audio_types.playable_audio import AudioType, PlayableAudio


async def connect_or_move(
    interaction: Interaction, *args, **kwargs
) -> VoiceClient | None:
    author_voice = interaction.user.voice
    # yeah this won't work
    if not author_voice:
        await interaction.response.send_message("Join a voice channel first.")
        return None

    author_voice_channel = author_voice.channel

    # Not connected anywhere, connect
    current_voice_client = interaction.guild.voice_client
    if not current_voice_client:
        print(f"Joining: {author_voice_channel}")
        return await author_voice_channel.connect(*args, **kwargs)

    # If we're already in a channel for that guild check to see
    # if we need to move channels or do nothing
    current_voice_channel = current_voice_client.channel
    if author_voice_channel == current_voice_channel:
        print(f"Already in {author_voice_channel}, not joining")
        return current_voice_client

    print(f"Joining: {author_voice_channel}")
    return await current_voice_client.move_to(author_voice_channel)


async def leave(interaction: Interaction) -> None:
    current_voice_client = interaction.guild.voice_client
    return await current_voice_client.disconnect()


async def queue_or_play(
    bot: commands.Bot,
    track: PlayableAudio,
    voice: VoiceClient,
    interaction: Interaction,
):
    if bot.queues[interaction.guild_id].empty() and not voice.is_playing():
        await bot.players[AudioType.spotify].play(track, voice, interaction)
    else:
        bot.queues[interaction.guild_id].put(track)
        await interaction.response.send_message(
            f"Added {track.name} by {track.artist} to the queue"
        )


async def play_track(
    bot: commands.Bot,
    guild_id: int,
    interaction: Interaction,
    exception: Exception | None,
):
    track = bot.queues[guild_id].get()
    player = bot.players[track.get_type()]
    # interaction.response.send_message()
    player.play(
        track,
        voice,
        interaction,
        after=lambda e: play_track(bot, guild_id, interaction, e),
    )
