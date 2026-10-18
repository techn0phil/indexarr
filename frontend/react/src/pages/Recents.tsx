import { useMemo } from 'react';
import { Movie, Series } from '../types';
import { apiClient } from '../api/client';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { MovieCard } from '../components/MovieCard';
import { SeriesCard } from '../components/SeriesCard';
import styles from '../styles/recents.module.css';
import { useTranslation } from 'react-i18next';

interface RecentsProps {
  onSelectMovie: (id: number) => void;
  onSelectSeries: (id: number) => void;
}

export const Recents = ({ onSelectMovie, onSelectSeries }: RecentsProps) => {
  const { t } = useTranslation('latest');
  const movieFilters = useMemo(() => ({ sort: 'added' }), []);
  const seriesFilters = useMemo(() => ({ sort: 'added' }), []);

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
        <div className={styles.sectionHeader}>{t('label.section.movies')}</div>
        {moviesInitialLoading ? (
          <div className={styles.state}>{t('message.moviesInitialLoading')}</div>
        ) : moviesError ? (
          <div className={styles.state}>Error: {moviesError}</div>
        ) : movies.length === 0 ? (
          <div className={styles.state}>{t('message.noMovies')}</div>
        ) : (
          <div className={styles.posterRow}>
            {movies.map((movie) => (
                <MovieCard movie={movie} onClick={() => onSelectMovie(movie.id)} />
            ))}
          </div>
        )}
        {moviesLoading && !moviesInitialLoading && <div className={styles.state}>{t('message.moviesLoading')}</div>}
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>{t('label.section.series')}</div>
        {seriesInitialLoading ? (
          <div className={styles.state}>{t('message.seriesInitialLoading')}</div>
        ) : seriesError ? (
          <div className={styles.state}>Error: {seriesError}</div>
        ) : series.length === 0 ? (
          <div className={styles.state}>{t('message.noSeries')}</div>
        ) : (
          <div className={styles.posterRow}>
            {series.map((item) => (
                <SeriesCard series={item} onClick={() => onSelectSeries(item.id)} />
            ))}
          </div>
        )}
        {seriesLoading && !seriesInitialLoading && <div className={styles.state}>{t('message.seriesLoading')}</div>}
      </section>
    </div>
  );
};
