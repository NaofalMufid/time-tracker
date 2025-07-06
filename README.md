# Task Time Tracker

Task Time Tracker is a Go-based application designed to help users manage and track their tasks efficiently. This project provides a simple yet powerful interface for task management, including features like task creation, updating, and deletion.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

## Features
- **Task Management**: Create, update, and delete tasks.
- **Pagination**: Efficiently handle large numbers of tasks with pagination.
- **JSON Handling**: Easy JSON parsing and response handling.
- **Static Files**: Serve static HTML, CSS, and JavaScript files for the frontend.

## Installation
1. **Clone the repository**:
   ```bash
   git clone https://github.com/NaofalMufid/time-tracker.git
   cd time-tracker
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Run the application**:
   ```bash
   go run main.go
   ```

## Usage
Once the application is running, you can access the frontend by navigating to `http://localhost:8080` in your web browser. The backend API can be accessed at `http://localhost:8080/api`.

### API Endpoints
- **GET /api/tasks**: Retrieve a list of tasks.
- **POST /api/tasks**: Create a new task.
- **PUT /api/tasks/{id}**: Update an existing task.
- **DELETE /api/tasks/{id}**: Delete a task.

## Project Structure
### Key Directories and Files
- **`db/`**: Contains database-related code, such as connection setup and queries.
- **`handlers/`**: Houses HTTP request handlers for managing tasks.
- **`models/`**: Defines the data models used in the application.
- **`static/`**: Stores static assets like HTML, CSS, and JavaScript for the frontend.
- **`utils/`**: Includes utility functions, such as JSON handling and pagination logic.
- **`main.go`**: The entry point of the application, where the server is initialized and routes are defined.
- **`go.mod` & `go.sum`**: Go module files for dependency management.
- **`README.md`**: Project documentation and usage guide.

## API Documentation
For detailed API documentation, refer to the [API Documentation](API_DOCS.md) file.

## Contributing
We welcome contributions! Please follow these steps to contribute:
1. Fork the repository.
2. Create a new branch (`git checkout -b feature/YourFeatureName`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/YourFeatureName`).
5. Open a pull request.

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
