# Cyware Onboarding Task

This project is a backend system for a Stack Overflow-like application, built with Go.


## Features Implemented

* **User Management:**
    * User registration with secure `bcrypt` password hashing.
    * User login with JWT generation.
* **Content Management:**
    * Authenticated users can post new questions.
    * Authenticated users can post answers to existing questions.
    * Public endpoint to fetch a question along with all its answers and their vote scores.
* **Voting System:**
    * Authenticated users can upvote or downvote any question or answer.
    * The system enforces a "one vote per user per post" rule and allows users to change their vote.
* **LLM Summarization:**
    * A public endpoint that takes a question ID and uses an LLM (via OpenRouter) to generate a summary of the question and its answers.

### Database Schema

The database is designed to be relational and normalized, using foreign key constraints to ensure data integrity.

* **`users` Table:** Stores user credentials and information.
    * `id (UUID, Primary Key)`: Unique, non-sequential identifier for each user.
    * `username (VARCHAR, Unique)`: The user's public name.
    * `email (VARCHAR, Unique)`: The user's login email.
    * `password_hash (VARCHAR)`: The securely hashed user password.
* **`questions` Table:** Stores all questions.
    * `id (UUID, Primary Key)`: Unique identifier for the question.
    * `user_id (UUID, Foreign Key)`: Links to the `id` of the user who posted it. `ON DELETE CASCADE` ensures a user's questions are deleted if the user is.
    * `body (VARCHAR, not null)`: The content of the question.

* **`answers` Table:** Stores all answers.
    * `id (UUID, Primary Key)`: Unique identifier for the answer.
    * `question_id (UUID, Foreign Key)`: Links to the `id` of the question it answers.
    * `user_id (UUID, Foreign Key)`: Links to the `id` of the user who posted it.
    * `body (VARCHAR, not null)`: The content of the answer.
* **`votes` Table:** A polymorphic table to store votes for both questions and answers.
    * `user_id (UUID, Foreign Key)`: The user who voted.
    * `post_id (UUID)`: The ID of the question or answer being voted on.
    * `post_type (VARCHAR)`: A string ('question' or 'answer') to differentiate the post type.
    * `vote_type (SMALLINT)`: Stores `1` for an upvote or `-1` for a downvote.
    * `PRIMARY KEY (user_id, post_id)`: A composite key that ensures a user can only vote once per post.

### API Endpoints

The API is versioned under the `/api` prefix.

#### Public Endpoints (No Authentication Required)

| Method | Path                                   | Description                                       |
| :----- | :------------------------------------- | :------------------------------------------------ |
| `POST` | `/users/register`                      | Creates a new user account.                       |
| `POST` | `/users/login`                         | Authenticates a user and returns a JWT.           |
| `GET`  | `/questions/:questionid`               | Fetches a single question and all of its answers. |
| `GET`  | `/questions/:questionid/summarize`     | Returns an LLM-generated summary of the thread.   |

#### Authenticated Endpoints (Valid JWT Required)

| Method | Path                                   | Description                               |
| :----- | :------------------------------------- | :---------------------------------------- |
| `POST` | `/questions`                           | Creates a new question.                   |
| `POST` | `/questions/:questionid/answer`        | Posts a new answer to a specific question. |
| `POST` | `/questions/:questionid/vote`          | Upvotes or downvotes a specific question. |
| `POST` | `/answers/:answerid/vote`              | Upvotes or downvotes a specific answer.   |

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

### 1. User Registration

**Example Request:**
```bash
curl -X POST http://localhost:8080/register \
-H "Content-Type: application/json" \
-d '{
  "username": "testuser",
  "email": "test@example.com",
  "password": "a-secure-password"
}'
```


### 2. User Login


**Example Request:**
```bash
curl -X POST http://localhost:8080/api/users/login \
-H "Content-Type: application/json" \
-d '{
  "email": "test@example.com",
  "password": "a-secure-password"
}'
```

### 3.Question Create


**Example Request:**
```bash
curl -X POST http://localhost:8080/api/questions \
-H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
-H "Content-Type: application/json" \
-d '{
  "title": "How to test Go APIs?",
  "body": "What are the best practices for testing a Go backend from the command line?"
}'
```
### 4. Answer


**Example Request:**
```bash
curl -X POST http://localhost:8080/api/questions/<QUESTION_ID>/answer \
-H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
-H "Content-Type: application/json" \
-d '{
  "body": "Using curl with jq is a great way to script tests for your API endpoints."
}'

```
### 5. Voting

**Example Request:**
```bash
# Upvote the question
curl -i -X POST http://localhost:8080/api/questions/<QUESTION_ID>/vote \
-H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
-H "Content-Type: application/json" \
-d '{"vote_type": 1}'

# Upvote the answer
curl -i -X POST http://localhost:8080/api/answers/<ANSWER_ID>/vote \
-H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
-H "Content-Type: application/json" \
-d '{"vote_type": 1}'
```

### 6. Get Question

**Example Request:**
```bash
curl -X GET http://localhost:8080/api/questions/<QUESTION_ID> | jq
```

### 7. LLM Summary

**Example Request:**
```bash
curl -X GET http://localhost:8080/api/questions/<QUESTION_ID>/summarize | jq
```

