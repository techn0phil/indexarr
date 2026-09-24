import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { MovieDetail } from './MovieDetail';
import { apiClient } from '../api/client';

const mockContext = {
  authMode: 'none' as 'none' | 'simple' | 'oidc',
  user: { id: 1, username: 'admin', role: 'admin' as 'admin' | 'guest' },
  config: { radarrUrl: 'http://radarr.local' },
  isDark: false,
};

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => mockContext,
}));

describe('MovieDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and renders movie details', async () => {
    vi.spyOn(apiClient, 'getMovie').mockResolvedValue({
      id: 7,
      title: 'Movie X',
      year: 2024,
      duration: 120,
      genres: 'Action',
      status: 'available',
      synopsis: 'Synopsis',
      tmdbId: 77,
      fileSize: 1024,
      filePath: '/media/movie-x.mkv',
      mediaInfo: {
        videoTracks: [{ resolution: '3840x2160', hdr: 'HDR10', codec: 'H.265' }],
        audioTracks: [{ codec: 'E-AC-3 Atmos', channels: 6, language: 'en' }],
        subtitleTracks: [],
      },
      cast: [],
    } as any);

    render(<MovieDetail movieId={7} />);

    const titles = await screen.findAllByText('Movie X');
    expect(titles.length).toBeGreaterThan(0);
    expect(apiClient.getMovie).toHaveBeenCalledWith(7);
  });

  it('shows not-found fallback when fetch fails', async () => {
    vi.spyOn(apiClient, 'getMovie').mockRejectedValue(new Error('boom'));

    render(<MovieDetail movieId={404} />);

    expect(await screen.findByText('message.movieNotFound')).toBeInTheDocument();
  });

  it('hides refresh menu for non-admin users when auth is simple', async () => {
    mockContext.authMode = 'simple';
    mockContext.user = { id: 2, username: 'guest', role: 'guest' };

    vi.spyOn(apiClient, 'getMovie').mockResolvedValue({
      id: 8,
      title: 'Guest Movie',
      year: 2022,
      duration: 100,
      genres: 'Drama',
      status: 'available',
      synopsis: 'Synopsis',
      tmdbId: 88,
      fileSize: 2048,
      filePath: '/media/guest.mkv',
      mediaInfo: {
        videoTracks: [{ resolution: '1920x1080', hdr: '', codec: 'H.264' }],
        audioTracks: [{ codec: 'AAC', channels: 2, language: 'en' }],
        subtitleTracks: [],
      },
      cast: [],
    } as any);

    render(<MovieDetail movieId={8} />);

    await screen.findByText('status.available');
    expect(screen.queryByRole('button', { name: 'Menu' })).not.toBeInTheDocument();

    mockContext.authMode = 'none';
    mockContext.user = { id: 1, username: 'admin', role: 'admin' };
  });

  it('refreshes movie when admin triggers refresh action', async () => {
    const user = userEvent.setup();

    vi.spyOn(apiClient, 'getMovie').mockResolvedValue({
      id: 9,
      title: 'Refreshable Movie',
      year: 2021,
      duration: 110,
      genres: 'Sci-Fi',
      status: 'available',
      synopsis: 'Synopsis',
      tmdbId: 99,
      fileSize: 4096,
      filePath: '/media/refresh.mkv',
      mediaInfo: {
        videoTracks: [{ resolution: '1920x1080', hdr: '', codec: 'H.265' }],
        audioTracks: [{ codec: 'AAC', channels: 2, language: 'en' }],
        subtitleTracks: [],
      },
      cast: [],
    } as any);
    vi.spyOn(apiClient, 'refreshMovie').mockResolvedValue({ success: true, result: { filesFound: 1 } });

    render(<MovieDetail movieId={9} />);

    await screen.findByText('status.available');
    await user.click(screen.getByRole('button', { name: 'Menu' }));
    await user.click(screen.getByText('button.refresh'));

    await waitFor(() => {
      expect(apiClient.refreshMovie).toHaveBeenCalledWith(9);
      expect(apiClient.getMovie).toHaveBeenCalledTimes(2);
    });
  });
});
