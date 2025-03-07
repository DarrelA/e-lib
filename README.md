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

## Usage

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Roadmap

- [x] Implement simple in-memory storage with pre-populated books on startup.
- [ ] Integrate PostgreSQL for persistent data storage.
- [ ] Implement logging for API requests and responses.
- [ ] Add validation for each API (e.g., missing or invalid input).
- [ ] Write tests for coverage and regression.
- [ ] Integrate SingPass MyInfo for user authentication.

# Acknowledgement

- [Fiber](https://docs.gofiber.io/)
- [zerolog](https://github.com/rs/zerolog)
- [mockaroo](https://www.mockaroo.com/)
