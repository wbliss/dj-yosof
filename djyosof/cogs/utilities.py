import logging

from discord import ApplicationContext, VoiceClient


async def connect_or_move(
    ctx: ApplicationContext, *args, **kwargs
) -> VoiceClient | None:
    author_voice = ctx.user.voice
    # yeah this won't work
    if not author_voice:
        logging.info("User not in voice channel")
        await ctx.respond("Join a voice channel first.")
        return None

    author_voice_channel = author_voice.channel

    # Not connected anywhere, connect
    current_voice_client = ctx.guild.voice_client
    if not current_voice_client:
        logging.info(f"Joining: {author_voice_channel}")
        return await author_voice_channel.connect(*args, **kwargs)

    # If we're already in a channel for that guild check to see
    # if we need to move channels or do nothing
    current_voice_channel = current_voice_client.channel
    if author_voice_channel == current_voice_channel:
        logging.info(f"Already in {author_voice_channel}, not joining")
        return current_voice_client

    logging.info(f"Joining: {author_voice_channel}")
    return await current_voice_client.move_to(author_voice_channel)


async def leave(ctx: ApplicationContext) -> None:
    current_voice_client = ctx.guild.voice_client

    if current_voice_client:
        await current_voice_client.disconnect(force=True)
