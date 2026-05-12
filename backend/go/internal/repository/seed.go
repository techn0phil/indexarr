package repository

import (
	"database/sql"
	"fmt"

	"indexarr/internal/models"
)

func SeedMockData(db *sql.DB) error {
	// Check if data already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Already seeded
	}

	// Insert movies
	movies := getMockMovies()
	for _, m := range movies {
		err := insertMovie(db, &m)
		if err != nil {
			return err
		}
	}

	// Insert series
	series := getMockSeries()
	for _, s := range series {
		err := insertSeries(db, &s)
		if err != nil {
			return err
		}
	}

	fmt.Println("✅ Mock data seeded successfully")
	return nil
}

func insertMovie(db *sql.DB, m *models.Movie) error {
	result, err := db.Exec(`
		INSERT INTO movies (title, year, duration, synopsis, genres, rating, popularity, status, file_size, file_path, container, date_added, tmdb_id, imdb_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.Title, m.Year, m.Duration, m.Synopsis, m.Genres, m.Rating, m.Popularity, m.Status, m.FileSize, m.FilePath, m.Container, m.DateAdded, m.TMDBId, m.IMDbId)

	if err != nil {
		return err
	}

	movieID, _ := result.LastInsertId()

	// Insert video tracks
	if m.MediaInfo != nil && len(m.MediaInfo.VideoTracks) > 0 {
		for _, vt := range m.MediaInfo.VideoTracks {
			_, err := db.Exec(`
				INSERT INTO video_tracks (movie_id, codec, resolution, fps, bitrate, hdr, color_space)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, movieID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
			if err != nil {
				return err
			}
		}
	}

	// Insert audio tracks
	if m.MediaInfo != nil && len(m.MediaInfo.AudioTracks) > 0 {
		for _, at := range m.MediaInfo.AudioTracks {
			_, err := db.Exec(`
				INSERT INTO audio_tracks (movie_id, codec, channels, language, sample_rate, bitrate)
				VALUES (?, ?, ?, ?, ?, ?)
			`, movieID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
			if err != nil {
				return err
			}
		}
	}

	// Insert cast
	for _, c := range m.Cast {
		_, err := db.Exec(`
			INSERT INTO cast (movie_id, name, role, avatar)
			VALUES (?, ?, ?, ?)
		`, movieID, c.Name, c.Role, c.Avatar)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertSeries(db *sql.DB, s *models.Series) error {
	result, err := db.Exec(`
		INSERT INTO series (title, year_start, year_end, season_count, episode_count, synopsis, genres, rating, popularity, status, file_size, date_added, tvdb_id, imdb_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.Title, s.YearStart, s.YearEnd, s.SeasonCount, s.EpisodeCount, s.Synopsis, s.Genres, s.Rating, s.Popularity, s.Status, s.FileSize, s.DateAdded, s.TVDBId, s.IMDbId)

	if err != nil {
		return err
	}

	seriesID, _ := result.LastInsertId()

	// Insert cast
	for _, c := range s.Cast {
		_, err := db.Exec(`
			INSERT INTO cast (series_id, name, role, avatar)
			VALUES (?, ?, ?, ?)
		`, seriesID, c.Name, c.Role, c.Avatar)
		if err != nil {
			return err
		}
	}

	// Insert seasons and episodes
	for _, season := range s.Seasons {
		_, err := db.Exec(`
			INSERT INTO seasons (series_id, number, file_size)
			VALUES (?, ?, ?)
		`, seriesID, season.Number, season.FileSize)
		if err != nil {
			return err
		}

		// Insert episodes
		for _, ep := range season.Episodes {
			epResult, err := db.Exec(`
				INSERT INTO episodes (series_id, season_num, episode_num, title, duration, status, file_size, file_path, date_added)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, seriesID, ep.SeasonNum, ep.EpisodeNum, ep.Title, ep.Duration, ep.Status, ep.FileSize, ep.FilePath, ep.DateAdded)
			if err != nil {
				return err
			}

			epID, _ := epResult.LastInsertId()

			// Insert video tracks
			if ep.MediaInfo != nil && len(ep.MediaInfo.VideoTracks) > 0 {
				for _, vt := range ep.MediaInfo.VideoTracks {
					_, err := db.Exec(`
						INSERT INTO video_tracks (episode_id, codec, resolution, fps, bitrate, hdr, color_space)
						VALUES (?, ?, ?, ?, ?, ?, ?)
					`, epID, vt.Codec, vt.Resolution, vt.FPS, vt.Bitrate, vt.HDR, vt.ColorSpace)
					if err != nil {
						return err
					}
				}
			}

			// Insert audio tracks
			if ep.MediaInfo != nil && len(ep.MediaInfo.AudioTracks) > 0 {
				for _, at := range ep.MediaInfo.AudioTracks {
					_, err := db.Exec(`
						INSERT INTO audio_tracks (episode_id, codec, channels, language, sample_rate, bitrate)
						VALUES (?, ?, ?, ?, ?, ?)
					`, epID, at.Codec, at.Channels, at.Language, at.SampleRate, at.Bitrate)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func getMockMovies() []models.Movie {
	// now := time.Now().Format(time.RFC3339)

	return []models.Movie{
		// {
		// 	Title:      "Interstellar",
		// 	Year:       2014,
		// 	Duration:   169,
		// 	Synopsis:   "Un groupe d'explorateurs emprunte un tunnel de ver récemment découvert pour dépasser les limites de l'exploration spatiale humaine.",
		// 	Genres:     "Science-fiction, Aventure",
		// 	Rating:     8.7,
		// 	Popularity: 97.4,
		// 	Status:     "available",
		// 	FileSize:   62500000000, // 62.5 GB
		// 	FilePath:   "/media/films/Interstellar (2014)/",
		// 	Container:  "MKV",
		// 	DateAdded:  now,
		// 	TMDBId:     157336,
		// 	IMDbId:     "tt0816692",
		// 	Cast: []models.Cast{
		// 		{Name: "Matthew McConaughey", Role: "Cooper", Avatar: "MM"},
		// 		{Name: "Anne Hathaway", Role: "Brand", Avatar: "AH"},
		// 		{Name: "Jessica Chastain", Role: "Murph", Avatar: "JC"},
		// 		{Name: "Michael Caine", Role: "Prof. Brand", Avatar: "MC"},
		// 	},
		// 	MediaInfo: &models.MediaInfo{
		// 		VideoTracks: []models.VideoTrack{
		// 			{Codec: "H.265", Resolution: "3840x2160", FPS: 23.976, Bitrate: "35.2 Mbps", HDR: "Dolby Vision + HDR10", ColorSpace: "BT.2020"},
		// 		},
		// 		AudioTracks: []models.AudioTrack{
		// 			{Codec: "TrueHD Atmos", Channels: "7.1", Language: "English", SampleRate: "48000 Hz", Bitrate: "4.8 Mbps"},
		// 			{Codec: "AC-3", Channels: "5.1", Language: "French", SampleRate: "48000 Hz", Bitrate: "384 kbps"},
		// 		},
		// 		SubtitleTracks: []models.SubtitleTrack{
		// 			{Language: "French", Format: "SRT"},
		// 			{Language: "English", Format: "SRT"},
		// 		},
		// 	},
		// },
		// {
		// 	Title:      "Dune: Part Two",
		// 	Year:       2024,
		// 	Duration:   166,
		// 	Synopsis:   "Paul Atreides voyage en Arrakis pour venger la mort de sa famille.",
		// 	Genres:     "Science-fiction, Aventure",
		// 	Rating:     8.1,
		// 	Popularity: 156.3,
		// 	Status:     "available",
		// 	FileSize:   48900000000,
		// 	FilePath:   "/media/films/Dune Part Two (2024)/",
		// 	Container:  "MKV",
		// 	DateAdded:  now,
		// 	TMDBId:     693134,
		// 	IMDbId:     "tt15239678",
		// 	Cast: []models.Cast{
		// 		{Name: "Timothée Chalamet", Role: "Paul", Avatar: "TC"},
		// 		{Name: "Zendaya", Role: "Chani", Avatar: "Z"},
		// 	},
		// 	MediaInfo: &models.MediaInfo{
		// 		VideoTracks: []models.VideoTrack{
		// 			{Codec: "AV1", Resolution: "3840x2160", FPS: 23.976, Bitrate: "32.5 Mbps", HDR: "HDR10+", ColorSpace: "BT.2020"},
		// 		},
		// 		AudioTracks: []models.AudioTrack{
		// 			{Codec: "DTS-HD MA", Channels: "5.1", Language: "English", SampleRate: "48000 Hz", Bitrate: "2.5 Mbps"},
		// 		},
		// 	},
		// },
		// {
		// 	Title:      "Oppenheimer",
		// 	Year:       2023,
		// 	Duration:   180,
		// 	Synopsis:   "Le parcours de J. Robert Oppenheimer et son rôle dans le Projet Manhattan.",
		// 	Genres:     "Drame, Histoire",
		// 	Rating:     8.1,
		// 	Popularity: 89.2,
		// 	Status:     "available",
		// 	FileSize:   35700000000,
		// 	FilePath:   "/media/films/Oppenheimer (2023)/",
		// 	Container:  "MKV",
		// 	DateAdded:  now,
		// 	TMDBId:     872585,
		// 	IMDbId:     "tt15398776",
		// 	Cast: []models.Cast{
		// 		{Name: "Cillian Murphy", Role: "Oppenheimer", Avatar: "CM"},
		// 	},
		// 	MediaInfo: &models.MediaInfo{
		// 		VideoTracks: []models.VideoTrack{
		// 			{Codec: "H.264", Resolution: "1920x1080", FPS: 24.0, Bitrate: "15.2 Mbps", HDR: "none", ColorSpace: "BT.709"},
		// 		},
		// 		AudioTracks: []models.AudioTrack{
		// 			{Codec: "AAC", Channels: "2.0", Language: "English", SampleRate: "48000 Hz", Bitrate: "320 kbps"},
		// 		},
		// 	},
		// },
		// {
		// 	Title:      "Poor Things",
		// 	Year:       2023,
		// 	Duration:   141,
		// 	Synopsis:   "Une jeune femme excentrique se lance dans une série d'aventures surréalistes.",
		// 	Genres:     "Fantastique, Comédie",
		// 	Rating:     7.5,
		// 	Popularity: 42.1,
		// 	Status:     "missing",
		// 	FileSize:   0,
		// 	FilePath:   "",
		// 	Container:  "",
		// 	DateAdded:  now,
		// 	TMDBId:     771452,
		// 	IMDbId:     "tt14209916",
		// 	Cast:       []models.Cast{},
		// 	MediaInfo:  nil,
		// },
	}
}

func getMockSeries() []models.Series {
	// now := time.Now().Format(time.RFC3339)

	return []models.Series{
		// {
		// 	Title:        "Breaking Bad",
		// 	YearStart:    2008,
		// 	YearEnd:      2013,
		// 	SeasonCount:  5,
		// 	EpisodeCount: 62,
		// 	Synopsis:     "Walter White, professeur de chimie, fabrique et vend de la méthamphétamine.",
		// 	Genres:       "Drame, Crime",
		// 	Rating:       9.5,
		// 	Popularity:   425.8,
		// 	Status:       "complete",
		// 	FileSize:     142000000000, // 142 GB
		// 	DateAdded:    now,
		// 	TVDBId:       81189,
		// 	IMDbId:       "tt0903747",
		// 	Cast: []models.Cast{
		// 		{Name: "Bryan Cranston", Role: "Walter White", Avatar: "BC"},
		// 		{Name: "Aaron Paul", Role: "Jesse Pinkman", Avatar: "AP"},
		// 	},
		// 	Seasons: []models.Season{
		// 		{
		// 			Number:       1,
		// 			FileSize:     25000000000,
		// 			AvailableEps: 7,
		// 			MissingEps:   0,
		// 			Episodes:     getBreakingBadSeason1(),
		// 		},
		// 		{
		// 			Number:       2,
		// 			FileSize:     30000000000,
		// 			AvailableEps: 4, // 1 missing
		// 			MissingEps:   1,
		// 			Episodes:     getBreakingBadSeason2(),
		// 		},
		// 		{
		// 			Number:       3,
		// 			FileSize:     28000000000,
		// 			AvailableEps: 2,
		// 			MissingEps:   0,
		// 			Episodes:     getBreakingBadSeason3(),
		// 		},
		// 		{
		// 			Number:       4,
		// 			FileSize:     29000000000,
		// 			AvailableEps: 2,
		// 			MissingEps:   0,
		// 			Episodes:     getBreakingBadSeason4(),
		// 		},
		// 		{
		// 			Number:       5,
		// 			FileSize:     30000000000,
		// 			AvailableEps: 3,
		// 			MissingEps:   0,
		// 			Episodes:     getBreakingBadSeason5(),
		// 		},
		// 	},
		// },
		// {
		// 	Title:        "Severance",
		// 	YearStart:    2022,
		// 	YearEnd:      2024,
		// 	SeasonCount:  2,
		// 	EpisodeCount: 10,
		// 	Synopsis:     "Les employés acceptent une procédure chirurgicale pour séparer complètement leur mémoire professionnelle.",
		// 	Genres:       "Science-fiction, Thriller",
		// 	Rating:       8.7,
		// 	Popularity:   134.2,
		// 	Status:       "complete",
		// 	FileSize:     125000000000,
		// 	DateAdded:    now,
		// 	TVDBId:       372310,
		// 	IMDbId:       "tt11280740",
		// 	Cast: []models.Cast{
		// 		{Name: "Adam Scott", Role: "Mark Scout", Avatar: "AS"},
		// 	},
		// 	Seasons: []models.Season{
		// 		{Number: 1, FileSize: 62000000000, AvailableEps: 9, MissingEps: 0, Episodes: getMockEpisodes(2, 1, 9, "available")},
		// 		{Number: 2, FileSize: 63000000000, AvailableEps: 1, MissingEps: 0, Episodes: getMockEpisodes(2, 2, 1, "available")},
		// 	},
		// },
		// {
		// 	Title:        "Shōgun",
		// 	YearStart:    2024,
		// 	YearEnd:      2024,
		// 	SeasonCount:  1,
		// 	EpisodeCount: 10,
		// 	Synopsis:     "Un navire anglais échoue au Japon au 17ème siècle.",
		// 	Genres:       "Drame, Histoire",
		// 	Rating:       8.5,
		// 	Popularity:   156.4,
		// 	Status:       "partial",
		// 	FileSize:     98000000000,
		// 	DateAdded:    now,
		// 	TVDBId:       407855,
		// 	IMDbId:       "tt13949816",
		// 	Cast:         []models.Cast{},
		// 	Seasons: []models.Season{
		// 		{Number: 1, FileSize: 98000000000, AvailableEps: 7, MissingEps: 3, Episodes: getShogonSeason1()},
		// 	},
		// },
		// {
		// 	Title:        "The Last of Us",
		// 	YearStart:    2023,
		// 	YearEnd:      2023,
		// 	SeasonCount:  1,
		// 	EpisodeCount: 9,
		// 	Synopsis:     "Un contrebandier cynique est engagé pour escorter une jeune fille à travers un monde infesté de champignons.",
		// 	Genres:       "Drame, Thriller",
		// 	Rating:       8.3,
		// 	Popularity:   187.3,
		// 	Status:       "complete",
		// 	FileSize:     110000000000,
		// 	DateAdded:    now,
		// 	TVDBId:       412410,
		// 	IMDbId:       "tt6468322",
		// 	Cast: []models.Cast{
		// 		{Name: "Pedro Pascal", Role: "Joel", Avatar: "PP"},
		// 		{Name: "Bella Ramsey", Role: "Ellie", Avatar: "BR"},
		// 	},
		// 	Seasons: []models.Season{
		// 		{Number: 1, FileSize: 110000000000, AvailableEps: 9, MissingEps: 0, Episodes: getMockEpisodes(4, 1, 9, "available")},
		// 	},
		// },
	}
}

func getBreakingBadSeason1() []models.Episode {
	episodes := []models.Episode{
		{SeasonNum: 1, EpisodeNum: 1, Title: "Pilot", Duration: 3480, Status: "available", FileSize: 3600000000, FilePath: "/media/series/Breaking Bad/S01/E01.mkv"},
		{SeasonNum: 1, EpisodeNum: 2, Title: "Cat's in the Bag", Duration: 2880, Status: "available", FileSize: 3100000000, FilePath: "/media/series/Breaking Bad/S01/E02.mkv"},
		{SeasonNum: 1, EpisodeNum: 3, Title: "And the Bag's in the River", Duration: 2880, Status: "available", FileSize: 3150000000, FilePath: "/media/series/Breaking Bad/S01/E03.mkv"},
		{SeasonNum: 1, EpisodeNum: 4, Title: "Cancer Man", Duration: 2880, Status: "available", FileSize: 3200000000, FilePath: "/media/series/Breaking Bad/S01/E04.mkv"},
		{SeasonNum: 1, EpisodeNum: 5, Title: "Gray Matter", Duration: 2880, Status: "available", FileSize: 3000000000, FilePath: "/media/series/Breaking Bad/S01/E05.mkv"},
		{SeasonNum: 1, EpisodeNum: 6, Title: "Crazy Handful of Nothin'", Duration: 2880, Status: "available", FileSize: 3350000000, FilePath: "/media/series/Breaking Bad/S01/E06.mkv"},
		{SeasonNum: 1, EpisodeNum: 7, Title: "A No-Rough-Stuff-Type Deal", Duration: 2880, Status: "available", FileSize: 3180000000, FilePath: "/media/series/Breaking Bad/S01/E07.mkv"},
	}
	for i := range episodes {
		episodes[i].MediaInfo = getMockEpisodeMediaInfo()
	}
	return episodes
}

func getBreakingBadSeason2() []models.Episode {
	episodes := []models.Episode{
		{SeasonNum: 2, EpisodeNum: 1, Title: "Seven Thirty-Seven", Duration: 2820, Status: "available", FileSize: 3100000000, FilePath: "/media/series/Breaking Bad/S02/E01.mkv"},
		{SeasonNum: 2, EpisodeNum: 2, Title: "Grilled", Duration: 2820, Status: "available", FileSize: 3250000000, FilePath: "/media/series/Breaking Bad/S02/E02.mkv"},
		{SeasonNum: 2, EpisodeNum: 3, Title: "Bit by a Dead Bee", Duration: 2820, Status: "available", FileSize: 3080000000, FilePath: "/media/series/Breaking Bad/S02/E03.mkv"},
		{SeasonNum: 2, EpisodeNum: 4, Title: "Down", Duration: 2820, Status: "missing", FileSize: 0, FilePath: ""},
		{SeasonNum: 2, EpisodeNum: 5, Title: "Breakage", Duration: 2820, Status: "available", FileSize: 3180000000, FilePath: "/media/series/Breaking Bad/S02/E05.mkv"},
	}
	for i, ep := range episodes {
		if ep.Status == "available" {
			episodes[i].MediaInfo = getMockEpisodeMediaInfo()
		}
	}
	return episodes
}

func getBreakingBadSeason3() []models.Episode {
	return []models.Episode{
		{SeasonNum: 3, EpisodeNum: 1, Title: "No Más", Duration: 2820, Status: "available", FileSize: 3200000000, FilePath: "/media/series/Breaking Bad/S03/E01.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 3, EpisodeNum: 2, Title: "Caballo sin Nombre", Duration: 2820, Status: "available", FileSize: 3120000000, FilePath: "/media/series/Breaking Bad/S03/E02.mkv", MediaInfo: getMockEpisodeMediaInfo()},
	}
}

func getBreakingBadSeason4() []models.Episode {
	return []models.Episode{
		{SeasonNum: 4, EpisodeNum: 1, Title: "Box Cutter", Duration: 2820, Status: "available", FileSize: 3350000000, FilePath: "/media/series/Breaking Bad/S04/E01.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 4, EpisodeNum: 2, Title: "Thirty-Eight Snub", Duration: 2820, Status: "available", FileSize: 3250000000, FilePath: "/media/series/Breaking Bad/S04/E02.mkv", MediaInfo: getMockEpisodeMediaInfo()},
	}
}

func getBreakingBadSeason5() []models.Episode {
	return []models.Episode{
		{SeasonNum: 5, EpisodeNum: 1, Title: "Live Free or Die", Duration: 2820, Status: "available", FileSize: 3400000000, FilePath: "/media/series/Breaking Bad/S05/E01.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 5, EpisodeNum: 2, Title: "Madrigal", Duration: 2820, Status: "available", FileSize: 3300000000, FilePath: "/media/series/Breaking Bad/S05/E02.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 5, EpisodeNum: 3, Title: "Hazard Pay", Duration: 2820, Status: "available", FileSize: 3220000000, FilePath: "/media/series/Breaking Bad/S05/E03.mkv", MediaInfo: getMockEpisodeMediaInfo()},
	}
}

func getShogonSeason1() []models.Episode {
	episodes := []models.Episode{
		{SeasonNum: 1, EpisodeNum: 1, Title: "Anjin", Duration: 3480, Status: "available", FileSize: 14000000000, FilePath: "/media/series/Shogun/S01/E01.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 2, Title: "Kagemusha", Duration: 3480, Status: "available", FileSize: 14100000000, FilePath: "/media/series/Shogun/S01/E02.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 3, Title: "Yaiba", Duration: 3480, Status: "available", FileSize: 13900000000, FilePath: "/media/series/Shogun/S01/E03.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 4, Title: "The Whirlwind of Time", Duration: 3480, Status: "available", FileSize: 14200000000, FilePath: "/media/series/Shogun/S01/E04.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 5, Title: "Broken to the Collar", Duration: 3480, Status: "available", FileSize: 14000000000, FilePath: "/media/series/Shogun/S01/E05.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 6, Title: "The Abyss of Longing", Duration: 3480, Status: "available", FileSize: 14150000000, FilePath: "/media/series/Shogun/S01/E06.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 7, Title: "Crimson Sky", Duration: 3480, Status: "available", FileSize: 14100000000, FilePath: "/media/series/Shogun/S01/E07.mkv", MediaInfo: getMockEpisodeMediaInfo()},
		{SeasonNum: 1, EpisodeNum: 8, Title: "Trades", Duration: 3480, Status: "missing", FileSize: 0, FilePath: ""},
		{SeasonNum: 1, EpisodeNum: 9, Title: "A Peaceful Day", Duration: 3480, Status: "missing", FileSize: 0, FilePath: ""},
		{SeasonNum: 1, EpisodeNum: 10, Title: "The Return of the Arrows", Duration: 3600, Status: "missing", FileSize: 0, FilePath: ""},
	}
	return episodes
}

func getMockEpisodes(seriesID int, season int, count int, status string) []models.Episode {
	episodes := make([]models.Episode, count)
	for i := 0; i < count; i++ {
		episodes[i] = models.Episode{
			SeriesID:   int64(seriesID),
			SeasonNum:  season,
			EpisodeNum: i + 1,
			Title:      fmt.Sprintf("Episode %d", i+1),
			Duration:   2700,
			Status:     status,
			FileSize:   3500000000,
			MediaInfo:  getMockEpisodeMediaInfo(),
		}
	}
	return episodes
}

func getMockEpisodeMediaInfo() *models.MediaInfo {
	return &models.MediaInfo{
		VideoTracks: []models.VideoTrack{
			{Codec: "H.265", Resolution: "3840x2160", FPS: 23.976, Bitrate: "28.5 Mbps", HDR: "Dolby Vision", ColorSpace: "BT.2020"},
		},
		AudioTracks: []models.AudioTrack{
			{Codec: "DTS-HD MA", Channels: "5.1", Language: "English", SampleRate: "48000 Hz", Bitrate: "2.3 Mbps"},
		},
		SubtitleTracks: []models.SubtitleTrack{
			{Language: "French", Format: "SRT"},
		},
	}
}
