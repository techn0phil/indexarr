import { useEffect, useState } from 'react';
import { Movie, PaginatedResponse } from '../types';
import { apiClient } from '../api/client';
import { MovieCard } from '../components/MovieCard';
import { MovieCardList } from '../components/MovieCardList';
import { StatCard } from '../components/StatCard';
import { FilterChip } from '../components/FilterChip';
import { FilterModal } from '../components/FilterModal';
import { ViewToggle } from '../components/ViewToggle';
import comStyles from '../styles/components.module.css';

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
    { value: '3840', label: '4K UHD (3840x2160)' },
    { value: '1920', label: '1080p (1920x1080)' },
    { value: '1280', label: '720p (1280x720)' },
  ],
  codec: [
    { value: 'H.265', label: 'H.265 (HEVC)' },
    { value: 'H.264', label: 'H.264 (AVC)' },
    { value: 'AV1', label: 'AV1' },
  ],
  audio: [
    { value: 'TrueHD Atmos', label: 'TrueHD Atmos' },
    { value: 'DTS-HD MA', label: 'DTS-HD MA' },
    { value: 'AAC', label: 'AAC' },
  ],
  hdr: [
    { value: 'Dolby Vision', label: 'Dolby Vision' },
    { value: 'HDR10+', label: 'HDR10+' },
    { value: 'HDR10', label: 'HDR10' },
  ],
};

export const ListFilms = ({ onSelectMovie, searchQuery = '' }: ListFilmsProps) => {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({ available: 0, total: 0, diskSpace: 0, fourK: 0 });
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

  const handleViewChange = (newView: ViewType) => {
    setView(newView);
    localStorage.setItem('films-view', newView);
  };

  useEffect(() => {
    const fetchMovies = async () => {
      setLoading(true);
      try {
        // Build filter params
        const filters: Record<string, string> = {};
        Object.entries(activeFilters).forEach(([key, values]) => {
          if (values.length > 0) {
            filters[key] = values.join(',');
          }
        });

        if (searchQuery) {
          filters.search = searchQuery;
        }

        const response = await apiClient.getMovies(1, 50, filters);
        const moviesData = response.data || [];
        setMovies(moviesData);
        
        // Calculate stats
        const available = moviesData.filter((m) => m.status === 'available').length;
        const diskSpace = moviesData.reduce((sum, m) => sum + (m.fileSize || 0), 0) / (1024 * 1024 * 1024);
        const fourK = moviesData.filter((m) => m.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840')).length;
        setStats({ available, total: moviesData.length, diskSpace, fourK });
      } catch (error) {
        console.error('Failed to fetch movies:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchMovies();
  }, [activeFilters, searchQuery]);

  const handleFilterApply = (filterType: FilterType, values: string[]) => {
    setActiveFilters((prev) => ({ ...prev, [filterType]: values }));
  };

  const getTotalActiveFilters = () => {
    return Object.values(activeFilters).reduce((sum, values) => sum + values.length, 0);
  };

  return (
    <div style={{ padding: '16px 20px' }}>
      {/* Filters */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '16px', paddingBottom: '8px', borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
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
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '10px', marginBottom: '16px' }}>
        <StatCard label="Films" value={stats.total} subLabel={`${stats.available} disponibles`} />
        <StatCard label="Espace" value={`${stats.diskSpace.toFixed(1)} Go`} subLabel="moy. disque" />
        <StatCard label="4K UHD" value={stats.fourK} subLabel={`${stats.total > 0 ? Math.round((stats.fourK / stats.total) * 100) : 0}%`} />
        <StatCard label="Problèmes" value="0" subLabel="fichiers manquants" />
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
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(148px, 1fr))', gap: '12px' }}>
          {movies.map((movie) => (
            <MovieCard key={movie.id} movie={movie} onClick={() => onSelectMovie(movie.id)} />
          ))}
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {movies.map((movie) => (
            <MovieCardList key={movie.id} movie={movie} onClick={() => onSelectMovie(movie.id)} />
          ))}
        </div>
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
