import type { Movie } from '../types';

export const mockMovie: Movie = {
  id: 1,
  title: 'Test Movie',
  year: 2024,
  duration: 120,
  synopsis: 'A test movie',
  genres: 'Action, Drama',
  rating: 8.5,
  popularity: 100,
  status: 'available',
  fileSize: 5_000_000_000,
  filePath: '/movies/test.mkv',
  container: 'mkv',
  dateAdded: '2024-01-01',
  tmdbId: 12345,
  imdbId: 'tt1234567',
  cast: [],
  mediaInfo: {
    id: 1,
    videoTracks: [],
    audioTracks: [],
    subtitleTracks: [],
  },
};
