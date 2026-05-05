import { useEffect, useState } from 'react';
import { Series, PaginatedResponse } from '../types';
import { apiClient } from '../api/client';
import { SeriesCard } from '../components/SeriesCard';
import { SeriesCardList } from '../components/SeriesCardList';
import { StatCard } from '../components/StatCard';
import { FilterChip } from '../components/FilterChip';
import { FilterModal } from '../components/FilterModal';
import { ViewToggle } from '../components/ViewToggle';

interface ListSeriesProps {
  onSelectSeries: (id: number) => void;
  searchQuery?: string;
}

type FilterType = 'status' | 'resolution' | 'codec' | 'audio' | 'hdr';
type ViewType = 'grid' | 'list';

const FILTER_OPTIONS: Record<FilterType, { value: string; label: string }[]> = {
  status: [
    { value: 'complete', label: 'Complète' },
    { value: 'ongoing', label: 'En cours' },
    { value: 'partial', label: 'Partielle' },
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

export const ListSeries = ({ onSelectSeries, searchQuery = '' }: ListSeriesProps) => {
  const [series, setSeries] = useState<Series[]>([]);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState({ complete: 0, total: 0, episodes: 0, diskSpace: 0 });
  const [activeFilters, setActiveFilters] = useState<Record<FilterType, string[]>>({
    status: [],
    resolution: [],
    codec: [],
    audio: [],
    hdr: [],
  });
  const [modalFilter, setModalFilter] = useState<FilterType | null>(null);
  const [view, setView] = useState<ViewType>(() => {
    const saved = localStorage.getItem('series-view');
    return (saved as ViewType) || 'grid';
  });

  const handleViewChange = (newView: ViewType) => {
    setView(newView);
    localStorage.setItem('series-view', newView);
  };

  useEffect(() => {
    const fetchSeries = async () => {
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

        const response = await apiClient.getSeries(1, 50, filters);
        const seriesData = response.data || [];
        setSeries(seriesData);
        
        // Calculate stats
        const complete = seriesData.filter((s) => s.status === 'complete').length;
        const episodes = seriesData.reduce((sum, s) => sum + s.episodeCount, 0);
        const diskSpace = seriesData.reduce((sum, s) => sum + (s.fileSize || 0), 0) / (1024 * 1024 * 1024 * 1024);
        setStats({ complete, total: seriesData.length, episodes, diskSpace });
      } catch (error) {
        console.error('Failed to fetch series:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSeries();
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
        <StatCard label="Séries" value={stats.total} subLabel={`${stats.complete} complètes`} />
        <StatCard label="Épisodes" value={stats.episodes} subLabel="total" />
        <StatCard label="Espace" value={`${stats.diskSpace.toFixed(1)} To`} subLabel="moy. par ep." />
        <StatCard label="Problèmes" value="0" subLabel="épisodes manquants" />
      </div>

      {/* Grid or List */}
      {loading ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Chargement...
        </div>
      ) : !series || series.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Aucune série trouvée
        </div>
      ) : view === 'grid' ? (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(148px, 1fr))', gap: '12px' }}>
          {series.map((s) => (
            <SeriesCard key={s.id} series={s} onClick={() => onSelectSeries(s.id)} />
          ))}
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {series.map((s) => (
            <SeriesCardList key={s.id} series={s} onClick={() => onSelectSeries(s.id)} />
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
