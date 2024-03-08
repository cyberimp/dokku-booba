# Booba watcher

[![Dynamic JSON Badge](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fhentai.tiddies.pics%2Fstats.json&query=%24.chats&logo=telegram&label=users&color=%2326A5E4&cacheSeconds=3600&link=https%3A%2F%2Ft.me%2Ftiddies2dbot)](https://hentai.tiddies.pics)
[![Go Report Card](https://goreportcard.com/badge/github.com/cyberimp/dokku-booba)](https://goreportcard.com/report/github.com/cyberimp/dokku-booba)

Automatic moderator for adult Telegram channel, now with dokku deployment

## Setup instructions

1. Install [dokku](https://dokku.com/) on your VPS
2. Install [dokku postgres plugin](https://github.com/dokku/dokku-postgres)
3. Create table "antibayan" with `antibayan.sql`
4. Push this repo for your app
5. Add SSL cert with [dokku-letsencrypt](https://github.com/dokku/dokku-letsencrypt)
6. Add webhook using [this guide](https://stackoverflow.com/questions/42554548/how-to-set-telegram-bot-webhook)
(endpoint should be https://your-hosting/your-token)
7. Set some vars:
    - `TG_CHAT` - numeric ID of chat
    - `DANBOORU_API_KEY` - your gold account danbooru API key
    - `DANBOORU_LOGIN` - your danbooru login
    - `DOKKU_LETSENCRYPT_EMAIL` - your let's encrypt email
    - `TG_TOKEN` - your bot token
8. You should restart app with `dokku ps:restart <YOUR_APP>` every day to rebuild cache (don't worry, bot will be up during restart, new container will replace old after building cache)
9. You could post into `TG_CHAT` channel with `docker exec <YOUR_APP>.web.1 /app/sigusr1.sh`
10. First start will take about 2 mins
