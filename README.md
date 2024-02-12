# Simple Go URL Shortener

## Setting up the database

To set up the database, run the following DDL queries in sqlite in your terminal.

```sql
-- Create the urls table which will store the URLs (shortened and unshortened) 
-- along with the clicks on the shortened URL.
CREATE TABLE IF NOT EXISTS "urls" (
    original_url TEXT PRIMARY KEY NOT NULL, 
    shortened_url TEXT NOT NULL, 
    clicks INTEGER DEFAULT 0, 
    CONSTRAINT uniq_original_url UNIQUE (original_url)
);

-- Add an index on the shortened_url column, as it's used to update the clicks
-- for the shortened URL.
CREATE index idx_shortened ON urls (shortened_url);
```