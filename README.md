# Simple Go URL Shortener

This is a small, simplistic URL shortener written in Go.
It doesn't have near the functionality of sites such as bit.ly.
However, it does a good job of showing the essence of how a URL shortener works.

## Prerequisites

To use the application, you'll need the following:

- A recent version of [Go](https://go.dev/dl/). 
  1.22.0 is the current version at the time of writing.
- [The Command Line Shell for SQLite](https://www.sqlite.org/cli.html).
  This is required for provisioning the SQLite database during setup.

To develop the application, you'll need the following, additional, dependencies:

- [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm).
  This is required to install and run the frontend tooling.

## Getting started

To set up the application, clone it locally, change into the cloned project directory, and install the required Go modules, by running the following commands.

```bash
git clone git@github.com:settermjd/small-go-url-shortener.git
cd small-go-url-shortener
go mod tidy
```

## Setting up the database

To set up the database, run the following command in your terminal.

```bash
sqlite3 data/database.sqlite3 < docs/database/load.sql
```

This will provision the database, creating the (sole) table, `urls`.
It won't load any default data into the table, however. 
So, you should insert some records by running the following commands, after replacing the placeholders with data of your choice.

```sql
INSERT INTO urls (original_url, shortened_url, clicks) VALUES
("<<ORIGINAL URL>>", "https://shoRtkl9187ds", 347),
("<<ORIGINAL URL>>", "https://sh0Rtkl9187es", 2809);
```