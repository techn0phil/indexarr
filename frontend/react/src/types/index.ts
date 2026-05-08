export interface Movie {
  id: number;
  title: string;
  year: number;
  duration: number; // minutes
  synopsis: string;
  genres: string; // comma-separated
  rating: number;
  popularity: number;
  status: 'available' | 'missing' | 'problem';
  fileSize: number;
  filePath: string;
  container: string;
  dateAdded: string;
  tmdbId: number;
  imdbId: string;
  poster?: string;
  cast: Cast[];
  mediaInfo?: MediaInfo;
}

export interface Series {
  id: number;
  title: string;
  yearStart: number;
  yearEnd: number;
  seasonCount: number;
  episodeCount: number;
  synopsis: string;
  genres: string;
  rating: number;
  popularity: number;
  status: 'complete' | 'ongoing' | 'partial';
  fileSize: number;
  dateAdded: string;
  tvdbId: number;
  imdbId: string;
  poster?: string;
  cast: Cast[];
  seasons: Season[];
}

export interface Season {
  id: number;
  seriesId: number;
  number: number;
  episodes: Episode[];
  fileSize: number;
  availableEps: number;
  missingEps: number;
}

export interface Episode {
  id: number;
  seriesId: number;
  seasonNum: number;
  episodeNum: number;
  title: string;
  duration: number; // seconds
  status: 'available' | 'missing';
  fileSize: number;
  filePath: string;
  dateAdded: string;
  mediaInfo?: MediaInfo;
}

export interface Cast {
  id: number;
  name: string;
  role: string;
  avatar: string;
}

export interface MediaInfo {
  id: number;
  videoTracks: VideoTrack[];
  audioTracks: AudioTrack[];
  subtitleTracks: SubtitleTrack[];
}

export interface VideoTrack {
  codec: string;
  resolution: string;
  fps: number;
  bitrate: string;
  hdr: string;
  colorSpace: string;
}

export interface AudioTrack {
  codec: string;
  channels: string;
  language: string;
  sampleRate: string;
  bitrate: string;
}

export interface SubtitleTrack {
  language: string;
  format: string;
  forced: boolean;
  default: boolean;
}

export interface FilterCriteria {
  status?: string;
  resolution?: string;
  codec?: string;
  audio?: string;
  hdr?: string;
  sort?: string;
  page?: number;
  pageSize?: number;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  error?: string;
}

export interface StatsResponse {
  success: boolean;
  totalMovies: number;
  totalSeries: number;
  totalEpisodes: number;
  diskSpaceGB: number;
  fourKCount: number;
  fourKPercent: number;
  problemsCount: number;
  availMovies: number;
  missingMovies: number;
  availEpisodes: number;
  missingEpisodes: number;
  error?: string;
}

export interface ScanStatus {
  id: number;
  status: 'idle' | 'running' | 'completed' | 'error' | 'stopped';
  startedAt?: string;
  completedAt?: string;
  filesFound: number;
  filesProcessed: number;
  errorMessage?: string;
}

export interface ScanResponse {
  success: boolean;
  message: string;
}
