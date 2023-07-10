# discord_ladder_bot
Discord bot to track and manage a ladder tournament

## Implemented features

- connection to discord and listening/responding to commands
- writing and reading ranking data to MongoDB
- commands
    - help
    - init
    - delete_tournament
    - register
    - unregister
    - challenge
    - cancel
    - forfeit
    - result
    - history
    - ladder
    - set
- some unit testing for ranking data
- admin id list to allow/disallow certain commands
- implemented /ladder command!

## TODO

- fix bugs
- move print functions into the discordbot handlers
- add more unit tests (the neverending TODO)
- cancel challenge should not be in the history

## Planned features

- periodic checks for match timeout
- periodic checks for players that have left the server

## Other ideas

- Hook into ChatGPT/OpenAI to interpret user messages into commands
