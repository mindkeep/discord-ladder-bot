# discord_ladder_bot

Discord bot to track and manage a ladder tournament

## Implemented features

- connection to discord and listening/responding to commands
- writing and reading ranking data to MongoDB
- commands
  - cancel
  - challenge
  - delete_tournament
  - forfeit
  - help
  - history
  - init
  - ladder
  - register
  - result
  - set
  - standings
  - unregister
- some unit testing for ranking data
- admin id list to allow/disallow certain commands

## TODO

- periodic checks for match/challenge timeout
- periodic checks for players that have left the server
- move print functions into the discordbot handlers
- cancel challenge should not be in the history
- fix mongoDB writes to be session based and per channel
- add more unit tests (the never ending TODO)
- fix bugs

## Other ideas

- Hook into some LLM to interpret user messages into commands or other chat banter
- Add a web interface to show standings and other data
