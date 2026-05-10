import { useEffect, useState } from 'react';
import { Series } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';

interface SeriesDetailProps {
  seriesId: number;
}

export const SeriesDetail = ({ seriesId }: SeriesDetailProps) => {
  const [series, setSeries] = useState<Series | null>(null);
  const [currentSeason, setCurrentSeason] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const fetchSeries = async () => {
      setLoading(true);
      try {
        const data = await apiClient.getSeriesById(seriesId);
        setSeries(data);
      } catch (error) {
        console.error('Failed to fetch series:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSeries();
  }, [seriesId]);

  if (loading) return <div style={{ padding: '20px' }}>Chargement...</div>;
  if (!series) return <div style={{ padding: '20px' }}>Série non trouvée</div>;

  const season = series.seasons?.[currentSeason];

  return (
    <div>
      {/* Hero */}
      <div style={{ background: 'var(--color-background-primary)', borderBottom: '0.5px solid var(--color-border-tertiary)', padding: '24px' }}>
        <div style={{ display: 'flex', gap: '20px', alignItems: 'flex-start' }}>
          {/* Poster */}
          <div style={{ width: '110px', minWidth: '110px', height: '160px', background: 'var(--color-background-secondary)', borderRadius: '8px', border: '0.5px solid var(--color-border-tertiary)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: '6px', overflow: 'hidden' }}>
            {series.poster ? (
              <img
                src={series.poster}
                alt={series.title}
                style={{
                  width: '100%',
                  height: '100%',
                  objectFit: 'contain',
                  background: 'var(--color-background-secondary)',
                  display: 'block',
                  objectPosition: 'center',
                }}
              />
            ) : (
              <div style={{ fontSize: '26px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
                {series.title
                  .split(' ')
                  .map((w) => w[0])
                  .join('')}
              </div>
            )}
          </div>

          {/* Info */}
          <div style={{ flex: 1 }}>
            <h1 style={{ fontSize: '22px', fontWeight: 500, color: 'var(--color-text-primary)', marginBottom: '4px' }}>
              {series.title}
            </h1>
            <div style={{ fontSize: '13px', color: 'var(--color-text-tertiary)', marginBottom: '10px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
              <span>
                {series.yearStart}–{series.yearEnd}
              </span>
              <span>·</span>
              <span>
                {series.seasonCount} saisons · {series.episodeCount} épisodes
              </span>
              <span>·</span>
              <span>{series.genres}</span>
              <span>·</span>
              <span style={{ color: '#1D9E75', fontWeight: 500 }}>
                {series.status === 'complete' ? 'Complète' : 'Ongoing'}
              </span>
            </div>

            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '5px', marginBottom: '12px' }}>
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840') && (
                <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  4K UHD
                </span>
              )}
            </div>

            <p style={{ fontSize: '12px', color: 'var(--color-text-secondary)', lineHeight: 1.6, maxWidth: '560px' }}>
              {series.synopsis}
            </p>

            <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
              <button style={{ background: '#1D9E75', color: 'white', border: 'none', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>
                + Rechercher épisodes
              </button>
              <button style={{ background: 'var(--color-background-secondary)', color: 'var(--color-text-secondary)', border: '0.5px solid var(--color-border-tertiary)', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer' }}>
                Ouvrir dans Sonarr
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Season Tabs */}
      <div style={{ display: 'flex', gap: '4px', padding: '0 24px', background: 'var(--color-background-primary)', borderBottom: '0.5px solid var(--color-border-tertiary)', overflowX: 'auto' }}>
        {(series.seasons || []).map((s, idx) => (
          <button
            key={idx}
            onClick={() => setCurrentSeason(idx)}
            style={{
              padding: '10px 14px',
              fontSize: '12px',
              color: idx === currentSeason ? '#1D9E75' : 'var(--color-text-secondary)',
              borderBottom: idx === currentSeason ? '2px solid #1D9E75' : '2px solid transparent',
              cursor: 'pointer',
              background: 'none',
              border: 'none',
              whiteSpace: 'nowrap',
              fontWeight: idx === currentSeason ? 500 : 400,
              transition: 'all 0.15s',
            }}
          >
            Saison {s.number}
          </button>
        ))}
      </div>

      {/* Episodes */}
      {season && (
        <div style={{ padding: '16px 24px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
            <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', margin: 0 }}>
              Saison {season.number} — {season.episodes.length} épisodes
            </h2>
            <div style={{ display: 'flex', gap: '12px', fontSize: '11px', color: 'var(--color-text-tertiary)' }}>
              <span style={{ color: '#1D9E75' }}>{season.availableEps} disponibles</span>
              {season.missingEps > 0 && <span style={{ color: '#E24B4A' }}>{season.missingEps} manquant{season.missingEps > 1 ? 's' : ''}</span>}
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            {season.episodes.map((ep) => (
              <div
                key={ep.id}
                style={{
                  background: 'var(--color-background-primary)',
                  border: '0.5px solid var(--color-border-tertiary)',
                  borderRadius: '8px',
                  padding: '10px 14px',
                  display: 'flex',
                  alignItems: 'center',
                  gap: '12px',
                  cursor: 'pointer',
                  transition: 'all 0.15s',
                }}
              >
                <div style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', minWidth: '28px' }}>
                  E{String(ep.episodeNum).padStart(2, '0')}
                </div>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {ep.title}
                  </div>
                  <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', marginTop: '2px' }}>
                    {Math.round(ep.duration / 60)} min · {ep.mediaInfo?.videoTracks?.[0]?.codec || 'N/A'}
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                  {ep.status === 'missing' && <span className={comStyles['badge-missing']}>Manquant</span>}
                </div>
                <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', minWidth: '55px', textAlign: 'right' }}>
                  {ep.fileSize ? `${(ep.fileSize / 1024 / 1024 / 1024).toFixed(1)} Go` : '—'}
                </div>
                <div
                  style={{
                    width: '7px',
                    height: '7px',
                    borderRadius: '50%',
                    background: ep.status === 'available' ? '#1D9E75' : '#E24B4A',
                    flexShrink: 0,
                  }}
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
