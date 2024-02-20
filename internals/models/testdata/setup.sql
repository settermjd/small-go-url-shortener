-- Create the urls table which will store the URLs (shortened and unshortened) 
-- along with the clicks on the shortened URL.
CREATE TABLE IF NOT EXISTS "urls" (
    original_url TEXT PRIMARY KEY NOT NULL,             -- the original URL that was shortened 
    shortened_url TEXT NOT NULL,                        -- the shortened URL 
    clicks INTEGER DEFAULT 0,                           -- stores the number of times the short URL has been clicked 
    created DATETIME DEFAULT CURRENT_TIMESTAMP,         -- marks when the record was first created
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,         -- marks when the record was last updated
    CONSTRAINT uniq_original_url UNIQUE (original_url)
);

-- Add an index on the shortened_url column, as it's used to update the clicks
-- for the shortened URL.
CREATE index idx_shortened ON urls (shortened_url);

-- Create a trigger to set the value of the updated column to the current date/time when a row is updated
CREATE TRIGGER IF NOT EXISTS trig_urls_update 
    AFTER UPDATE 
    ON urls
BEGIN
    UPDATE urls 
    SET updated = DATETIME('NOW') 
    WHERE original_url = old.original_url;
END;

INSERT INTO urls (original_url, shortened_url, clicks)
VALUES (
        'https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/424',
        'https://4C2P1PC8+',
        0
    );