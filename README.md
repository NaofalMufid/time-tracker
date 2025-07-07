# Task Time Tracker

Task Time Tracker is a lightweight and efficient Go-based web application designed to help users manage and track their tasks effortlessly. It offers a simple yet powerful interface for task creation, tracking, pausing, resuming, stopping, and deletion. This application is built as a self-contained executable for ease of use on various desktop platforms.

## Table of Contents
- [Features](#features)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Development Setup](#development-setup)
- [Usage](#usage)
  - [Running the Executable](#running-the-executable)
  - [Data Persistence](#data-persistence)
  - [Stopping the Application Gracefully](#stopping-the-application-gracefully)
- [Project Structure](#project-structure)
- [API Endpoints](#api-endpoints)
- [Contributing](#contributing)
- [License](#license)

## Features
- **Task Management**: Create, pause, resume, stop, and delete tasks.
- **Real-time Active Task Tracking**: Displays the currently running task with live duration updates.
- **Pagination & Filtering**: Efficiently browse and manage tasks with pagination, status filtering (all, paused, stopped), and sorting options (start time, title).
- **Self-Contained Executable**: Frontend assets (HTML, CSS, JavaScript) are embedded directly into the Go binary for easy distribution and offline use.
- **Persistent Data Storage**: Task data is stored locally using SQLite in a user-specific directory, ensuring data is retained across application runs and updates.
- **Automatic Browser Launch**: Upon execution, the application automatically opens in your default web browser for a seamless user experience.
- **Graceful Shutdown**: The server handles termination signals (`Ctrl+C`, etc.) to ensure proper cleanup and data integrity.

## How It Works
This application functions as a single Go executable that bundles both the backend API and the frontend assets.
- **Backend (Go)**: Handles API requests, interacts with the SQLite database, and serves the embedded static files.
- **Frontend (HTML, Tailwind CSS, JavaScript)**: A simple single-page application (SPA) that communicates with the Go backend via Fetch API calls.
- **Static File Embedding**: All files in the `static/` directory (HTML, CSS, JS) are embedded into the Go binary using Go's `embed` package, making the application self-contained.
- **Database Persistence**: The SQLite database file (`tracker.db`) is stored in a platform-specific user configuration directory (e.g., `~/.config/time-tracker` on Linux, `~/Library/Application Support/time-tracker` on macOS, `%APPDATA%\time-tracker` on Windows) to ensure user data is preserved across application versions.
- **Browser Automation**: The Go application automatically detects the operating system and uses the appropriate command (`start`, `open`, `xdg-open`) to launch the default web browser to `http://localhost:8080` (or configured port) after the server has initialized.
- **Graceful Shutdown**: The server listens for OS termination signals (like `Ctrl+C`). Upon receiving such a signal, it attempts to shut down gracefully within a timeout period, ensuring ongoing requests are completed and resources (like the database connection) are properly closed.

## Installation
To get a local copy up and running for development:

1.  **Clone the repository**:
    ```bash
    git clone [https://github.com/NaofalMufid/time-tracker.git](https://github.com/NaofalMufid/time-tracker.git)
    cd time-tracker
    ```

2.  **Install Go dependencies**:
    ```bash
    go mod download
    ```

3.  **Install Node.js dependencies (for Tailwind CSS):**
    ```bash
    npm install
    ```

## Development Setup
If you're making changes to the frontend or backend:

1.  **Frontend (Tailwind CSS):**
    If you modify any HTML or use new Tailwind classes, you'll need to re-compile your CSS. Ensure your `tailwind.config.js` is correctly configured to purge unused CSS.
    ```bash
    npx tailwindcss -i ./static/input.css -o ./static/output.css --minify
    ```
    For development, you can use `npx tailwindcss -i ./static/input.css -o ./static/output.css --watch` to automatically recompile on changes.

2.  **Run the application (for development):**
    ```bash
    go run main.go
    ```
    This will compile and run the application, serving it at `http://localhost:8080` (or the port specified in your `.env` file).

3.  **Build the Executable (for release/distribution):**
    After making changes, compile your production CSS (step 1) and then build the Go executable:
    * **For your current OS:**
        ```bash
        go build -o time-tracker
        ```
    * **Cross-compile for Windows (64-bit):**
        ```bash
        GOOS=windows GOARCH=amd64 go build -o time-tracker.exe
        ```
    * **Cross-compile for macOS (Intel):**
        ```bash
        GOOS=darwin GOARCH=amd64 go build -o time-tracker_darwin_amd64
        ```
    * **Cross-compile for macOS (Apple Silicon):**
        ```bash
        GOOS=darwin GOARCH=arm64 go build -o time-tracker_darwin_arm64
        ```
    * **Cross-compile for Linux (64-bit):**
        ```bash
        GOOS=linux GOARCH=amd64 go build -o time-tracker_linux_amd64
        ```

## Usage

### Running the Executable
Once you have built the executable (e.g., `time-tracker` or `time-tracker.exe`), simply run it from your terminal or by double-clicking it.
Example (from terminal):
```bash
./time-tracker
```

The application will automatically launch in your default web browser at `http://localhost:8080`.

### Data Persistence
Your task data is stored in an SQLite database file named tracker.db. This file is located in a persistent user-specific directory, ensuring your data is saved even if you update or move the application executable.

The typical locations are:
- Windows: `C:\Users\USERNAME\AppData\Roaming\time-tracker\tracker.db`
- macOS: `~/Library/Application Support/time-tracker/tracker.db`
- Linux: `~/.config/time-tracker/tracker.db`

### Stopping the Application Gracefully
To ensure data integrity and proper resource cleanup, please stop the application gracefully:

- If running from a terminal: Press `Ctrl+C`. The application will log a "Shutting down server..." message and then "Server gracefully stopped."

- If running by double-clicking the executable (without a visible terminal): Close the command prompt/terminal window that might have opened in the background. The application process will terminate gracefully.

Simply closing the browser tab will not stop the application. The server will continue to run in the background.

## Project Structure
### Key Directories and Files
- `db/`: Contains database-related code, such as connection setup, migration logic, and queries.
- `handlers/`: Houses HTTP request handlers for managing tasks.
- `models/`: Defines the data models used in the application.
- `static/`: Stores static assets like index.html, input.css, output.css, and script.js for the frontend. These are embedded into the Go executable.
- `utils/`: Includes utility functions, such as JSON handling and pagination logic.
- `.env`: Environment variables (e.g., PORT).
- `main.go`: The entry point of the application, where the server is initialized, routes are defined, and graceful shutdown is configured.
- `go.mod & go.sum`: Go module files for dependency management.
- `package.json & package-lock.json`: Node.js dependency files for Tailwind CSS.
- `tailwind.config.js`: Tailwind CSS configuration.

## API Endpoints
The backend API can be accessed at `http://localhost:8080` (or your configured port).

- GET `/tasks`: Retrieve a list of tasks with pagination, filtering, and sorting.
- POST `/tasks`: Create a new task.
- POST `/tasks/{id}/pause`: Pause a running task.
- POST `/tasks/{id}/resume`: Resume a paused task.
- POST `/tasks/{id}/stop`: Stop a task.
- DELETE `/tasks/{id}`: Delete a task.
- GET `/tasks/running`: Get the currently running task.
- GET `/health`: Health check endpoint.

## Contributing
We welcome contributions! Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/YourFeatureName`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/YourFeatureName`).
5. Open a pull request.

## License
This project is licensed under the MIT License. See the LICENSE file for details.