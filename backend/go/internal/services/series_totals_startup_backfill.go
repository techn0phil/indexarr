package services

import (
	"database/sql"
	"fmt"

	"indexarr/internal/config"
	"indexarr/internal/models"
	"indexarr/internal/repository"
)

// BackfillSeriesTotalsOnStartup is a temporary migration helper.
// It runs only when at least one series has unset total counts and then refreshes
// total season/episode counts from metadata APIs to account for episodes/seasons
// that are not present in the local database.
// This is a temporary measure to repair series total counts after a migration that introduced these fields.
// It should be removed once the migration is stable and all existing installations have been repaired.
func BackfillSeriesTotalsOnStartup(db *sql.DB, cfg *config.Config) error {
	if db == nil || cfg == nil {
		return fmt.Errorf("invalid startup backfill dependencies")
	}

	var invalidCount int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM series
		WHERE total_season_count IS NULL
			OR total_episode_count IS NULL
			OR total_season_count = 0
			OR total_episode_count = 0
	`).Scan(&invalidCount)
	if err != nil {
		return fmt.Errorf("failed to check invalid series totals: %w", err)
	}

	if invalidCount == 0 {
		return nil
	}

	config.GlobalLogger.Debug().Int("series_count", invalidCount).Msg("Starting series totals metadata backfill")

	tmdb := NewTMDBClient(cfg.TMDBAPIKey, cfg.DetectionLanguage, cfg.MetadataLanguage)
	tv := NewTVClient(cfg.TVDBAPIKey, db)

	updatedCount := 0
	failureCount := 0

	filters := &models.FilterCriteria{Page: 1, PageSize: 200}
	for {
		seriesBatch, total, err := repository.GetSeries(db, filters)
		if err != nil {
			return fmt.Errorf("failed to fetch series for totals backfill: %w", err)
		}

		if len(seriesBatch) == 0 {
			break
		}

		for i := range seriesBatch {
			series := &seriesBatch[i]
			beforeSeasonCount := series.TotalSeasonCount
			beforeEpisodeCount := series.TotalEpisodeCount

			if err := enrichSeriesTotalsFromMetadata(series, tmdb, tv, cfg.MetadataLanguage); err != nil {
				failureCount++
				config.GlobalLogger.Warn().Err(err).Int64("series_id", series.ID).Str("title", series.Title).Msg("Failed to enrich series totals during startup backfill")
				continue
			}

			if series.TotalSeasonCount <= 0 || series.TotalEpisodeCount <= 0 {
				failureCount++
				config.GlobalLogger.Warn().Int64("series_id", series.ID).Str("title", series.Title).Msg("Metadata enrichment did not provide valid total counts")
				continue
			}

			if beforeSeasonCount == series.TotalSeasonCount && beforeEpisodeCount == series.TotalEpisodeCount {
				continue
			}

			if err := repository.UpdateSeries(db, series); err != nil {
				failureCount++
				config.GlobalLogger.Warn().Err(err).Int64("series_id", series.ID).Str("title", series.Title).Msg("Failed to persist series totals during startup backfill")
				continue
			}

			updatedCount++
		}

		if int64(filters.Page*filters.PageSize) >= total {
			break
		}
		filters.Page++
	}

	config.GlobalLogger.Debug().Int("updated", updatedCount).Int("failed", failureCount).Msg("Series totals metadata backfill completed")
	return nil
}

func enrichSeriesTotalsFromMetadata(series *models.Series, tmdb *TMDBClient, tv *TVClient, metadataLanguage string) error {
	if series == nil {
		return fmt.Errorf("nil series")
	}

	// Most reliable path: direct TMDB details lookup by known ID.
	if tmdb != nil && tmdb.apiKey != "" && series.TMDBId > 0 {
		details, err := tmdb.GetTVDetails(int(series.TMDBId))
		if err == nil {
			series.TotalSeasonCount = details.NumberOfSeasons
			series.TotalEpisodeCount = details.NumberOfEpisodes
			return nil
		}
		config.GlobalLogger.Debug().Err(err).Int64("series_id", series.ID).Int64("tmdb_id", series.TMDBId).Msg("TMDB details lookup failed during totals backfill")
	}

	// Fallback: direct TVDB episodes lookup by known ID.
	if tv != nil && tv.apiKey != "" && series.TVDBId > 0 {
		allEpisodes, err := tv.GetAllEpisodes(int(series.TVDBId), metadataLanguage)
		if err == nil {
			seasons := make(map[int]bool)
			for _, episode := range allEpisodes.Data.Episodes {
				if episode.SeasonNumber > 0 {
					seasons[episode.SeasonNumber] = true
				}
			}
			series.TotalEpisodeCount = len(allEpisodes.Data.Episodes)
			series.TotalSeasonCount = len(seasons)
			return nil
		}
		config.GlobalLogger.Debug().Err(err).Int64("series_id", series.ID).Int64("tvdb_id", series.TVDBId).Msg("TVDB episodes lookup failed during totals backfill")
	}

	// Last fallback: search by title/year through existing enrichment flow.
	if tmdb != nil && tmdb.apiKey != "" {
		if err := tmdb.EnrichSeries(series); err == nil {
			return nil
		} else {
			config.GlobalLogger.Debug().Err(err).Int64("series_id", series.ID).Str("title", series.Title).Msg("TMDB title/year enrichment fallback failed during totals backfill")
		}
	}

	return fmt.Errorf("no metadata source available for totals backfill")
}
