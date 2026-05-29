import { useMemo } from 'react';
import { Movie, Series } from '../types';
import { apiClient } from '../api/client';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { MovieCard } from '../components/MovieCard';
import { SeriesCard } from '../components/SeriesCard';
import styles from '../styles/recents.module.css';

interface RecentsProps {
  onSelectMovie: (id: number) => void;
  onSelectSeries: (id: number) => void;
}

export const Recents = ({ onSelectMovie, onSelectSeries }: RecentsProps) => {
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
        <span className={styles.title}>Ajoutés récemment</span>
      </div>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>Films</div>
        {moviesInitialLoading ? (
          <div className={styles.state}>Chargement des films récents...</div>
        ) : moviesError ? (
          <div className={styles.state}>Erreur: {moviesError}</div>
        ) : movies.length === 0 ? (
          <div className={styles.state}>Aucun film récemment ajouté.</div>
        ) : (
          <div className={styles.posterRow}>
            {movies.map((movie) => (
                <MovieCard movie={movie} onClick={() => onSelectMovie(movie.id)} />
            ))}
          </div>
        )}
        {moviesLoading && !moviesInitialLoading && <div className={styles.state}>Mise à jour des films...</div>}
      </section>

      <section className={styles.section}>
        <div className={styles.sectionHeader}>Séries</div>
        {seriesInitialLoading ? (
          <div className={styles.state}>Chargement des séries récentes...</div>
        ) : seriesError ? (
          <div className={styles.state}>Erreur: {seriesError}</div>
        ) : series.length === 0 ? (
          <div className={styles.state}>Aucune série récemment ajoutée.</div>
        ) : (
          <div className={styles.posterRow}>
            {series.map((item) => (
                <SeriesCard series={item} onClick={() => onSelectSeries(item.id)} />
            ))}
          </div>
        )}
        {seriesLoading && !seriesInitialLoading && <div className={styles.state}>Mise à jour des séries...</div>}
      </section>
    </div>
  );
};
