-- Replace language name by language code
UPDATE audio_tracks
SET language = CASE
    WHEN language = 'English' THEN 'en'
    WHEN language = 'French' THEN 'fr'
    WHEN language = 'Spanish' THEN 'es'
    WHEN language = 'German' THEN 'de'
    WHEN language = 'Italian' THEN 'it'
    WHEN language = 'Japanese' THEN 'ja'
    WHEN language = 'Korean' THEN 'ko'
    WHEN language = 'Chinese' THEN 'zh'
    ELSE language
END;

UPDATE subtitle_tracks
SET language = CASE
    WHEN language = 'English' THEN 'en'
    WHEN language = 'French' THEN 'fr'
    WHEN language = 'Spanish' THEN 'es'
    WHEN language = 'German' THEN 'de'
    WHEN language = 'Italian' THEN 'it'
    WHEN language = 'Japanese' THEN 'ja'
    WHEN language = 'Korean' THEN 'ko'
    WHEN language = 'Chinese' THEN 'zh'
    ELSE language
END;


-- Replace "Unknown" value by empty value ""
UPDATE video_tracks
SET codec = ''
WHERE codec = 'Unknown';

UPDATE video_tracks
SET resolution = ''
WHERE resolution = 'Unknown';

UPDATE video_tracks
SET bitrate = ''
WHERE bitrate = 'Unknown';

UPDATE audio_tracks
SET codec = ''
WHERE codec = 'Unknown';

UPDATE audio_tracks
SET bitrate = ''
WHERE bitrate = 'Unknown';

UPDATE audio_tracks
SET sample_rate = ''
WHERE sample_rate = 'Unknown';

UPDATE audio_tracks
SET language = ''
WHERE language = 'Unknown';

UPDATE subtitle_tracks
SET language = ''
WHERE language = 'Unknown';
