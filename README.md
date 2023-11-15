# Access Governance System

### How to run
1. Install [Docker](https://docs.docker.com/get-docker)
2. Install [Task](https://taskfile.dev/installation)
3. Create a [bot](https://t.me/BotFather) in Telegram and get a token
4. Setup environment variables
   1. Copy `.env.example` to `.env`
   2. Edit `.env` file
5. Run `task up`

| Environment Variable                       | Description                                                                                                   | Required |
|--------------------------------------------|---------------------------------------------------------------------------------------------------------------|----------|
| `ENVIRONMENT`                              | Specifies the environment in which your application is running (development, production).            | Yes   |
| `COMMUNITY_NAME`                           | The name of the community that this instance of the application is serving.                                   | Yes   |
| `LOKI_URL`                                 | The URL of a Loki instance for logging.                                                                       | No   |
| `VOTING_DURATION_DAYS`                     | The duration of voting events in your application, in days.                                                   | Yes   |
| `INITIAL_SEEDERS`                          | A comma-separated list of telegram usernames who are initial seeders, likely having special roles or permissions.                 | No   |
| `RENOMINATION_PERIOD_DAYS`                 | Defines the period in days during which the same individual cannot be re-nominated for voting.                                            | Yes   |
| `TELEGRAM_MEMBERS_CHAT_ID`                          | The ID of a chat in Telegram where members are present.                                                       | Yes   |
| `DB_PASSWORD`                              | The password used to connect to your PostgreSQL database.                                                     | Yes   |
| `DB_URL`                                   | The connection string for your PostgreSQL database.                                                           | Yes   |
| `TELEGRAM_ACCESS_GOVERNANCE_BOT_TOKEN`     | The API token used to authenticate with the Telegram API for a bot that manages access to community.         | Yes   |
| `TELEGRAM_VOTE_BOT_TOKEN`                  | The API token used to authenticate with the Telegram API for a bot that manages voting.                       | Yes   |
| `TELEGRAM_AUTHORIZATION_BOT_TOKEN`         | The API token used to authenticate with the Telegram API for a bot that manages authorization to a community.                | Yes   |
| `DISCORD_AUTHORIZATION_BOT_TOKEN`          | The API token used to authenticate with the Discord API for a bot that manages authorization to a community.                | Yes   |
| `DISCORD_SERVER_ID`                        | The ID of your server on Discord.                                                                             | Yes   |
| `DISCORD_GUEST_ROLE_ID`                    | The ID of the guest role on your Discord server.                                                              | Yes   |
| `DISCORD_MEMBER_ROLE_ID`                   | The ID of the member role on your Discord server.                                                             | Yes   |
| `VOTE_API_URL`                             | The URL of an API related to voting functionality.                                                            | Yes   |
| `QUORUM`                                   | The minimum proportion of members who must participate in a vote for it to be valid.                          | Yes   |
| `MIN_YES_PERCENTAGE`                       | The minimum proportion of "yes" votes required for a vote to pass.                                            | Yes   |
| `YES_VOTES_TO_OVERCOME_NO`                 | The proportion of "yes" votes required to overcome any "no" votes and pass a vote.                            | Yes   |

### How to stop
Run `task down`
