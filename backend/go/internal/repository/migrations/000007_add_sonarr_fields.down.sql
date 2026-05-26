-- Remove Sonarr integration fields from series table
DROP INDEX IF EXISTS idx_series_sonarr_id;

ALTER TABLE series DROP COLUMN sonarr_id;
ALTER TABLE series DROP COLUMN title_slug;
