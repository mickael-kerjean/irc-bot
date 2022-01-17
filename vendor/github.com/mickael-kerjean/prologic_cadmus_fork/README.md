This is a fork of https://github.com/prologic/cadmus/

Build:
```
docker build . -t machines/irc_bot
```
Run:
```
docker run --name irc_bot -v /tmp/irclog:/app/logs -e BOT_CHANNELS="#freenode" -e BOT_NICK=nickname -e BOT_USER=username -e BOT_REALNAME=RealName -e BOT_PASSWORD=password machines/irc_bot
```