import { useMemo } from 'react';
import { Movie, Series } from '../types';
import { apiClient } from '../api/client';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { MovieCard } from '../components/MovieCard';
import { SeriesCard } from '../components/SeriesCard';
import styles from '../styles/recents.module.css';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

const MOVIES_LATEST_URL = '/movies?sort=-added';
const SERIES_LATEST_URL = '/series?sort=-added';

interface RecentsProps {
  onSelectMovie: (id: number) => void;
  onSelectSeries: (id: number) => void;
}

export const Recents = ({ onSelectMovie, onSelectSeries }: RecentsProps) => {
  const { t } = useTranslation('latest');
  const movieFilters = useMemo(() => ({ sort: '-added' }), []);
  const seriesFilters = useMemo(() => ({ sort: '-added' }), []);

  const {
    items: movies,
    loading: moviesLoading,
    isInitialLoading: moviesInitialLoading,
    error: moviesError,
  } = useInfiniteList<Movie>({
    fetchFn: apiClient.getMovies,
    pageSize: 10,
    filters: movieFilters,
  });

  const {
    items: series,
    loading: seriesLoading,
    isInitialLoading: seriesInitialLoading,
    error: seriesError,
  } = useInfiniteList<Series>({
    fetchFn: apiClient.getSeries,
    pageSize: 10,
    filters: seriesFilters,
  });

  return (
    <div className={styles.page}>
      <div className={styles.topbar}>
        <span className={styles.title}>{t('label.title')}</span>
      </div>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <span>{t('label.section.movies')}</span>
          <Link to={MOVIES_LATEST_URL} className={styles.seeAll}>
            {t('label.seeAll')} →
          </Link>
        </div>
        {moviesInitialLoading ? (
          <div className={styles.state}>{t('message.moviesInitialLoading')}</div>
        ) : moviesError ? (
          <div className={styles.state}>Error: {moviesError}</div>
        ) : movies.length === 0 ? (
          <div className={styles.state}>{t('message.noMovies')}</div>
        ) : (
          <div className={styles.posterRow}>
            {movies.map((movie) => (
              <MovieCard key={movie.id} movie={movie} onClick={() => onSelectMovie(movie.id)} />
            ))}
            <Link to={MOVIES_LATEST_URL} className={styles.seeAllTile}>
              <svg viewBox="0 0 24 24" width="96" height="96">
                <path d="M18 3v2h-2V3H8v2H6V3H4v18h2v-2h2v2h8v-2h2v2h2V3h-2zM8 17H6v-2h2v2zm0-4H6v-2h2v2zm0-4H6V7h2v2zm10 8h-2v-2h2v2zm0-4h-2v-2h2v2zm0-4h-2V7h2v2z" />
              </svg>
              <span className={styles.seeAllArrow}>→</span>
              <span>{t('label.seeAll')}</span>
            </Link>
          </div>
        )}
        {moviesLoading && !moviesInitialLoading && <div className={styles.state}>{t('message.moviesLoading')}</div>}
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <span>{t('label.section.series')}</span>
          <Link to={SERIES_LATEST_URL} className={styles.seeAll}>
            {t('label.seeAll')} →
          </Link>
        </div>
        {seriesInitialLoading ? (
          <div className={styles.state}>{t('message.seriesInitialLoading')}</div>
        ) : seriesError ? (
          <div className={styles.state}>Error: {seriesError}</div>
        ) : series.length === 0 ? (
          <div className={styles.state}>{t('message.noSeries')}</div>
        ) : (
          <div className={styles.posterRow}>
            {series.map((item) => (
              <SeriesCard key={item.id} series={item} onClick={() => onSelectSeries(item.id)} />
            ))}
            <Link to={SERIES_LATEST_URL} className={styles.seeAllTile}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="96" height="96">
                <path d="M21 3H3c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h5v2h8v-2h5c1.1 0 1.9-.9 1.9-2L23 5c0-1.1-.9-2-2-2zm0 14H3V5h18v12z"/>
              </svg>
              <span className={styles.seeAllArrow}>→</span>
              <span>{t('label.seeAll')}</span>
            </Link>
          </div>
        )}
        {seriesLoading && !seriesInitialLoading && <div className={styles.state}>{t('message.seriesLoading')}</div>}
      </section>
    </div>
  );
};
