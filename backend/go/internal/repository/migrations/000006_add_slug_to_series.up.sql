-- Add slug column to series table
ALTER TABLE series ADD COLUMN slug TEXT;
UPDATE series SET slug = '';
