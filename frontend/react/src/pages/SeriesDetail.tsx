import { useContext, useEffect, useState } from 'react';
import { Series } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';
import { AppContext } from '../hooks/useAppContext';

interface SeriesDetailProps {
  seriesId: number;
}

export const SeriesDetail = ({ seriesId }: SeriesDetailProps) => {
  const [series, setSeries] = useState<Series | null>(null);
  const [currentSeason, setCurrentSeason] = useState(0);
  const [loading, setLoading] = useState(false);
  const appContext = useContext(AppContext);

  // Slugify function to create URL-friendly strings
  const slugify = (text: string) =>
    text
      .toString()
      .normalize('NFD') // Normalize accented characters
      .replace(/[\u0300-\u036f]/g, '') // Remove accents
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, '-') // Replace non-alphanumeric with hyphens
      .replace(/^-+|-+$/g, ''); // Remove leading/trailing hyphens

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
            <h1 style={{ fontSize: '22px', fontWeight: 500, color: 'var(--color-text-primary)', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '10px' }}>
              {series.title}
            {typeof series.rating === 'number' && (
              <span style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '3px',
                background: 'var(--color-badge-rating)',
                color: 'var(--color-badge-rating-text)',
                borderRadius: '99px',
                fontSize: '12px',
                fontWeight: 500,
                padding: '2px 10px 2px 7px',
                border: 'none',
                lineHeight: 1,
                minWidth: '36px',
                height: '22px',
              }}>
                <svg width="11" height="11" viewBox="0 0 12 12" fill="var(--color-badge-rating-text)" style={{ marginRight: '2px', flexShrink: 0 }} aria-hidden="true"><path d="M6 1l1.4 3h3.1l-2.5 1.9 1 3L6 7.2l-3 1.7 1-3L1.5 4H4.6z"></path></svg>
                <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-badge-rating-text)', lineHeight: 1 }}>{series.rating?.toFixed(1)}</span>
              </span>
            )}
            </h1>
            <div style={{ fontSize: '13px', color: 'var(--color-text-tertiary)', marginBottom: '10px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
              <span>
                {series.yearStart}–{series.yearEnd}
              </span>
              <span>·</span>
              <span>
                {series.seasonCount} saisons
              </span>
              <span>·</span>
              <span>
                {series.episodeCount} épisodes
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
                  4K
                </span>
              )}
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('1080') && (
                <span className={comStyles['badge-1080p']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  1080p
                </span>
              )}
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
                <span className={comStyles['badge-dv']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  Dolby Vision
                </span>
              )}
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
                <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  HDR10+
                </span>
              )}
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && (
                <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  HDR10
                </span>
              )}
              {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'TrueHD') && (
                <span className={comStyles['badge-truehd']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  TrueHD
                </span>
              )}
              {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'E-AC-3') && (
                <span className={comStyles['badge-ddplus']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  Dolby Digital Plus
                </span>
              )}
              {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && (
                <span className={comStyles['badge-atmos']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  Atmos
                </span>
              )}
              {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && (
                <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  DTS
                </span>
              )}
              {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.codec && (
                <span className={comStyles['badge-codec']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                  {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.codec}
                </span>
              )}
            </div>

            <p style={{ fontSize: '12px', color: 'var(--color-text-secondary)', lineHeight: 1.6, maxWidth: '560px' }}>
              {series.synopsis}
            </p>

            {/* Actions */}
            <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
              {appContext?.config?.sonarrUrl && (
                <a href={`${appContext.config.sonarrUrl}/series/${slugify(series.title)}`} target="_blank" rel="noopener noreferrer" style={{ background: '#1D9E75', color: 'white', border: '0', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <img src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/sonarr-light.png" alt="Sonarr Light" style={{ width: '12px', height: '12px' }} />
                  Sonarr
                </a>
              )}
              <a href={`https://thetvdb.com/series/${slugify(series.title)}`} target="_blank" rel="noopener noreferrer" style={{ background: 'var(--color-background-secondary)', color: 'var(--color-text-secondary)', border: '0.5px solid var(--color-border-tertiary)', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <img src={appContext?.isDark ? 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tvdb-light.png' : 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tvdb-dark.png'} alt="TVDB Light" style={{ width: '12px', height: '12px' }} />
                TVDB
              </a>
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

                <div style={{ flex: '1', fontSize: '10px', color: 'var(--color-text-tertiary)', marginTop: '2px' }}>
                  {(ep.mediaInfo?.videoTracks ?? []).map((track, idx) => (
                    <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '7px', translate: '-3px 0' }}>
                      <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                        <rect x="2.5" y="5.5" width="11" height="7" rx="1.2" />
                        <path d="M2.5 5.5l1.5-3 2 3 1.5-3 2 3 1.5-3 2 3" />
                      </svg>
                      <span>{track.resolution || 'N/A'} · {track.fps || 'N/A'} fps · {track.bitrate || 'N/A'} · {track.colorSpace || 'N/A'}</span>
                    </div>
                  ))}
                  {(ep.mediaInfo?.audioTracks ?? []).map((track, idx) => (
                    <div key={idx} style={{ display: 'flex', alignItems: 'center', gap: '4px' }}>
                      <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                        <path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path>
                      </svg>
                      <span>{track.channels || 'N/A'} · {track.language || 'N/A'} · {track.bitrate || 'N/A'} · {track.sampleRate || 'N/A'}</span>
                    </div>
                  ))}
                </div>

                {/* Display badges: 4K, 1080p, Dolby Vision, HDR10+, HDR10, TrueHD, Dolby Digital Plus, Atmos, DTS, codec */}
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: '5px' }}>
                  {ep.mediaInfo?.videoTracks?.[0]?.resolution.includes('x2160') && (
                    <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      4K
                    </span>
                  )}
                  {ep.mediaInfo?.videoTracks?.[0]?.resolution.includes('x1080') && (
                    <span className={comStyles['badge-1080p']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      1080p
                    </span>
                  )}
                  {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
                    <span className={comStyles['badge-dv']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      Dolby Vision
                    </span>
                  )}
                  {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
                    <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      HDR10+
                    </span>
                  )}
                  {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && (
                    <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      HDR10
                    </span>
                  )}
                  {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'TrueHD') && (
                    <span className={comStyles['badge-truehd']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      TrueHD
                    </span>
                  )}
                  {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'E-AC-3') && (
                    <span className={comStyles['badge-ddplus']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      Dolby Digital Plus
                    </span>
                  )}
                  {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && (
                    <span className={comStyles['badge-atmos']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      Atmos
                    </span>
                  )}
                  {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && (
                    <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      DTS
                    </span>
                  )}
                  {ep.mediaInfo?.videoTracks?.[0]?.codec && (
                    <span className={comStyles['badge-codec']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                      {ep.mediaInfo.videoTracks?.[0]?.codec}
                    </span>
                  )}
                </div>

                <div style={{ display: 'flex', gap: '4px', flexShrink: 0 }}>
                  {ep.status === 'missing' && <span className={comStyles['badge-missing']}>Manquant</span>}
                </div>
                <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', minWidth: '55px', textAlign: 'right' }}>
                  {ep.fileSize ? (ep.fileSize < 1024 * 1024 * 1024 ? `${(ep.fileSize / 1024 / 1024).toFixed(1)} Mo` : `${(ep.fileSize / 1024 / 1024 / 1024).toFixed(1)} Go`) : '—'}
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
