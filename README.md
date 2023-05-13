# twitch-kraken by devusSs

## Disclaimer

DO NOT (!) USE THIS PROGRAM WITHOUT KNOWING WHAT IT DOES OR GETTING LECTURED BY THE CREATOR (devusSs).

This program is not affiliated with any services mentioned inside or outside of it.

This program may also not be safe for production use. It should be seen as a fun / site project by the developer.

Please contact the owner (devusSs) at devuscs@gmail.com to resolve any issues (especially trademarks and copyright stuff).

## What does this program do?

This program provides a bot to connect to the [Twitch](https://twitch.tv) IRC (basically chat and events) server to collect information about your stream. This includes and is not limited to: users connecting and disconnecting from your chat / stream, messages the users sent, certain Twitch events like follows, subs, etc. or Twitch global state (not fully supported yet) events. The bot can also add, edit, delete and execute commands (more docs on that later).

The program is still in an early stage so please do not expect everything to work properly or to be fully implemented yet. It is a `work in progress`.

## Setup

Make sure you have a working [Postgres](https://www.postgresql.org/) instance running. Please read their documentation on how to achieve that. For cleaner infrastructure it may also be useful to use [Docker](https://www.docker.com/) for that purpose. Docker support for the entire project will be added later.

You will also need a [Twitch](https://twitch.tv) account for your bot to use or you may use your own broadcaster acccount. Since Twitch does not allow logging in to the IRC server using your plain password, you will need to generate an oauth password. This can be done [here](https://twitchapps.com/tmi/). Please make sure you login with the account your bot is supposed to use.

To use certain built-in [Twitch](https://twitch.tv) as well as [Spotify](https://spotify.com) features you will need a developer account on each platform. You can do that on [Twitch's dev platform](https://dev.twitch.tv) and [Spotify's dev platform](https://developers.spotify.com) respectively.

The bot needs a valid config file to work. You may check the [example config]("./files/config.json") for more information. The config can be placed anywhere you'd like it to but the `-c` flag inside the program is configured to use `./files/config.json` as default input, so using that path is highly recommended.

## Building and running the app

After setting up the config file properly you can either build the app manually (see below) or download the [latest release](https://github.com/devusSs/twitch-kraken/releases/latest).

To build the app manually you will need to clone the repository, make sure you have [Make](https://www.gnu.org/software/make/) installed (pre-installed on most UNIX systems) and make sure you have [Go/Golang](https://go.dev) installed and use the latest version.

You may then head to the cloned repository's directory and run `make build` to build the project which will create a `release/` folder with versions for Windows (AMD64), Linux (AMD64) and MacOS (ARM64). Choose your corresponding version and make sure the config file is set via the `-c` flag when running the program.

If you would like to run the app instantly after building it you may also use the `make dev` command in the repository's directory. This is not recommended however since this uses the `-c` flag with `./files/config.dev.json` as default value and will need configuration.

## Debugging the app

It is possible to simply run the app to get certain build information or check why some features may not be working as expected. Currently the `-d` flag (for diagnosis mode) and the `-v` flag (for build information) are supported. You may also use the command `make diag` in the repository's directory to automatically run the diagnosis mode, the app will rebuild itself before.

## Further features (soonTM)

- more built in Twitch commands like settitle, setgame, getfollowers, getsubs
- support for Spotify to show current song via !song command
- support for Faceit to show current elo, level, potentially match
- built in giveaway system based on information collected from users
- api for public command listing support, private information for bot owner
