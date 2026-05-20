-- Add Sonarr integration fields to series table
ALTER TABLE series ADD COLUMN sonarr_id INTEGER;
ALTER TABLE series ADD COLUMN title_slug TEXT;

-- Create index for Sonarr ID lookups
CREATE INDEX idx_series_sonarr_id ON series(sonarr_id);
