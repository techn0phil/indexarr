import { useEffect, useState } from 'react';
import { Series } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';
import { useAppContext } from '../hooks/useAppContext';
import { useTranslation } from 'react-i18next';

interface SeriesDetailProps {
  seriesId: number;
}

export const SeriesDetail = ({ seriesId }: SeriesDetailProps) => {
  const { t } = useTranslation('series-details');
  const [series, setSeries] = useState<Series | null>(null);
  const [currentSeason, setCurrentSeason] = useState(0);
  const [loading, setLoading] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [expandedEpisodeId, setExpandedEpisodeId] = useState<number | null>(null);
  const { authMode, config, isDark, user } = useAppContext();

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

  // Toggle episode expansion
  const toggleEpisode = (episodeId: number) => {
    setExpandedEpisodeId(expandedEpisodeId === episodeId ? null : episodeId);
  };

  // Refresh handler for menu
  const handleRefresh = async () => {
    const response = await apiClient.refreshSeries(seriesId);
        
    if (response.result?.filesFound) {
      fetchSeries();
    }
    else {
      window.location.reload();
    }
  };

  useEffect(() => {
    fetchSeries();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [seriesId]);

  if (loading) return <div style={{ padding: '20px' }}>{t('message.loading')}</div>;
  if (!series) return <div style={{ padding: '20px' }}>{t('message.seriesNotFound')}</div>;

  const season = series.seasons?.[currentSeason];
  const statusColor = series.status === 'complete' ? '#1D9E75' : series.status === 'ongoing' ? '#EF9F27' : '#E24B4A';

  return (
    <div>
      {/* Hero */}
      <div style={{ background: 'var(--color-background-primary)', borderBottom: '0.5px solid var(--color-border-tertiary)', padding: '24px' }}>
        <div className={comStyles['series-hero']}>
          {/* Poster */}
          <div className={comStyles['series-hero-poster']}>
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
              <>
                <div style={{ fontSize: '30px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
                  {series.title
                    .split(' ')
                    .map((w) => w[0])
                    .join('')}
                </div>
                <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', opacity: 0.4, textAlign: 'center', maxWidth: '90%' }}>
                  {series.title}
                </div>
                <div style={{ fontSize: '14px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
                  {`${series.yearStart || '?'} - ${series.yearEnd || '?'}`}
                </div>
              </>
            )}
          </div>

          {/* Info */}
          <div style={{ flex: 1 }}>
            <h1 className={comStyles['series-hero-title']}>
              {series.title}
              {typeof series.rating === 'number' && (
                <span className={comStyles['series-hero-rating']}>
                  <svg width="11" height="11" viewBox="0 0 12 12" fill="var(--color-badge-rating-text)" style={{ marginRight: '2px', flexShrink: 0 }} aria-hidden="true"><path d="M6 1l1.4 3h3.1l-2.5 1.9 1 3L6 7.2l-3 1.7 1-3L1.5 4H4.6z"></path></svg>
                  <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-badge-rating-text)', lineHeight: 1 }}>{series.rating?.toFixed(1)}</span>
                </span>
              )}

              {/* Popup contextual menu */}
              {(authMode === 'none' || user?.role === 'admin') && (<div className={comStyles['floating-context-menu']}>
                <button
                  className={comStyles['menu-button']}
                  aria-label="Menu"
                  onClick={(e) => {
                    e.stopPropagation();
                    setMenuOpen((open) => !open);
                  }}
                  onBlur={() => {
                    setTimeout(() => setMenuOpen(false), 120);
                  }}
                >
                  <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ verticalAlign: 'middle' }}>
                    <line x1="2" y1="4" x2="14" y2="4" />
                    <line x1="2" y1="8" x2="14" y2="8" />
                    <line x1="2" y1="12" x2="14" y2="12" />
                  </svg>
                </button>
                {menuOpen && (
                  <div
                    style={{
                      position: 'absolute',
                      top: '36px',
                      right: 0,
                      background: 'var(--color-background-secondary)',
                      border: '0.5px solid var(--color-border-tertiary)',
                      borderRadius: '8px',
                      boxShadow: '0 2px 8px rgba(0,0,0,0.07)',
                      zIndex: 10,
                      minWidth: '140px',
                    }}
                    tabIndex={-1}
                  >
                    <button
                      style={{
                        width: '100%',
                        background: 'none',
                        border: 'none',
                        padding: '10px 16px',
                        textAlign: 'left',
                        fontSize: '13px',
                        color: 'var(--color-text-primary)',
                        cursor: 'pointer',
                        borderRadius: '8px',
                      }}
                      onClick={() => {
                        setMenuOpen(false);
                        handleRefresh();
                      }}
                    >
                      <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ marginRight: 7, verticalAlign: 'center' }}>
                        <path d="M2.5 8A5.5 5.5 0 018 2.5c1.5 0 2.9.6 3.9 1.6M13.5 8A5.5 5.5 0 018 13.5c-1.5 0-2.9-.6-3.9-1.6" />
                        <path d="M12 2.5v2.5H9.5" />
                        <path d="M4 13.5v-2.5H6.5" />
                      </svg>
                      {t('button.refresh')}
                    </button>
                  </div>
                )}
              </div>)}
            </h1>
            <div className={comStyles['series-hero-metadata']}>
              {(series.yearStart > 0) ? (<>
                <span>
                  {series.yearStart}{series.yearEnd ? ` – ${series.yearEnd}` : ''}
                </span>
                <span>·</span>
              </>) : null}
              <span>
                {t('label.season', { count: series.seasonCount })}
              </span>
              <span>·</span>
              <span>
                {t('label.episode', { count: series.episodeCount })}
              </span>
              <span>·</span>
              <span>{series.genres}</span>
              <span>·</span>
              <span style={{ color: statusColor, fontWeight: 500 }}>
                {series.status === 'complete' ? t('status.complete') : series.status === 'ongoing' ? t('status.ongoing') : t('status.upcoming')}
              </span>
            </div>

            {series.seasons?.[0]?.episodes && (
              <div className={comStyles['badges-container']}>
                {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840x') && (
                  <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    4K
                  </span>
                )}
                {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('1920x') && (
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
                {series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && !series.seasons?.[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
                  <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    HDR10
                  </span>
                )}
                {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('TrueHD')) && (
                  <span className={comStyles['badge-truehd']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    TrueHD
                  </span>
                )}
                {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('E-AC-3')) && (
                  <span className={comStyles['badge-ddplus']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    Dolby Digital Plus
                  </span>
                )}
                {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && (
                  <span className={comStyles['badge-atmos']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    Atmos
                  </span>
                )}
                {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS:X') && (
                  <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    DTS:X
                  </span>
                )}
                {(series.seasons?.[0]?.episodes[0]?.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS-HD MA') && (
                  <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                    DTS-HD Master Audio
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
            )}

            <p className={comStyles['series-hero-synopsis']}>
              {series.synopsis}
            </p>

            {/* Actions */}
            <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
              {(authMode === 'none' || user?.role === 'admin') && config?.sonarrUrl && (
                <a href={`${config.sonarrUrl}/series/${slugify(series.title)}`} target="_blank" rel="noopener noreferrer" style={{ background: '#1D9E75', color: 'white', border: '0', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <img src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/sonarr-light.png" alt="Sonarr Light" style={{ width: '12px', height: '12px' }} />
                  Sonarr
                </a>
              )}
              <a href={`https://thetvdb.com/series/${series.slug || slugify(series.title)}`} target="_blank" rel="noopener noreferrer" style={{ background: 'var(--color-background-secondary)', color: 'var(--color-text-secondary)', border: '0.5px solid var(--color-border-tertiary)', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <img src={isDark ? 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tvdb-light.png' : 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tvdb-dark.png'} alt="TVDB Light" style={{ width: '12px', height: '12px' }} />
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
              borderWidth: '0 0 2px 0',
              borderStyle: 'solid',
              borderColor: idx === currentSeason ? '#1D9E75' : 'transparent',
              cursor: 'pointer',
              background: 'none',
              whiteSpace: 'nowrap',
              fontWeight: idx === currentSeason ? 500 : 400,
              transition: 'all 0.15s',
            }}
          >
            {t('section.season', { number: s.number })}
          </button>
        ))}
      </div>

      {/* Episodes */}
      {season && (
        <div style={{ padding: '16px 24px' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '12px' }}>
            <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', margin: 0 }}>
              {t('section.season', { number: season.number })} — {t('label.episode', { count: season.episodes?.length || 0 })}
            </h2>
            <div style={{ display: 'flex', gap: '12px', fontSize: '11px', color: 'var(--color-text-tertiary)' }}>
              <span style={{ color: '#1D9E75' }}>{t('label.available', { count: season.availableEps })}</span>
              {season.missingEps > 0 && <span style={{ color: '#E24B4A' }}>{t('label.missing', { count: season.missingEps })}</span>}
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            {(season.episodes || []).map((ep) => {
              const isExpanded = expandedEpisodeId === ep.id;
              return (
                <div
                  key={ep.id}
                  className={comStyles['episode-card']}
                  style={{
                    border: `0.5px solid ${isExpanded ? 'var(--color-border-secondary)' : 'var(--color-border-tertiary)'}`,
                  }}
                >
                  {/* Episode Row */}
                  <div className={comStyles['episode-card-header']} onClick={() => toggleEpisode(ep.id)}>
                    <div className={comStyles['episode-card-number']}>
                      E{String(ep.episodeNum).padStart(2, '0')}
                    </div>
                    <div className={comStyles['episode-card-info']}>
                      <div className={comStyles['episode-card-title']}>
                        {ep.title}
                      </div>
                      <div className={comStyles['episode-card-duration']}>
                        {Math.round(ep.duration / 60)} min
                      </div>
                    </div>

                    {/* Display badges: 4K, 1080p, Dolby Vision, HDR10+, HDR10, TrueHD, Dolby Digital Plus, Atmos, DTS, codec */}
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px', flexShrink: 0 }}>
                      {ep.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840x') && (
                        <span className={comStyles['badge-4k']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          4K
                        </span>
                      )}
                      {ep.mediaInfo?.videoTracks?.[0]?.resolution.includes('1920x') && (
                        <span className={comStyles['badge-1080p']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          1080p
                        </span>
                      )}
                      {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
                        <span className={comStyles['badge-dv']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          DV
                        </span>
                      )}
                      {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
                        <span className={comStyles['badge-hdr']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          HDR10+
                        </span>
                      )}
                      {ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && !ep.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
                        <span className={comStyles['badge-hdr']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          HDR10
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('TrueHD')) && (
                        <span className={comStyles['badge-truehd']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          TrueHD
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('E-AC-3')) && (
                        <span className={comStyles['badge-ddplus']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          DD+
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && (
                        <span className={comStyles['badge-atmos']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          Atmos
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS:X') && (
                        <span className={comStyles['badge-dts']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          DTS:X
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS-HD MA') && (
                        <span className={comStyles['badge-dts']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          DTS-HD MA
                        </span>
                      )}
                      {(ep.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && (
                        <span className={comStyles['badge-dts']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          DTS
                        </span>
                      )}
                      {ep.mediaInfo?.videoTracks?.[0]?.codec && (
                        <span className={comStyles['badge-codec']} style={{ fontSize: '9px', padding: '2px 6px' }}>
                          {ep.mediaInfo.videoTracks?.[0]?.codec}
                        </span>
                      )}
                      {ep.status === 'missing' && <span className={comStyles['badge-missing']} style={{ fontSize: '9px', padding: '2px 6px' }}>{t('status.missing')}</span>}
                    </div>

                    <div className={comStyles['episode-card-filesize']}>
                      {ep.fileSize ? (ep.fileSize < 1024 * 1024 * 1024 ? `${(ep.fileSize / 1024 / 1024).toFixed(1)} Mo` : `${(ep.fileSize / 1024 / 1024 / 1024).toFixed(1)} Go`) : '—'}
                    </div>

                    <div className={comStyles['episode-card-status']} style={{ background: ep.status === 'available' ? '#1D9E75' : '#E24B4A' }} />

                    {/* Expand button */}
                    <div
                      className={comStyles['episode-card-expand-button']}
                      style={{
                        border: `0.5px solid ${isExpanded ? 'var(--color-border-secondary)' : 'var(--color-border-tertiary)'}`,
                        background: isExpanded ? 'var(--color-background-secondary)' : 'var(--color-background-secondary)',
                      }}
                    >
                      <svg
                        width="9"
                        height="9"
                        viewBox="0 0 9 9"
                        fill="none"
                        stroke={isExpanded ? 'var(--color-text-secondary)' : 'var(--color-text-tertiary)'}
                        strokeWidth="1.5"
                        style={{
                          transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)',
                          transition: 'transform 0.2s',
                        }}
                      >
                        <path d="M3 1.5L6 4.5L3 7.5" />
                      </svg>
                    </div>
                  </div>

                  {/* Expandable Detail Panel */}
                  {isExpanded && ep.status === 'available' && ep.mediaInfo && (
                    <div
                      style={{
                        borderTop: '0.5px solid var(--color-border-tertiary)',
                        background: 'var(--color-background-secondary)',
                      }}
                    >
                      <div
                        style={{
                          padding: '12px 14px',
                          display: 'grid',
                          gridTemplateColumns: '1fr 1fr 1fr',
                          gap: '12px',
                        }}
                      >

                        {/* Video Section */}
                        <div>
                          <div
                            style={{
                              fontSize: '9px',
                              fontWeight: 600,
                              color: 'var(--color-text-tertiary)',
                              textTransform: 'uppercase',
                              letterSpacing: '0.08em',
                              marginBottom: '6px',
                              display: 'flex',
                              alignItems: 'center',
                              gap: '5px',
                            }}
                          >
                            <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                              <rect x="2.5" y="5.5" width="11" height="7" rx="1.2" />
                              <path d="M2.5 5.5l1.5-3 2 3 1.5-3 2 3 1.5-3 2 3" />
                            </svg>
                            {t('metadata.video.title')}
                            <div style={{ flex: 1, height: '1px', background: 'var(--color-border-secondary)' }} />
                          </div>
                          {ep.mediaInfo.videoTracks?.[0] && (
                            <>
                              <div className={comStyles['media-info-property']}>
                                <span className={comStyles['media-info-property-name']}>{t('metadata.video.codec')}</span>
                                <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].codec || '—'}</span>
                              </div>
                              <div className={comStyles['media-info-property']}>
                                <span className={comStyles['media-info-property-name']}>{t('metadata.video.resolution')}</span>
                                <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].resolution || '—'}</span>
                              </div>
                              {ep.mediaInfo.videoTracks[0].hdr && (
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.video.hdr')}</span>
                                  <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].hdr || '—'}</span>
                                </div>
                              )}
                              <div className={comStyles['media-info-property']}>
                                <span className={comStyles['media-info-property-name']}>{t('metadata.video.bitrate')}</span>
                                <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].bitrate || '—'}</span>
                              </div>
                              <div className={comStyles['media-info-property']}>
                                <span className={comStyles['media-info-property-name']}>{t('metadata.video.frameRate')}</span>
                                <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].fps || '—'} fps</span>
                              </div>
                              <div className={comStyles['media-info-property']}>
                                <span className={comStyles['media-info-property-name']}>{t('metadata.video.colorSpace')}</span>
                                <span className={comStyles['media-info-property-value']}>{ep.mediaInfo.videoTracks[0].colorSpace || '—'}</span>
                              </div>
                            </>
                          )}
                        </div>

                        {/* Audio Section */}
                        <div>
                          {ep.mediaInfo.audioTracks && ep.mediaInfo.audioTracks.length > 0 ? (
                            ep.mediaInfo.audioTracks.map((track, trackIdx) => (
                              <div key={trackIdx} style={{ marginBottom: trackIdx < (ep.mediaInfo?.audioTracks?.length || 0) - 1 ? '12px' : '0' }}>
                                <div
                                  style={{
                                    fontSize: '9px',
                                    fontWeight: 600,
                                    color: 'var(--color-text-tertiary)',
                                    textTransform: 'uppercase',
                                    letterSpacing: '0.08em',
                                    marginBottom: '6px',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '5px',
                                  }}
                                >
                                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                                    <path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path>
                                  </svg>
                                  {t('metadata.audio.title', { track: trackIdx + 1 })}
                                  <div style={{ flex: 1, height: '1px', background: 'var(--color-border-secondary)' }} />
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.audio.codec')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.codec || '—'}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.audio.channels')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.channels || '—'}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.audio.sampleRate')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.sampleRate || '—'}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.audio.bitrate')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.bitrate || '—'}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.audio.language')}</span>
                                  <span className={comStyles['media-info-property-value']}>{t(`value.language.${track.language}`) === `value.language.${track.language}` ? track.language || '—' : t(`value.language.${track.language}`)}</span>
                                </div>
                              </div>
                            ))
                          ) : (<>
                            <div
                              style={{
                                fontSize: '9px',
                                fontWeight: 600,
                                color: 'var(--color-text-tertiary)',
                                textTransform: 'uppercase',
                                letterSpacing: '0.08em',
                                marginBottom: '6px',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '5px',
                              }}
                            >
                              <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                                <path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path>
                              </svg>
                              {t('metadata.audio.title', { track: '' })}
                              <div style={{ flex: 1, height: '1px', background: 'var(--color-border-secondary)' }} />
                            </div>
                            <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', padding: '3px 0' }}>{t('message.none')}</div>
                          </>)}
                        </div>

                        {/* Subtitles Section */}
                        <div>
                          {ep.mediaInfo.subtitleTracks && ep.mediaInfo.subtitleTracks.length > 0 ? (
                            ep.mediaInfo.subtitleTracks.map((track, trackIdx) => (
                              <div key={trackIdx} style={{ marginBottom: trackIdx < (ep.mediaInfo?.subtitleTracks?.length || 0) - 1 ? '12px' : '0' }}>
                                <div
                                  style={{
                                    fontSize: '9px',
                                    fontWeight: 600,
                                    color: 'var(--color-text-tertiary)',
                                    textTransform: 'uppercase',
                                    letterSpacing: '0.08em',
                                    marginBottom: '6px',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '5px',
                                  }}
                                >
                                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                                    <rect x="2.5" y="4.5" width="11" height="7" rx="1.2" />
                                    <path d="M5 8h6M5 10h4" />
                                  </svg>
                                  {t('metadata.subtitle.title', { track: trackIdx + 1 })}
                                  <div style={{ flex: 1, height: '1px', background: 'var(--color-border-secondary)' }} />
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.subtitle.language')}</span>
                                  <span className={comStyles['media-info-property-value']}>{t(`value.language.${track.language}`) === `value.language.${track.language}` ? track.language || '—' : t(`value.language.${track.language}`)}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.subtitle.format')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.format || '—'}</span>
                                </div>
                                <div className={comStyles['media-info-property']}>
                                  <span className={comStyles['media-info-property-name']}>{t('metadata.subtitle.forced')}</span>
                                  <span className={comStyles['media-info-property-value']}>{track.forced ? t('value.yes') : t('value.no')}</span>
                                </div>
                              </div>
                            ))
                          ) : (<>
                            <div
                              style={{
                                fontSize: '9px',
                                fontWeight: 600,
                                color: 'var(--color-text-tertiary)',
                                textTransform: 'uppercase',
                                letterSpacing: '0.08em',
                                marginBottom: '6px',
                                display: 'flex',
                                alignItems: 'center',
                                gap: '5px',
                              }}
                            >
                              <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                                <rect x="2.5" y="4.5" width="11" height="7" rx="1.2" />
                                <path d="M5 8h6M5 10h4" />
                              </svg>
                              {t('metadata.subtitle.title', { track: '' })}
                              <div style={{ flex: 1, height: '1px', background: 'var(--color-border-secondary)' }} />
                            </div>

                            <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', padding: '3px 0' }}>{t('message.none')}</div>
                          </>)}
                        </div>

                        {/* File Path */}
                        <div className={comStyles['file-path']} title={ep.filePath}>
                          <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75, marginRight: 5 }}>
                            <path d="M4.5 2h5l3 3v9a1 1 0 01-1 1h-7a1 1 0 01-1-1V3a1 1 0 011-1z"></path>
                            <path d="M9.5 2v3h3"></path>
                          </svg>
                          {ep.filePath || '—'}
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
