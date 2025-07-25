# Go get url! 🏃‍♂️‍➡

GoGetURL is a web application built with Go and Gin for analyzing HTML web pages. The app takes a URL input from the user, fetches the page content, and performs a detailed analysis including HTML version detection, title extraction, link categorization, header counting, and login form detection.

![Demo](https://github.com/user-attachments/assets/1c726847-2fe3-4c22-bdb2-b7dd5cac037b)

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
- Dockerized for easy deployment with multi-stage builds

## Technical Overview & Engineering Practices

This project uses standard Go practices to keep the code clean, easy to follow, and practical for real-world use.
### Stack & Tools

- **Go (Golang)**: The core programming language used.
- **Gin**: Lightweight HTTP web framework.
- **HTML Tokenizer**: `golang.org/x/net/html` for parsing HTML content.
- **slog**: Used for structured logging without manually formatting log output.
- **air**: Provides hot-reloading during development.
- **stretchr/testify**: Enables expressive and readable test assertions.
- **Standard Testing Tools**: Leveraged for unit and integration testing.
- **Go Modules**: For dependency management.

### Engineering Highlights

- **URL Validation**: Handled using Go’s standard `net/url` and `regexp` packages.
- **Logging**: Structured and leveled logging with `slog`, adhering to modern practices.
- **Concurrency**: Applied appropriately with goroutines and `sync.WaitGroup`—especially for broken link checking.
- **Error Handling**: Proper HTTP status codes and user-friendly messages are returned for all failure cases.
- **No JavaScript**: The UI avoids JS for simplicity and backend focus, using Go templates for rendering.
- **Testing Strategy**: Over 70% coverage with both unit and integration tests, testing user input flows and analyzer logic.
- **CI Workflow**: Includes a basic GitHub Actions workflow that runs build and test steps using Makefile commands for consistency.

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

   Or, build the app first and run the binary:
   ```bash
   go build -o gogeturl ./cmd/gogeturl
   ./gogeturl
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

### Run with Docker

You can run the app with docker by using `make` commands:

```bash
make docker-build
make docker-run
```
This uses the `.env` file to set the port and builds the Go application with a minimal runtime image.

### Run with Makefile

This project includes a `Makefile` to simplify development and deployment workflows. Just run `make` in the terminal to see the available commands:

```bash
  make build        Build the Go binary
  make run          Run the application locally
  make test         Run tests
  make clean        Remove the built binary
  make docker-build Build Docker image
  make docker-run   Run Docker container
  make lint         Run linters
```

You can explore the `Makefile` for more supported targets.

These Makefile targets are also utilized in the GitHub Actions CI workflow.

For more details, see the [INSTRUCTIONS.md](./INSTRUCTIONS.md) file.
