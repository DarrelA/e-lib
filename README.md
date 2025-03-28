<a id="readme-top"></a>

- [About The Project](#about-the-project)
  - [Assumptions](#assumptions)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
    - [Golang](#golang)
    - [PostgreSQL](#postgresql)
  - [Usage](#usage)
- [Integration Test](#integration-test)
  - [Test Cases](#test-cases)
  - [JQ](#jq)
- [Roadmap](#roadmap)
- [Acknowledgement](#acknowledgement)
  - [Packages](#packages)

# About The Project

This is a simple RESTful API to manage loan of e-book in an electronic library. The API allow users to:

1. Search for availability based on book title.
2. Borrow a book.
3. Extend a book loan.
4. Return a book.

## Assumptions

- Each user may only borrow one copy per book.
- Each user may only extend each book loan once.

# Getting Started

## Prerequisites

### Golang

- [Go Documentation -> Download and install](https://go.dev/doc/install)
- [Homebrew - Go](https://formulae.brew.sh/formula/go)

### PostgreSQL

- [PostgreSQL Downloads](https://www.postgresql.org/download/)
- [Homebrew - postgresql@14](https://formulae.brew.sh/formula/postgresql@14)

```sh
psql -U postgres
CREATE USER myuser WITH PASSWORD 'mypassword';
CREATE DATABASE elib WITH OWNER myuser;
GRANT ALL PRIVILEGES ON DATABASE elib TO myuser;

psql -d elib -U myuser
\l  # List databases
\du # List users
\dt # List tables
SELECT * FROM users LIMIT 10;
SELECT * FROM books LIMIT 10;
SELECT * FROM loans LIMIT 10;
\d loans # Describe the table
\d+ loans # Describe the table
```

![elib-er-diagram](/docs/images/elib-er-diagram.svg)

```sh
dbdocs db2dbml postgres "postgresql://myuser:mypassword@localhost/elib" -o database.dbml
```

## Usage

Refer to `Makefile` for all the commands.

- `make dev`
- `GET`: `localhost:3000/Book?title=Badlands`
- `POST`: `localhost:3000/Borrow` (with JSON body)
- `POST`: `localhost:3000/Extend` (with JSON body)
- `POST`: `localhost:3000/Return` (with JSON body)

**JSON Body Example (Common for POST requests):**

```json
{
  "title": "Badlands"
}
```

# Integration Test

Mocking Google OAuth2 in our integration tests allows us to rigorously validate how our backend handles user profile retrieval for seamless session management. By simulating scenarios like service unavailability or incomplete data, we ensure reliable testing of our logic without external dependencies that introduce unpredictability. This approach accelerates test cycles, reduces maintenance costs tied to third-party changes, and safeguards against disruptions in our development workflow. It complements broader validations by isolating critical authentication paths, ensuring we focus engineering effort where it matters most while maintaining confidence in system-wide integrity.

## Test Cases

- [e-lib-test-cases spreadsheet](https://docs.google.com/spreadsheets/d/1qSSr5BKv9U1xnTNzGl7a-ubbvUVa_WLRL3-GkoI4g54/edit?usp=sharing)

## JQ

- [./jq](https://jqlang.org/download/)

```sh
chmod +x ./deployment/scripts/integration_test.sh
chmod +x ./deployment/scripts/wrap_test_for_coverage.sh

# Get Postgres container id
docker ps
# Execute a command inside a running Docker container
docker exec -it <your_container_id> psql -U <your_username> -d elib
```

# Roadmap

- [x] Implement simple in-memory storage with pre-populated books on startup.
- [x] Integrate PostgreSQL for persistent data storage.
- [x] Implement logging for API requests and responses.
- [x] Add validation for each API (e.g., missing or invalid input).
- [x] Write tests for coverage and regression.
- [x] Integrate Google OAuth2 for user authentication.
- [x] Implement session management with Redis.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Acknowledgement

|                 Tool                  |     Description     |
| :-----------------------------------: | :-----------------: |
| [mockaroo](https://www.mockaroo.com/) | Mock data generator |

## Packages

|                         Package                         |                            Description                            |
| :-----------------------------------------------------: | :---------------------------------------------------------------: |
|            [Fiber](https://docs.gofiber.io/)            |             Fast and lightweight web framework for Go             |
|        [zerolog](https://github.com/rs/zerolog)         |                Zero-allocation JSON logger for Go                 |
|     [testify](https://github.com/stretchr/testify)      |                        Go testing toolkit                         |
| [validator](https://github.com/go-playground/validator) | Value validations for structs and individual fields based on tags |
|       [redis](https://github.com/redis/go-redis)        |      Golang Redis client for Redis Server and Redis Cluster       |
|        [dbdiagram.io](https://dbdiagram.io/home)        |              A free, simple tool to draw ER diagrams              |
