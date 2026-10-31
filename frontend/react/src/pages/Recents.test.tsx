import { beforeEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { render, screen } from '../test/testUtils';
import { Recents } from './Recents';
import { apiClient } from '../api/client';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { language: 'en', changeLanguage: vi.fn() },
  }),
}));

vi.mock('../components/MovieCard', () => ({
  MovieCard: ({ movie }: any) => <div>{movie.title}</div>,
}));

vi.mock('../components/SeriesCard', () => ({
  SeriesCard: ({ series }: any) => <div>{series.title}</div>,
}));

const page = (title: string) => ({
  success: true,
  data: [{ id: 1, title } as any],
  total: 1,
  page: 1,
  pageSize: 10,
});

describe('Recents', () => {
  beforeEach(() => {
    vi.spyOn(apiClient, 'getMovies').mockResolvedValue(page('Movie A'));
    vi.spyOn(apiClient, 'getSeries').mockResolvedValue(page('Series A'));
  });

  it('links each section to its list sorted by date added (descending)', async () => {
    render(
      <MemoryRouter>
        <Recents onSelectMovie={vi.fn()} onSelectSeries={vi.fn()} />
      </MemoryRouter>
    );

    await screen.findByText('Movie A');
    await screen.findByText('Series A');

    const links = screen.getAllByRole('link');
    const movieLinks = links.filter((l) => l.getAttribute('href') === '/movies?sort=-added');
    const seriesLinks = links.filter((l) => l.getAttribute('href') === '/series?sort=-added');

    // Header link + end-of-row tile for each section
    expect(movieLinks).toHaveLength(2);
    expect(seriesLinks).toHaveLength(2);
  });
});
