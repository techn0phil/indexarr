import { useMemo, useState, useRef, useEffect } from 'react';
import { Movie } from '../types';
import { apiClient } from '../api/client';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { MovieCard } from '../components/MovieCard';
import { MovieCardList } from '../components/MovieCardList';
import { StatCard } from '../components/StatCard';
import { FilterChip } from '../components/FilterChip';
import { FilterModal } from '../components/FilterModal';
import { ViewToggle } from '../components/ViewToggle';
import { ScanStatusCard } from '../components/ScanStatusCard';

interface ListFilmsProps {
  onSelectMovie: (id: number) => void;
  searchQuery?: string;
}

type FilterType = 'status' | 'resolution' | 'codec' | 'audio' | 'hdr';
type ViewType = 'grid' | 'list';

const FILTER_OPTIONS: Record<FilterType, { value: string; label: string }[]> = {
  status: [
    { value: 'available', label: 'Disponible' },
    { value: 'missing', label: 'Manquant' },
    { value: 'problem', label: 'Problème' },
  ],
  resolution: [
    { value: '2160', label: '4K UHD (3840x2160)' },
    { value: '1080', label: 'Full HD (1920x1080)' },
    { value: '720', label: 'HD (1280x720)' },
    { value: '480', label: 'SD (720x480)' },
  ],
  codec: [
    { value: 'AV1', label: 'AV1' },
    { value: 'H.265', label: 'H.265 (HEVC)' },
    { value: 'H.264', label: 'H.264 (AVC)' },
    { value: 'MPEG-4', label: 'MPEG-4' },
  ],
  audio: [
    { value: 'TrueHD Atmos', label: 'Dolby TrueHD Atmos' },
    { value: 'TrueHD', label: 'Dolby TrueHD' },
    { value: 'E-AC-3 Atmos', label: 'Dolby Digital Plus Atmos' },
    { value: 'E-AC-3', label: 'Dolby Digital Plus' },
    { value: 'AC-3', label: 'Dolby Digital' },
    { value: 'DTS:X', label: 'DTS:X' },
    { value: 'DTS-HD MA', label: 'DTS-HD Master Audio' },
    { value: 'DTS', label: 'DTS' },
    { value: 'AAC', label: 'AAC' },
  ],
  hdr: [
    { value: 'Dolby Vision', label: 'Dolby Vision' },
    { value: 'HDR10+', label: 'HDR10+' },
    { value: 'HDR10', label: 'HDR10' },
  ],
};

