# discord_ladder_bot
Discord bot to track and manage a ladder tournament

## Implemented features

- working commands
    - bot is able to connect to discord and accept a handful of commands
    - we'll define "working" as able to take action on the database

## Planned features

- Add some testing before it gets too late...
- Bot is able to accept commands
    - register
    - print
        - full listing
        - slim results (+/- 2 from active player)
        - skip locked players
    - challenge
    - result
        - maybe require both players to report result?
    - forefeit
    - set optional info
        - timezone
        - preferred playtimes
        - preferred server
    - info
    - history
- Hook into MongoDB Atlas as a data source
- Create tunable settings

## Other ideas

- Hook into ChatGPT/OpenAI to interpret user messages into commands
