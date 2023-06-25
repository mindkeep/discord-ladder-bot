# discord_ladder_bot
Discord bot to track and manage a ladder tournament

## Implemented features

- connection to discord and listening/responding to commands
- writing and reading ranking data to MongoDB
- working commands
    - help, provides a list of commands
    - init, initializes ranking data for the channel
    - delete_tournament, resets the tournament and drops data
    - register, adds the requesting player to the ladder
    - unregister, removes the requesting player
    - maybe more, but need some volunteers to test things
- some unit testing for ranking data

## Planned features

- Bot is able to accept commands
- implement more commands
    - active challenges, Display the current active challenges.
    - challenge
    - cancel challenge
    - report results
    - settings
    - print tournament standings
    - result history
- add user controls for admin users


## Other ideas

- Hook into ChatGPT/OpenAI to interpret user messages into commands
