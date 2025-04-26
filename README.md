# Oura Line Bot

This is a Go-based application that integrates with the Oura API and LINE Messaging API to fetch daily activity, sleep, and sleep score data from the Oura Ring and send it as a message to a LINE user.

## Features

- Fetches daily activity data (e.g., active score, calories burned, non-wear time) from the Oura API.
- Fetches sleep data (e.g., total sleep duration, deep sleep, REM sleep) from the Oura API.
- Fetches sleep score data (e.g., sleep efficiency, restfulness, deep sleep score) from the Oura API.
- Sends the fetched data as a formatted message to a LINE user.
- Alerts the user if the Oura Ring's battery is low or if no data is available.

## Prerequisites

- Go 1.18 or later installed on your system.
- An Oura API token. You can generate one from the [Oura Developer Portal](https://cloud.ouraring.com/).
- A LINE Messaging API token and user ID. You can set up a LINE bot and obtain these from the [LINE Developers Console](https://developers.line.biz/).

## Installation

1. Clone this repository:
	```bash
	git clone <repository-url>
	cd <repository-folder>
	```

2. Install dependencies:
	```bash
	go mod tidy
	```

3. Create a `.env` file in the root directory and add the following environment variables:
	```
	OURA_API_TOKEN=<your-oura-api-token>
	LINE_API_TOKEN=<your-line-api-token>
	TO_LINE_USER=<line-user-id>
	```

## Usage

1. Run the application:
	```bash
	go run main.go
	```

2. Build the application:
    ```bash
    go build -o app main.go
    ```
### Supported Build Targets

The application can be built for various operating systems and architectures. Below is a table of supported build targets:

| `GOOS`   | `GOARCH` | Description               |
|----------|----------|---------------------------|
| linux    | amd64    | 64-bit Linux              |
| linux    | arm64    | 64-bit ARM Linux          |
| darwin   | amd64    | macOS x86_64              |
| darwin   | arm64    | macOS M1 (Apple Silicon)  |
| windows  | amd64    | 64-bit Windows            |
| windows  | 386      | 32-bit Windows            |

You can specify the target operating system and architecture using the `GOOS` and `GOARCH` environment variables when building the application. For example:

```bash
# Build for 64-bit Linux
GOOS=linux GOARCH=amd64 go build -o app main.go

# Build for macOS M1 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o app main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o app.exe main.go
```

3. Schedule the application to run daily using a cron job or any task scheduler of your choice.

## Environment Variables

- `OURA_API_TOKEN`: Your Oura API token.
- `LINE_API_TOKEN`: Your LINE Messaging API token.
- `TO_LINE_USER`: The LINE user ID to whom the message will be sent.

## File Structure

- `main.go`: The main application file containing all logic for fetching data, formatting messages, and sending them to LINE.
- `.env`: Environment variables file (not included in the repository; you need to create it).

## Functions

### Data Fetching
- `fetchOuraData(endpoint, token, startdate, enddate string)`: Fetches data from the Oura API.

### Message Formatting
- `formatDate(date time.Time)`: Formats a date to `YYYY-MM-DD`.
- `formatSecondsToTime(seconds int)`: Converts seconds into `hh:mm:ss` format.
- `getBatteryStatusMessage(batterystatus bool)`: Returns a message based on the battery status.

### LINE Messaging
- `sendLineMessage(token, userId, message string)`: Sends a message to a LINE user.

## Error Handling

- If any required environment variable is missing, the application will terminate with an error.
- If the Oura API or LINE API returns an error, the application will log the error and terminate.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
