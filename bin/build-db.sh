#!/bin/sh

# Build the SQLite database
sqlite3 $DATABASE_FILE < bin/load.sql