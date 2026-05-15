import { Movie, Series, PaginatedResponse, StatsResponse, ScanStatus, ScanResponse } from '../types/index';

const API_BASE = '/api';

export const apiClient = {
  getMovies: async (page: number = 1, pageSize: number = 50, filters: Record<string, string> = {}) => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
      ...filters,
    });
    const response = await fetch(`${API_BASE}/movies?${params}`);
    return response.json() as Promise<PaginatedResponse<Movie>>;
  },

  getMovie: async (id: number) => {
    const response = await fetch(`${API_BASE}/movies/${id}`);
    return response.json() as Promise<Movie>;
  },

  getSeries: async (page: number = 1, pageSize: number = 50, filters: Record<string, string> = {}) => {
    const params = new URLSearchParams({
      page: page.toString(),
      page_size: pageSize.toString(),
      ...filters,
    });
    const response = await fetch(`${API_BASE}/series?${params}`);
    return response.json() as Promise<PaginatedResponse<Series>>;
  },

  getSeriesById: async (id: number) => {
    const response = await fetch(`${API_BASE}/series/${id}`);
    return response.json() as Promise<Series>;
  },

  getStats: async () => {
    const response = await fetch(`${API_BASE}/stats`);
    return response.json() as Promise<StatsResponse>;
  },

  getConfig: async () => {
    const response = await fetch(`${API_BASE}/config`);
    return response.json() as Promise<{
      radarrUrl: string;
      sonarrUrl: string;
    }>;
  },

  // Scan endpoints
  triggerScan: async () => {
    const response = await fetch(`${API_BASE}/scan`, { method: 'POST' });
    return response.json() as Promise<ScanResponse>;
  },

  getScanStatus: async () => {
    const response = await fetch(`${API_BASE}/scan/status`);
    return response.json() as Promise<ScanStatus>;
  },

  stopScan: async () => {
    const response = await fetch(`${API_BASE}/scan/stop`, { method: 'POST' });
    return response.json() as Promise<ScanResponse>;
  },

  purgeDatabase: async () => {
    const response = await fetch(`${API_BASE}/purge`, { method: 'POST' });
    return response.json() as Promise<{ success: boolean; message?: string; error?: string }>;
  },

  refreshMovie: async (id: number) => {
    const response = await fetch(`${API_BASE}/movies/${id}/refresh`, { method: 'POST' });
    return response.json() as Promise<{ success: boolean; message?: string; result?: { filesFound: number } }>;
  },

  refreshSeries: async (id: number) => {
    const response = await fetch(`${API_BASE}/series/${id}/refresh`, { method: 'POST' });
    return response.json() as Promise<{ success: boolean; message?: string; result?: { filesFound: number } }>;
  },
};
