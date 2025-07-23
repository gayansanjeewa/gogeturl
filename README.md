# Go get url! 🏃‍♂️‍➡

GoGetURL is a web application built with Go and Gin for analyzing HTML web pages. The app takes a URL input from the user, fetches the page content, and performs a detailed analysis including HTML version detection, title extraction, link categorization, header counting, and login form detection.

## Features

- Detects the HTML version from the document doctype
- Extracts the page title
- Counts all headings (h1-h6) with a detailed breakdown
- Identifies and categorizes internal, external, and broken links
- Detects the presence of login forms based on input fields
- Provides clear error messages if the URL is unreachable or invalid
- Includes unit and integration tests
- Leaner Git commit history with reference to the related PR 
- Hot-reloading with Air for development
- TODO: Dockerized for easy deployment

## Technologies Used

- Go (Golang)
- Gin web framework
- HTML tokenizer (`golang.org/x/net/html`)
- `slog` for structured logging
- `air` for live reload
- `stretchr/testify` for test assertions
- Standard Go testing


## Instructions

### Setup Instructions

1. **Clone the repository**
   ```bash
   git clone https://github.com/gayansanjeewa/gogeturl.git
   cd gogeturl
   ```

2. **Install dependencies**
   Ensure you have Go installed (1.20 or later), then run:
   ```bash
   go mod tidy
   ```

3. **Configure environment**
   Copy the provided `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```
   You can modify the `PORT` variable inside `.env` as needed.

4. **Run the application**
   You can start the server using:
   ```bash
   go run ./cmd/gogeturl
   ```
   Or, for hot reload during development, install [Air](https://github.com/air-verse/air) and run:
   ```bash
   air
   ```

5. **Access the application**
   Open your browser and go to (if the port is 8080):
   ```
   http://localhost:8080
   ```

For more details, see the [INSTRUCTIONS.md](./INSTRUCTIONS.md) file.
