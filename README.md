<a id="readme-top"></a>

- [About The Project](#about-the-project)
  - [Assumptions](#assumptions)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
    - [Golang](#golang)
    - [PostgreSQL](#postgresql)
  - [Usage](#usage)
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

- Each user can only borrow one copy per book.

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

# Roadmap

- [x] Implement simple in-memory storage with pre-populated books on startup.
- [x] Integrate PostgreSQL for persistent data storage.
- [ ] Implement logging for API requests and responses.
- [ ] Add validation for each API (e.g., missing or invalid input).
- [ ] Write tests for coverage and regression.
- [ ] Integrate SingPass MyInfo for user authentication.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Acknowledgement

|                 Tool                  |     Description     |
| :-----------------------------------: | :-----------------: |
| [mockaroo](https://www.mockaroo.com/) | Mock data generator |

## Packages

|                    Package                     |                Description                |
| :--------------------------------------------: | :---------------------------------------: |
|       [Fiber](https://docs.gofiber.io/)        | Fast and lightweight web framework for Go |
|    [zerolog](https://github.com/rs/zerolog)    |    Zero-allocation JSON logger for Go     |
| [testify](https://github.com/stretchr/testify) |            Go testing toolkit             |
|   [dbdiagram.io](https://dbdiagram.io/home)    |  A free, simple tool to draw ER diagrams  |
