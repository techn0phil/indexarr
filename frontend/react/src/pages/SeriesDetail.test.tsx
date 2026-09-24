import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { SeriesDetail } from './SeriesDetail';
import { apiClient } from '../api/client';

const mockContext = {
  authMode: 'none' as 'none' | 'simple' | 'oidc',
  user: { id: 1, username: 'admin', role: 'admin' as 'admin' | 'guest' },
  config: { sonarrUrl: 'http://sonarr.local' },
  isDark: false,
};

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: Record<string, unknown>) => {
      if (options && typeof options.number === 'number') {
        return `${key} ${options.number}`;
      }
      return key;
    },
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => mockContext,
}));

const seriesPayload = {
  id: 10,
  title: 'Series X',
  yearStart: 2020,
  yearEnd: 2024,
  seasonCount: 2,
  episodeCount: 3,
  totalEpisodeCount: 4,
  genres: 'Drama',
  status: 'ongoing',
  synopsis: 'Series synopsis',
  slug: 'series-x',
  seasons: [
    {
      number: 1,
      availableEps: 1,
      missingEps: 1,
      episodes: [
        {
          id: 101,
          episodeNum: 1,
          title: 'Pilot',
          duration: 3000,
          status: 'available',
          mediaInfo: {
            videoTracks: [{ resolution: '1920x1080', hdr: '', codec: 'H.264', bitrate: 1000, frameRate: 24, colorSpace: 'BT.709' }],
            audioTracks: [{ codec: 'AAC', channels: 2, sampleRate: 48000, bitrate: 256, language: 'en' }],
            subtitleTracks: [],
          },
        },
      ],
    },
    {
      number: 2,
      availableEps: 1,
      missingEps: 0,
      episodes: [
        {
          id: 201,
          episodeNum: 1,
          title: 'Season Two Premiere',
          duration: 3200,
          status: 'available',
          mediaInfo: {
            videoTracks: [{ resolution: '3840x2160', hdr: 'HDR10', codec: 'H.265', bitrate: 2000, frameRate: 24, colorSpace: 'BT.2020' }],
            audioTracks: [{ codec: 'E-AC-3 Atmos', channels: 6, sampleRate: 48000, bitrate: 640, language: 'en' }],
            subtitleTracks: [],
          },
        },
      ],
    },
  ],
};

describe('SeriesDetail', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads and renders series details', async () => {
    vi.spyOn(apiClient, 'getSeriesById').mockResolvedValue(seriesPayload as any);

    render(<SeriesDetail seriesId={10} />);

    const titles = await screen.findAllByText('Series X');
    expect(titles.length).toBeGreaterThan(0);
    expect(screen.getByText('Pilot')).toBeInTheDocument();
  });

  it('switches seasons and shows selected season episodes', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'getSeriesById').mockResolvedValue(seriesPayload as any);

    render(<SeriesDetail seriesId={10} />);

    await screen.findByText('Pilot');
    await user.click(screen.getByRole('button', { name: 'section.season 2' }));

    expect(screen.getByText('Season Two Premiere')).toBeInTheDocument();
  });

  it('hides refresh menu for guest users in simple auth mode', async () => {
    mockContext.authMode = 'simple';
    mockContext.user = { id: 2, username: 'guest', role: 'guest' };
    vi.spyOn(apiClient, 'getSeriesById').mockResolvedValue(seriesPayload as any);

    render(<SeriesDetail seriesId={10} />);

    await screen.findByText('Pilot');
    expect(screen.queryByRole('button', { name: 'Menu' })).not.toBeInTheDocument();

    mockContext.authMode = 'none';
    mockContext.user = { id: 1, username: 'admin', role: 'admin' };
  });

  it('refreshes series when admin triggers refresh action', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'getSeriesById').mockResolvedValue(seriesPayload as any);
    vi.spyOn(apiClient, 'refreshSeries').mockResolvedValue({ success: true, result: { filesFound: 1 } });

    render(<SeriesDetail seriesId={10} />);

    await screen.findByText('Pilot');
    await user.click(screen.getByRole('button', { name: 'Menu' }));
    await user.click(screen.getByText('button.refresh'));

    await waitFor(() => {
      expect(apiClient.refreshSeries).toHaveBeenCalledWith(10);
      expect(apiClient.getSeriesById).toHaveBeenCalledTimes(2);
    });
  });
});