export const ListFilms = ({ onSelectMovie, searchQuery = '' }: ListFilmsProps) => {
  const [activeFilters, setActiveFilters] = useState<Record<FilterType, string[]>>({
    status: [],
    resolution: [],
    codec: [],
    audio: [],
    hdr: [],
  });
  const [modalFilter, setModalFilter] = useState<FilterType | null>(null);
  const [view, setView] = useState<ViewType>(() => {
    const saved = localStorage.getItem('films-view');
    return (saved as ViewType) || 'grid';
  });

  // Build filter params for API
  const filters = useMemo(() => {
    const params: Record<string, string> = {};
    Object.entries(activeFilters).forEach(([key, values]) => {
      if (values.length > 0) {
        params[key] = values.join(',');
      }
    });
    if (searchQuery) {
      params.search = searchQuery;
    }
    return params;
  }, [activeFilters, searchQuery]);

  // Use infinite list hook for pagination
  const { items: movies, loading, hasMore, loadMore, reset } = useInfiniteList<Movie>({
    fetchFn: apiClient.getMovies,
    pageSize: 50,
    filters,
  });

  // Infinite scroll sentinel ref
  const sentinelRef = useRef<HTMLDivElement>(null);

  // Set up intersection observer for infinite scroll
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (entry.isIntersecting && hasMore && !loading) {
          loadMore();
        }
      },
      {
        root: null,
        rootMargin: '100px',
        threshold: 0.1,
      }
    );

    observer.observe(sentinel);

    return () => {
      observer.disconnect();
    };
  }, [hasMore, loading, loadMore]);

  // Calculate stats from loaded movies
  const stats = useMemo(() => {
    const available = movies.filter((m) => m.status === 'available').length;
    const diskSpace = movies.reduce((sum, m) => sum + (m.fileSize || 0), 0) / (1024 * 1024 * 1024);
    const fourK = movies.filter((m) => m.mediaInfo?.videoTracks?.[0]?.resolution.includes('x2160')).length;
    return { available, total: movies.length, diskSpace, fourK };
  }, [movies]);

  const handleViewChange = (newView: ViewType) => {
    setView(newView);
    localStorage.setItem('films-view', newView);
  };

  const handleScanComplete = () => {
    // Refresh movies and stats after scan
    reset();
  };

  const handleFilterApply = (filterType: FilterType, values: string[]) => {
    setActiveFilters((prev) => ({ ...prev, [filterType]: values }));
  };

  const getTotalActiveFilters = () => {
    return Object.values(activeFilters).reduce((sum, values) => sum + values.length, 0);
  };

  return (
    <div style={{ paddingBottom: '16px' }}>
      {/* Filters */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px', padding: '8px 20px', background: 'var(--color-background-primary)', borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
          <span style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', marginRight: '2px' }}>Filtres</span>
        
          <FilterChip
            label="Statut"
            active={activeFilters.status.length > 0}
            count={activeFilters.status.length}
          onClick={() => setModalFilter('status')}
        />
        
        <FilterChip
          label="Résolution"
          active={activeFilters.resolution.length > 0}
          count={activeFilters.resolution.length}
          onClick={() => setModalFilter('resolution')}
        />
        
        <FilterChip
          label="Codec"
          active={activeFilters.codec.length > 0}
          count={activeFilters.codec.length}
          onClick={() => setModalFilter('codec')}
        />
        
        <FilterChip
          label="Audio"
          active={activeFilters.audio.length > 0}
          count={activeFilters.audio.length}
          onClick={() => setModalFilter('audio')}
        />
        
        <FilterChip
          label="HDR"
          active={activeFilters.hdr.length > 0}
          count={activeFilters.hdr.length}
          onClick={() => setModalFilter('hdr')}
        />

        {getTotalActiveFilters() > 0 && (
          <button
            onClick={() => setActiveFilters({ status: [], resolution: [], codec: [], audio: [], hdr: [] })}
            style={{
              fontSize: '11px',
              color: 'var(--color-text-tertiary)',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              textDecoration: 'underline',
              marginLeft: '4px',
            }}
          >
            Effacer tout
          </button>
        )}
        </div>

        <ViewToggle view={view} onViewChange={handleViewChange} />
      </div>

      {/* Stats */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '10px', marginBottom: '16px', padding: '0 20px' }}>
        <StatCard label="Films" value={stats.total} subLabel={`${stats.available} disponibles`} />
        <StatCard label="Espace" value={`${stats.diskSpace.toFixed(1)} Go`} subLabel="moy. disque" />
        <StatCard label="4K UHD" value={stats.fourK} subLabel={`${stats.total > 0 ? Math.round((stats.fourK / stats.total) * 100) : 0}%`} />
        <StatCard label="Problèmes" value="0" subLabel="fichiers manquants" />
        <ScanStatusCard onScanComplete={handleScanComplete} />
      </div>

      {/* Grid or List */}
      {loading ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Chargement...
        </div>
      ) : !movies || movies.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Aucun film trouvé
        </div>
      ) : view === 'grid' ? (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(148px, 1fr))', gap: '12px', padding: '0 20px' }}>
            {movies.map((movie) => (
              <MovieCard key={movie.id} movie={movie} onClick={() => onSelectMovie(movie.id)} />
            ))}
          </div>
          {hasMore && (
            <>
              <div ref={sentinelRef} style={{ height: '1px' }} />
              {loading && (
                <div style={{ display: 'flex', justifyContent: 'center', marginTop: '24px', fontSize: '13px', color: 'var(--color-text-tertiary)' }}>
                  Chargement...
                </div>
              )}
            </>
          )}
        </>
      ) : (
        <>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', padding: '0 20px' }}>
            {movies.map((movie) => (
              <MovieCardList key={movie.id} movie={movie} onClick={() => onSelectMovie(movie.id)} />
            ))}
          </div>
          {hasMore && (
            <>
              <div ref={sentinelRef} style={{ height: '1px' }} />
              {loading && (
                <div style={{ display: 'flex', justifyContent: 'center', marginTop: '24px', fontSize: '13px', color: 'var(--color-text-tertiary)' }}>
                  Chargement...
                </div>
              )}
            </>
          )}
        </>
      )}

      {/* Filter Modals */}
      {modalFilter && (
        <FilterModal
          isOpen={true}
          onClose={() => setModalFilter(null)}
          onApply={(values) => handleFilterApply(modalFilter, values)}
          title={`Filtrer par ${modalFilter === 'status' ? 'statut' : modalFilter === 'resolution' ? 'résolution' : modalFilter === 'codec' ? 'codec' : modalFilter === 'audio' ? 'audio' : 'HDR'}`}
          options={FILTER_OPTIONS[modalFilter]}
          selectedValues={activeFilters[modalFilter]}
        />
      )}
    </div>
  );
};
