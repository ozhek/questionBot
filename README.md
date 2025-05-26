# qaBot

qaBot is a Telegram bot built using Go (Golang) that provides a question and answer service. It is designed for high-concurrency environments and uses PostgreSQL as a backend. Users can explore predefined questions and subquestions interactively.

## Features

- List of predefined questions and answers.
- Support for hierarchical subquestions.
- Easy configuration using YAML files.
- PostgreSQL database for storing questions and answers.
- Support for multiple concurrent users.

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
│   │       └── pg_repository.go # PostgreSQL database connection and operations
│   ├── bot
│       ├── handler.go        # Logic for handling incoming messages
│       └── routes.go         # Routing for bot commands and messages
├── pkg
│   └── config
│       └── config.go         # Configuration structure and access methods
├── go.mod                    # Module definition and dependencies
├── go.sum                    # Checksums for module dependencies
└── README.md                 # Project documentation
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

3. Set up PostgreSQL:
   - Ensure PostgreSQL is running locally or in Docker.
   - Create a database and user.
   - Run the provided migration script if applicable.

4. Configure the bot:
   - Edit the `config/config_local.yml` file with your bot token and PostgreSQL connection settings (host, port, user, password, dbname, sslmode, max connections).

## Usage

To run the bot with the local configuration file:
```
go run cmd/main.go -config=local
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any suggestions or improvements.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.