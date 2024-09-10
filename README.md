## How to Use the Bot

### 1. Clone the Repository

Make sure you have cloned the repository to your local machine.

### 2. Edit the Configuration File

1. **Locate the `env-examples` File**

   - Find the file named `env-examples` in the repository. This file contains example environment variables required for the bot.

2. **Rename and Edit the File**

   - Rename this file to `.env` or another configuration file name used by your bot.
   - Open the renamed file in a text editor.
   - Look for the line starting with `DISCORD_TOKEN`.
   - Replace `your-bot-token-here` with your actual Discord bot token. For example:
     
     ```env
     DISCORD_TOKEN=your-actual-bot-token
     ```

### 3. Obtain Your Discord Bot Token

1. **Create a Bot on Discord**

   - Go to the [Discord Developer Portal](https://discord.com/developers/applications).
   - Log in with your Discord account if you aren't already logged in.

2. **Create a New Application**

   - Click the "New Application" button.
   - Give your application a name and click "Create."

3. **Add a Bot to Your Application**

   - Select your newly created application from the list.
   - Go to the "Bot" tab on the left side.
   - Click "Add Bot" and confirm by clicking "Yes, do it!"

4. **Get Your Bot Token**

   - Under the "Bot" tab, you will see a section titled "TOKEN."
   - Click the "Copy" button to copy your bot token to your clipboard.
   - Paste this token into the `DISCORD_TOKEN` field in your `.env` file.

### 4. Initialize the Bot

1. **Open Terminal or Command Prompt**

2. **Navigate to the Bot Directory**

   ```bash
   cd bot_discord_go
   go run .
    ```
3. **You should see a message like** "Bot está corriendo. Presiona CTRL+C para salir." **in your console**



