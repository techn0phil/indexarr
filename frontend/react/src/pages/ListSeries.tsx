import { useMemo, useState } from 'react';
import { Series } from '../types';
import { apiClient } from '../api/client';
import { useInfiniteList } from '../hooks/useInfiniteList';
import { useAppContext } from '../hooks/useAppContext';
import { SeriesCard } from '../components/SeriesCard';
import { SeriesCardList } from '../components/SeriesCardList';
import { StatCard } from '../components/StatCard';
import { FilterChip } from '../components/FilterChip';
import { FilterModal } from '../components/FilterModal';
import { ViewToggle } from '../components/ViewToggle';
import { ScanStatusCard } from '../components/ScanStatusCard';

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
  const context = useAppContext();

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
  const { items: series, loading, isInitialLoading, hasMore, loadMore, reset } = useInfiniteList<Series>({
    fetchFn: apiClient.getSeries,
    pageSize: 50,
    filters,
  });

  // Calculate stats from loaded series
  const loadedStats = useMemo(() => {
    const complete = series.filter((s) => s.status === 'complete').length;
    const episodes = series.reduce((sum, s) => sum + s.episodeCount, 0);
    const diskSpace = series.reduce((sum, s) => sum + (s.fileSize || 0), 0) / (1024 * 1024 * 1024);

    // -----------------------------------------------------------------------
    // -------------------------- To review !! -------------------------------
    // -----------------------------------------------------------------------
    // This is how to get count of missing episodes per season: const em = series[0].seasons[0].missingEps;
    // Count missing episodes based on seasons with missingEps > 0 for the whole series
    // const ms = series.reduce((sum, s) => sum + (s.seasons || []).reduce((seasonSum, season) => seasonSum + (season.missingEps > 0 ? season.episodes.length : 0), 0), 0);
    // const missingEpisodes = series.reduce((sum, s) => sum + (s.seasons || []).reduce((seasonSum, season) => seasonSum + season.missingEps, 0), 0);
    // TODO fix it later
    // For now we don't have missingEps data from API, so we'll just show the total from stats
    const missingEpisodes = context?.stats?.missingEpisodes || 0;
    // -----------------------------------------------------------------------
    // -----------------------------------------------------------------------
    // -----------------------------------------------------------------------

    return { complete, total: series.length, episodes, diskSpace, missingEpisodes };
  }, [series, context?.stats]);

  const handleViewChange = (newView: ViewType) => {
    setView(newView);
    localStorage.setItem('series-view', newView);
  };

  const handleScanComplete = () => {
    // Refresh series and stats after scan
    reset();
    context?.refreshStats();
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
            icon={<svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="6" cy="6" r="4.5"></circle></svg>}
            label="Statut"
            active={activeFilters.status.length > 0}
            count={activeFilters.status.length}
            onClick={() => setModalFilter('status')}
          />
        
          <FilterChip
            icon={<svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="1.5" y="2.5" width="9" height="7" rx="1"></rect></svg>}
            label="Résolution"
            active={activeFilters.resolution.length > 0}
            count={activeFilters.resolution.length}
            onClick={() => setModalFilter('resolution')}
          />
          
          <FilterChip
            icon={<svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M2 6h2l2-4 2 8 2-4 1 0"></path></svg>}
            label="Codec"
            active={activeFilters.codec.length > 0}
            count={activeFilters.codec.length}
            onClick={() => setModalFilter('codec')}
          />
          
          <FilterChip
            icon={<svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path></svg>}
            label="Audio"
            active={activeFilters.audio.length > 0}
            count={activeFilters.audio.length}
            onClick={() => setModalFilter('audio')}
          />
          
          <FilterChip
            icon={<svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="6" cy="6" r="2.5"/><line x1="6" y1="1" x2="6" y2="3"/><line x1="6" y1="9" x2="6" y2="11"/><line x1="1" y1="6" x2="3" y2="6"/><line x1="9" y1="6" x2="11" y2="6"/><line x1="3.5" y1="3.5" x2="2.2" y2="2.2"/><line x1="8.5" y1="3.5" x2="9.8" y2="2.2"/><line x1="3.5" y1="8.5" x2="2.2" y2="9.8"/><line x1="8.5" y1="8.5" x2="9.8" y2="9.8"/></svg>}
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
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(5, 1fr)', gap: '10px', padding: '0 20px', marginBottom: '16px' }}>
        <StatCard label="Séries" value={loadedStats.total} subLabels={[`${loadedStats.complete} / ${loadedStats.total} complètes`, `${context?.stats?.totalSeries || 0} total`]} />
        <StatCard label="Épisodes" value={loadedStats.episodes} subLabels={[`${loadedStats.episodes - loadedStats.missingEpisodes} / ${loadedStats.episodes} disponibles`, `${context?.stats?.totalEpisodes || 0} total`]} />
        <StatCard label="Espace" value={`${loadedStats.diskSpace.toFixed(1)} Go`} subLabels={['occupation disque', `${context?.stats?.diskSpaceGB?.toFixed(1) || 0} Go total`]} />
        <StatCard label="Problèmes" value={loadedStats.missingEpisodes || 0} subLabels={['épisodes manquants', `${context?.stats?.missingEpisodes || 0} total`]} />
        <ScanStatusCard onScanComplete={handleScanComplete} />
      </div>

      {/* Grid or List */}
      {isInitialLoading ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Chargement...
        </div>
      ) : !series || series.length === 0 ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-tertiary)' }}>
          Aucune série trouvée
        </div>
      ) : view === 'grid' ? (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(148px, 1fr))', gap: '12px', padding: '0 20px' }}>
            {series.map((s) => (
              <SeriesCard key={s.id} series={s} onClick={() => onSelectSeries(s.id)} />
            ))}
          </div>
          {hasMore && (
            <div style={{ display: 'flex', justifyContent: 'center', marginTop: '24px' }}>
              <button
                onClick={loadMore}
                disabled={loading}
                style={{
                  padding: '10px 24px',
                  fontSize: '13px',
                  fontWeight: 500,
                  color: loading ? 'var(--color-text-tertiary)' : 'var(--color-text-secondary)',
                  background: 'var(--color-background-secondary)',
                  border: '0.5px solid var(--color-border-tertiary)',
                  borderRadius: 'var(--border-radius-md)',
                  cursor: loading ? 'not-allowed' : 'pointer',
                  transition: 'all 0.2s ease',
                }}
              >
                {loading ? 'Chargement...' : 'Charger plus'}
              </button>
            </div>
          )}
        </>
      ) : (
        <>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', padding: '0 20px' }}>
            {series.map((s) => (
              <SeriesCardList key={s.id} series={s} onClick={() => onSelectSeries(s.id)} />
            ))}
          </div>
          {hasMore && (
            <div style={{ display: 'flex', justifyContent: 'center', marginTop: '24px' }}>
              <button
                onClick={loadMore}
                disabled={loading}
                style={{
                  padding: '10px 24px',
                  fontSize: '13px',
                  fontWeight: 500,
                  color: loading ? 'var(--color-text-tertiary)' : 'var(--color-text-secondary)',
                  background: 'var(--color-background-secondary)',
                  border: '0.5px solid var(--color-border-tertiary)',
                  borderRadius: 'var(--border-radius-md)',
                  cursor: loading ? 'not-allowed' : 'pointer',
                  transition: 'all 0.2s ease',
                }}
              >
                {loading ? 'Chargement...' : 'Charger plus'}
              </button>
            </div>
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
