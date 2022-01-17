This is an IRC bot that makes it possible to log your IRC chanel. This is what powers the support website of Filestash: https://support.filestash.app/

# Build

Build:
```
docker build --no-cache --pull . -t machines/irc_bot
docker push machines/irc_bot
```
Run:
```
docker run --name irc_bot -v /tmp/irclog:/app/logs -e BOT_CHANNELS="#freenode" -e BOT_NICK=nickname -e BOT_USER=username -e BOT_REALNAME=RealName -e BOT_PASSWORD=password machines/irc_bot
```

# Credit

- @cadmus is the original author
- authors of library without whom I probably would have use discord, slack or whatever chat message of the day
