import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { ListSeries } from './ListSeries';
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
      totalSeries: 12,
      totalEpisodes: 120,
      seriesDiskSpaceGB: 200,
      missingEpisodes: 3,
    },
    refreshStats: vi.fn(),
  }),
}));

vi.mock('../components/SeriesCard', () => ({
  SeriesCard: ({ series, onClick }: any) => (
    <button type="button" onClick={onClick}>
      {series.title}
    </button>
  ),
}));

vi.mock('../components/SeriesCardList', () => ({
  SeriesCardList: ({ series, onClick }: any) => (
    <button type="button" onClick={onClick}>
      list-{series.title}
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

describe('ListSeries', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();

    vi.spyOn(apiClient, 'getSeries').mockImplementation(async (_page, _pageSize, filters) => {
      const title = filters.search === 'dark' ? 'Dark' : 'Series A';
      return {
        success: true,
        data: [
          {
            id: 1,
            title,
            yearStart: 2020,
            yearEnd: 2024,
            episodeCount: 10,
            totalEpisodeCount: 10,
            seasonCount: 1,
            genres: 'Drama',
            status: 'complete',
            fileSize: 2048,
          } as any,
        ],
        total: 1,
        page: 1,
        pageSize: 50,
      };
    });
  });

  it('fetches and renders series on initial load', async () => {
    render(<ListSeries onSelectSeries={vi.fn()} />);

    expect(await screen.findByText('Series A')).toBeInTheDocument();
    expect(apiClient.getSeries).toHaveBeenCalledWith(1, 50, {});
  });

  it('includes search query in API filters', async () => {
    render(<ListSeries onSelectSeries={vi.fn()} searchQuery="dark" />);

    expect(await screen.findByText('Dark')).toBeInTheDocument();
    expect(apiClient.getSeries).toHaveBeenCalledWith(1, 50, { search: 'dark' });
  });

  it('applies status filter and composes API request params', async () => {
    const user = userEvent.setup();
    render(<ListSeries onSelectSeries={vi.fn()} />);

    await screen.findByText('Series A');

    await user.click(screen.getByText('filter.status.label'));
    await user.click(screen.getByText('filter.status.option.complete'));
    await user.click(screen.getByText('filter.button.apply'));

    await waitFor(() => {
      expect(apiClient.getSeries).toHaveBeenCalledWith(
        1,
        50,
        expect.objectContaining({ status: 'complete' })
      );
    });
  });

  it('persists list view selection in localStorage', async () => {
    const user = userEvent.setup();
    render(<ListSeries onSelectSeries={vi.fn()} />);

    await screen.findByText('Series A');
    await user.click(screen.getByRole('button', { name: 'Vue liste' }));

    expect(localStorage.getItem('series-view')).toBe('list');
  });
});
