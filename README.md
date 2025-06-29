# DASA BOT

Welcome to the official repo for the Dasa Bot for the [Dasa Server](https://discord.gg/VJCYUjf6bu)!

---

## What is the Dasa Bot?

The Dasa Bot is a Discord Bot aimed at helping DASA members with their DASA ranks and solving queries regarding to placements. It adds helpful commands to be able to

- View Cutoffs for Colleges and Branches
- Analyze their Ranks and be able to figure out which college they can get into

---

## How Do I Use It?

Our **Discord server** has the bot and its DB hosted 24/7, feel free to join and check it out

Join us here: [Discord Invite](https://discord.gg/VJCYUjf6bu)

---

## What is This Repository?

This repository contains the entire infrastructure for the DASA BOT, including:

1. **Bot Code**
2. **DB Files and Structure**
3. **Github Workflows for Continuous Deployment**
4. **Docker Compose File for Orchestration of Bot and DB**

## How to Setup the Bot?

To setup the bot, you need to have a working Go environment and Docker installed. Once done, you need to make a `.env` file in the `bot` directory. Once done, run the command `docker compose up pocketbase -d` which will start the pocketbase db, where you can insert the `/db/migrations.json` file for setting it up.

## ENV Structure

The `.env` file contains the following variables:

| Variable       | Description               |
| -------------- | ------------------------- |
| TOKEN          | Discord Bot Token         |
| GUILD_ID       | Discord Guild ID          |
| ADMIN_EMAIL    | Pocketbase Admin Email    |
| ADMIN_PASSWORD | Pocketbase Admin Password |
| BASE_DOMAIN    | Pocketbase Base Domain    |
| MOD_ROLE       | Discord Role ID           |
| ADMIN_CHANNEL  | Discord Channel ID        |

All of the values are required to run the bot.

---

### Built by Arinji

This project was proudly built by [Arinji](https://www.arinji.com/). Check out my website for more cool projects!

---
