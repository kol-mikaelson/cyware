# Cyware Onboarding Task

This project is a backend system for a Stack Overflow-like application, built with Go.


## Current Status

The project is currently in the initial development phase. The following features have been implemented:

* **User Registration:** New users can create an account. Passwords are securely hashed using `bcrypt`.
* **User Login:** Registered users can log in to receive a JWT for authenticating future requests.

## How to Run the Project

This project is fully containerized, so you only need Docker and Docker Compose installed to run it.

### Prerequisites

* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/) (usually included with Docker Desktop)

### Steps

1.  **Clone the repository:**
    ```bash
    git clone <your-repository-url>
    cd <repository-folder>
    ```

2.  **Build and run the services:**
    From the root of the project directory, run the following command:
    ```bash
    docker compose up --build
    ```
    This command will build the Go application image, pull the PostgreSQL image, and start both services. The API will be available at `http://localhost:8080`.

## Available API Endpoints

You can use a tool like `curl` or Postman to interact with the API.

### 1. User Registration

* **Endpoint:** `POST /register`
* **Description:** Creates a new user account.

**Example Request:**
```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "a-secure-password"
}'