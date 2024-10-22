# What the fuck?

Say you want to get VLC hooked up to discord RPC (Rich Presence) absolutely refuse to install Discord desktop, and would rather not have a suspicious tampermonkey script running in browser.

A very niche set of requirement so don't worry about it if it doesn't suite your needs.

## Dependencies

- https://github.com/acheong08/vlc-discord-rpc (if you use `yt-dlp`. Helps with thumbnails for songs not in an album. Otherwise just use original)
- https://github.com/OpenAsar/arrpc/


Run `arrpc` first, then `vlc-discord-rpc`.

## Build and run

You need to set the `DISCORD_TOKEN` environment variable. [Here's how to get that token](https://stackoverflow.com/questions/67348339/any-way-to-get-my-discord-token-from-browser-dev-console)

`go run main.go`

Now when you play something in VLC, you get the nice status with thumbnail.

![image](https://github.com/user-attachments/assets/4c386a75-af86-4027-a1b3-2aa9e6138d72)

