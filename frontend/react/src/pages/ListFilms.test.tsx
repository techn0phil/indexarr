import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { ListFilms } from './ListFilms';
import { apiClient } from '../api/client';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'fr', changeLanguage: vi.fn() },
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => ({
    stats: {
      totalMovies: 10,
      fourKCount: 2,
      fourKPercent: 20,
      moviesDiskSpaceGB: 100,
      missingMovies: 1,
    },
    refreshStats: vi.fn(),
  }),
}));

vi.mock('../components/MovieCard', () => ({
  MovieCard: ({ movie, onClick }: any) => (
    <button type="button" onClick={onClick}>
      {movie.title}
    </button>
  ),
}));

vi.mock('../components/MovieCardList', () => ({
  MovieCardList: ({ movie, onClick }: any) => (
    <button type="button" onClick={onClick}>
      list-{movie.title}
    </button>
  ),
}));

vi.mock('../components/StatCard', () => ({
  StatCard: ({ label, value }: any) => (
    <div>
      {label}:{String(value)}
    </div>
  ),
}));

vi.mock('../components/ScanStatusCard', () => ({
  ScanStatusCard: () => <div>scan-status-card</div>,
}));

describe('ListFilms', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();

    vi.spyOn(apiClient, 'getMovies').mockImplementation(async (_page, _pageSize, filters) => {
      const title = filters?.search === 'matrix' ? 'Matrix' : 'Movie A';
      return {
        success: true,
        data: [
          {
            id: 1,
            title,
            year: 2024,
            duration: 120,
            genres: 'Action',
            status: 'available',
            synopsis: 'test',
            fileSize: 1024,
            tmdbId: 1,
            mediaInfo: {
              videoTracks: [{ resolution: '3840x2160', hdr: 'HDR10', codec: 'H.265' }],
              audioTracks: [],
            },
          } as any,
        ],
        total: 1,
        page: 1,
        pageSize: 50,
      };
    });
  });

  it('fetches and renders movies on initial load', async () => {
    render(<ListFilms onSelectMovie={vi.fn()} />);

    expect(await screen.findByText('Movie A')).toBeInTheDocument();
    expect(apiClient.getMovies).toHaveBeenCalledWith(1, 50, {});
  });

  it('includes search query in API filters', async () => {
    render(<ListFilms onSelectMovie={vi.fn()} searchQuery="matrix" />);

    expect(await screen.findByText('Matrix')).toBeInTheDocument();
    expect(apiClient.getMovies).toHaveBeenCalledWith(1, 50, { search: 'matrix' });
  });

  it('applies resolution filter and composes API request params', async () => {
    const user = userEvent.setup();
    render(<ListFilms onSelectMovie={vi.fn()} />);

    await screen.findByText('Movie A');

    await user.click(screen.getByText('filter.resolution.label'));
    await user.click(screen.getByText('4K - Ultra HD (3840 x 2160)'));
    await user.click(screen.getByText('filter.button.apply'));

    await waitFor(() => {
      expect(apiClient.getMovies).toHaveBeenCalledWith(
        1,
        50,
        expect.objectContaining({ resolution: '3840' })
      );
    });
  });

  it('persists list view selection in localStorage', async () => {
    const user = userEvent.setup();
    render(<ListFilms onSelectMovie={vi.fn()} />);

    await screen.findByText('Movie A');
    await user.click(screen.getByRole('button', { name: 'Vue liste' }));

    expect(localStorage.getItem('films-view')).toBe('list');
  });
});
