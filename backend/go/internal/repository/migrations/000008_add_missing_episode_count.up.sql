ALTER TABLE series ADD COLUMN missing_episode_count INTEGER;

-- Update missing_episode_count for existing series
UPDATE series AS s SET missing_episode_count = (
    SELECT COUNT(*)
    FROM episodes e
    WHERE e.series_id = s.id AND e.status = 'missing'
);
