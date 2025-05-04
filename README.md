# qaBot

qaBot is a Telegram bot built using Go (Golang) that provides a question and answer service. The bot allows users to ask predefined questions and receive corresponding answers. It also supports subquestions, enabling a more interactive experience.

## Features

- List of predefined questions and answers.
- Support for subquestions.
- Easy configuration using YAML files.
- SQLite database for storing questions and answers.

## Project Structure

```
qaBot
├── cmd
│   └── main.go               # Entry point of the application
├── config
│   └── config_local.yml      # Configuration settings for the bot
├── internal
│   ├── infrastructure
│   │   └── database
│   │       └── sqlite.go     # SQLite database connection and operations
│   ├── bot
│       ├── handler.go        # Logic for handling incoming messages
│       └── routes.go         # Routing for bot commands and messages
├── pkg
│   └── config
│       ├── config.go         # Configuration structure and access methods
│       └── loader.go         # Logic for loading configuration files
├── go.mod                     # Module definition and dependencies
├── go.sum                     # Checksums for module dependencies
└── README.md                  # Project documentation
```

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/qaBot.git
   cd qaBot
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Configure the bot by editing the `config/config_local.yml` file with your bot token and database connection string.

## Usage

To run the bot, use the following command, specifying the configuration file if needed:
```
go run cmd/main.go -config=local
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any suggestions or improvements.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.