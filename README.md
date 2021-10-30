# DuelBot
A Telegram game-bot with real time fights between actual players.
Created in [Go](https://golang.org/) using [echotron](https://github.com/NicoNex/echotron)

> **Notes:**
> This bot is currently on beta and might be pretty unstable 


## Beta available
Play now at [@DuellingRobot](https://t.me/DuellingRobot)


## How to run
To run your own bot you will need to have Go installed.
After you clone this repository you need to build it using: `go build`
Now you should have a executable file called _"DuelBot"_ (Linux) or _"DuelBot.exe"_ (Windows).
Create a new bot using [@BotFather](https://t.me/BotFather) and grab the API Token.
You can now run the bot by typing
`<executable> <token>`
Or save the token in a txt file and type on the shell:
`<executable> --readfrom <filepath>`

> "<executable>" is the name of the file that you obtain after you build the repo. Typically "DuelBot" or "DuelBot.exe"
> "<token>" is the API Bot Token that BotFather generated for your bot
> "<filepath>" is the path where you saved the txt file containing the token
