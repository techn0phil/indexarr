-- Replace language code by language name
UPDATE audio_tracks
SET language = CASE
    WHEN language = 'en' THEN 'English'
    WHEN language = 'fr' THEN 'French'
    WHEN language = 'es' THEN 'Spanish'
    WHEN language = 'de' THEN 'German'
    WHEN language = 'it' THEN 'Italian'
    WHEN language = 'ja' THEN 'Japanese'
    WHEN language = 'ko' THEN 'Korean'
    WHEN language = 'zh' THEN 'Chinese'
    ELSE language
END;

UPDATE subtitle_tracks
SET language = CASE
    WHEN language = 'en' THEN 'English'
    WHEN language = 'fr' THEN 'French'
    WHEN language = 'es' THEN 'Spanish'
    WHEN language = 'de' THEN 'German'
    WHEN language = 'it' THEN 'Italian'
    WHEN language = 'ja' THEN 'Japanese'
    WHEN language = 'ko' THEN 'Korean'
    WHEN language = 'zh' THEN 'Chinese'
    ELSE language
END;


-- Replace empty value "" value by "Unknown" value
UPDATE video_tracks
SET codec = 'Unknown'
WHERE codec = '';

UPDATE video_tracks
SET resolution = 'Unknown'
WHERE resolution = '';

UPDATE video_tracks
SET bitrate = 'Unknown'
WHERE bitrate = '';

UPDATE audio_tracks
SET codec = 'Unknown'
WHERE codec = '';

UPDATE audio_tracks
SET bitrate = 'Unknown'
WHERE bitrate = '';

UPDATE audio_tracks
SET sample_rate = 'Unknown'
WHERE sample_rate = '';

UPDATE audio_tracks
SET language = 'Unknown'
WHERE language = '';

UPDATE subtitle_tracks
SET language = 'Unknown'
WHERE language = '';
