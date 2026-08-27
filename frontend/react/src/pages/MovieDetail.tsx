import { Fragment, useEffect, useState } from 'react';
import { Movie } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';
import { useAppContext } from '../hooks/useAppContext';
import { useTranslation } from 'react-i18next';

interface MovieDetailProps {
  movieId: number;
}

export const MovieDetail = ({ movieId }: MovieDetailProps) => {
  const { t } = useTranslation('movie-details');
  const [movie, setMovie] = useState<Movie | null>(null);
  const [loading, setLoading] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const { authMode, config, isDark, user } = useAppContext();

  // Fetch movie function, used for both initial load and refresh
  const fetchMovie = async () => {
    setLoading(true);
    try {
      const data = await apiClient.getMovie(movieId);
      setMovie(data);
    } catch (error) {
      console.error('Failed to fetch movie:', error);
    } finally {
      setLoading(false);
    }
  };

  // Refresh handler for menu
  const handleRefresh = async () => {
    const response = await apiClient.refreshMovie(movieId);
    
    if (response.result?.filesFound) {
      fetchMovie();
    }
    else {
      window.location.reload();
    }
  };

  useEffect(() => {
    fetchMovie();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [movieId]);

  if (loading) return <div style={{ padding: '20px' }}>{t('message.loading')}</div>;
  if (!movie) return <div style={{ padding: '20px' }}>{t('message.movieNotFound')}</div>;

  return (
    <div style={{ paddingBottom: '24px' }}>
      {/* Hero */}
      <div className={comStyles['movie-hero']}>
        {/* Poster */}
        <div className={comStyles['movie-hero-poster']}>
          {movie.poster && movie.poster.startsWith('http') ? (
            <img
              src={movie.poster}
              alt={movie.title}
              style={{ width: '100%', height: '100%', objectFit: 'cover', borderRadius: '8px' }}
            />
          ) : (
            <>
              <div style={{ fontSize: '30px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
                {movie.title[0]}
              </div>
              <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', opacity: 0.4, textAlign: 'center', padding: '0 6px' }}>
                {movie.title}
              </div>
              <div style={{ fontSize: '14px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
                {`${movie.year || ''}`}
              </div>
            </>
          )}
        </div>

        {/* Info */}
        <div style={{ flex: 1 }}>
          <h1 className={comStyles['movie-hero-title']}>
            {movie.title}
            {typeof movie.rating === 'number' && (
              <span className={comStyles['movie-hero-rating']}>
                <svg width="11" height="11" viewBox="0 0 12 12" fill="var(--color-badge-rating-text)" style={{ marginRight: '2px', flexShrink: 0 }} aria-hidden="true"><path d="M6 1l1.4 3h3.1l-2.5 1.9 1 3L6 7.2l-3 1.7 1-3L1.5 4H4.6z"></path></svg>
                <span style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-badge-rating-text)', lineHeight: 1 }}>{movie.rating?.toFixed(1)}</span>
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
          <div className={comStyles['movie-hero-metadata']}>
            <span>{movie.year}</span>
            <span>·</span>
            <span>{Math.floor(movie.duration / 60)}h {movie.duration % 60}min</span>
            <span>·</span>
            <span>{movie.genres}</span>
            <span>·</span>
            <span style={{ color: '#1D9E75', fontWeight: 500 }}>
              {movie.status === 'available' ? t('status.available') : t('status.missing')}
            </span>
          </div>

          {/* Badges */}
          <div className={comStyles['badges-container']}>
            {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840x') && (
              <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                4K
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('1920x') && (
              <span className={comStyles['badge-1080p']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                1080p
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
              <span className={comStyles['badge-dv']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Dolby Vision
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
              <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                HDR10+
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && !movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
              <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                HDR10
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('TrueHD')) && (
              <span className={comStyles['badge-truehd']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                TrueHD
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('E-AC-3')) && (
              <span className={comStyles['badge-ddplus']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Dolby Digital Plus
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && (
              <span className={comStyles['badge-atmos']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Atmos
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS:X') && (
              <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                DTS:X
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS-HD MA') && (
              <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                DTS-HD Master Audio
              </span>
            )}
            {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && (
              <span className={comStyles['badge-dts']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                DTS
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.codec && (
              <span className={comStyles['badge-codec']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                {movie.mediaInfo.videoTracks?.[0]?.codec}
              </span>
            )}
          </div>

          {/* Synopsis */}
          <p className={comStyles['movie-hero-synopsis']}>
            {movie.synopsis}
          </p>

          {/* Actions */}
          <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
            {(authMode === 'none' || user?.role === 'admin') && config?.radarrUrl && (
              <a href={`${config.radarrUrl}/movie/${movie.tmdbId}`} target="_blank" rel="noopener noreferrer" style={{ background: '#1D9E75', color: 'white', border: '0', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
                <img src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/radarr-light.png" alt="Radarr Light" style={{ width: '12px', height: '12px' }} />
                Radarr
              </a>
            )}
            <a href={`https://www.themoviedb.org/movie/${movie.tmdbId}`} target="_blank" rel="noopener noreferrer" style={{ background: 'var(--color-background-secondary)', color: 'var(--color-text-secondary)', border: '0.5px solid var(--color-border-tertiary)', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <img src={isDark ? 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tmdb-light.png' : 'https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/tmdb-dark.png'} alt="TMDB Light" style={{ width: '12px', height: '12px' }} />
              TMDB
            </a>
          </div>
        </div>
      </div>

      {/* Cast */}
      {movie.cast?.length && (
        <div style={{ marginBottom: '24px', padding: '0 24px' }}>
          <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', marginBottom: '12px' }}>
            {t('section.cast')}
          </h2>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(80px, 1fr))', gap: '10px', background: 'var(--color-background-primary)', border: '0.5px solid var(--color-border-secondary)', borderRadius: '8px', padding: '14px 16px' }}>
            {movie.cast.map((c) => (
              <div key={c.id} style={{ textAlign: 'center' }}>
                {c.avatar && c.avatar.startsWith('http') ? (
                  <img
                    src={c.avatar}
                    alt={c.name}
                    style={{
                      width: '44px',
                      height: '44px',
                      borderRadius: '50%',
                      objectFit: 'cover',
                      border: '0.5px solid var(--color-border-tertiary)',
                      margin: '0 auto 6px',
                      display: 'block',
                    }}
                  />
                ) : (
                  <div
                    style={{
                      width: '44px',
                      height: '44px',
                      borderRadius: '50%',
                      background: 'var(--color-background-tertiary)',
                      border: '0.5px solid var(--color-border-tertiary)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontSize: '12px',
                      fontWeight: 500,
                      color: 'var(--color-text-tertiary)',
                      margin: '0 auto 6px',
                    }}
                  >
                    {c.name ? c.name[0] : '?'}
                  </div>
                )}
                <div style={{ fontSize: '10px', fontWeight: 500, color: 'var(--color-text-primary)' }}>
                  {c.name}
                </div>
                <div style={{ fontSize: '9px', color: 'var(--color-text-tertiary)' }}>
                  {c.role}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* MediaInfo Table */}
      {movie.mediaInfo && (
        <div style={{ padding: '0 24px' }}>
          <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', marginBottom: '12px' }}>
            {t('section.metadata')}
          </h2>
          <div style={{ background: 'var(--color-background-primary)', border: '0.5px solid var(--color-border-tertiary)', borderRadius: '8px', overflow: 'hidden' }}>
            {/* File */}
            <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
              <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                <path d="M4.5 2h5l3 3v9a1 1 0 01-1 1h-7a1 1 0 01-1-1V3a1 1 0 011-1z"></path>
                <path d="M9.5 2v3h3"></path>
              </svg>
              {t('metadata.file.title')}
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <tbody>
                <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                  <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                    {t('metadata.file.path')}
                  </td>
                  <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                    {movie.filePath}
                  </td>
                </tr>
                <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                  <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                    {t('metadata.file.size')}
                  </td>
                  <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                  {movie.fileSize < 1024 * 1024 * 1024 ? `${(movie.fileSize / 1024 / 1024).toFixed(1)} Mo` : `${(movie.fileSize / 1024 / 1024 / 1024).toFixed(1)} Go`}
                  </td>
                </tr>
              </tbody>
            </table>

            {/* Video */}
            {(movie.mediaInfo?.videoTracks ?? []).map((videoTrack, index) => (
              <Fragment key={index}>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                    <rect x="2.5" y="5.5" width="11" height="7" rx="1.2" />
                    <path d="M2.5 5.5l1.5-3 2 3 1.5-3 2 3 1.5-3 2 3" />
                  </svg>
                  {t('metadata.video.title', { track: index + 1 })}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.codec')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.codec || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.resolution')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.resolution || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.hdr')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.hdr || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.bitrate')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.bitrate || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.fps')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.fps || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.video.colorSpace')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.colorSpace || t('value.unknown')}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </Fragment>
            ))}

            {/* Audio */}
            {(movie.mediaInfo?.audioTracks ?? []).map((audioTrack, index) => (
              <Fragment key={index}>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                    <path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path>
                  </svg>
                  {t('metadata.audio.title', { track: index + 1 })}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.audio.codec')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.codec || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.audio.channels')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.channels || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.audio.language')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {t(`value.language.${audioTrack.language}`) === `value.language.${audioTrack.language}` ? audioTrack.language : t(`value.language.${audioTrack.language}`)}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.audio.bitrate')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.bitrate || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.audio.sampleRate')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.sampleRate || t('value.unknown')}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </Fragment>
            ))}

            {/* Subtitles */}
            {(movie.mediaInfo?.subtitleTracks ?? []).map((subtitleTrack, index) => (
              <Fragment key={index}>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                    <rect x="2.5" y="4.5" width="11" height="7" rx="1.2" />
                    <path d="M5 8h6M5 10h4" />
                  </svg>
                  {t('metadata.subtitle.title', { track: index + 1 })}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.subtitle.format')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {subtitleTrack.format || t('value.unknown')}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.subtitle.language')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {t(`value.language.${subtitleTrack.language}`) === `value.language.${subtitleTrack.language}` ? subtitleTrack.language : t(`value.language.${subtitleTrack.language}`)}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        {t('metadata.subtitle.forced')}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {subtitleTrack.forced ? t('value.yes') : t('value.no')}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </Fragment>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};
