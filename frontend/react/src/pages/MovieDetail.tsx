import { Fragment, useEffect, useState } from 'react';
import { Movie } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';

interface MovieDetailProps {
  movieId: number;
}

export const MovieDetail = ({ movieId }: MovieDetailProps) => {
  const [movie, setMovie] = useState<Movie | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
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

    fetchMovie();
  }, [movieId]);

  if (loading) return <div style={{ padding: '20px' }}>Chargement...</div>;
  if (!movie) return <div style={{ padding: '20px' }}>Film non trouvé</div>;

  return (
    <div style={{ padding: '24px' }}>
      {/* Hero */}
      <div style={{ display: 'flex', gap: '20px', alignItems: 'flex-start', marginBottom: '24px', paddingBottom: '20px', borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
        {/* Poster */}
        <div style={{ width: '110px', minWidth: '110px', height: '160px', background: 'var(--color-background-secondary)', borderRadius: '8px', border: '0.5px solid var(--color-border-tertiary)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: '6px', overflow: 'hidden' }}>
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
            </>
          )}
        </div>

        {/* Info */}
        <div style={{ flex: 1 }}>
          <h1 style={{ fontSize: '22px', fontWeight: 500, color: 'var(--color-text-primary)', marginBottom: '4px' }}>
            {movie.title}
          </h1>
          <div style={{ fontSize: '13px', color: 'var(--color-text-tertiary)', marginBottom: '10px', display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
            <span>{movie.year}</span>
            <span>·</span>
            <span>{movie.duration} min</span>
            <span>·</span>
            <span>{movie.genres}</span>
            <span>·</span>
            <span style={{ color: '#1D9E75', fontWeight: 500 }}>
              {movie.status === 'available' ? 'Disponible' : 'Manquant'}
            </span>
          </div>

          {/* Badges */}
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: '5px', marginBottom: '12px' }}>
            {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('x2160') && (
              <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                4K
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('x1080') && (
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
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && (
              <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                HDR10
              </span>
            )}
            {movie.mediaInfo?.audioTracks.find((track) => track.codec === 'TrueHD') && (
              <span className={comStyles['badge-truehd']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                TrueHD
              </span>
            )}
            {movie.mediaInfo?.audioTracks.find((track) => track.codec === 'E-AC-3') && (
              <span className={comStyles['badge-ddplus']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Dolby Digital Plus
              </span>
            )}
            {movie.mediaInfo?.audioTracks.find((track) => track.codec.includes('Atmos')) && (
              <span className={comStyles['badge-atmos']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Atmos
              </span>
            )}
            {movie.mediaInfo?.audioTracks.find((track) => track.codec === 'DTS') && (
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
          <p style={{ fontSize: '12px', color: 'var(--color-text-secondary)', lineHeight: 1.6, maxWidth: '560px' }}>
            {movie.synopsis}
          </p>

          {/* Actions */}
          <div style={{ display: 'flex', gap: '8px', marginTop: '14px' }}>
            <button style={{ background: '#1D9E75', color: 'white', border: 'none', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', fontWeight: 500, cursor: 'pointer' }}>
              + Rechercher upgrade
            </button>
            <button style={{ background: 'var(--color-background-secondary)', color: 'var(--color-text-secondary)', border: '0.5px solid var(--color-border-tertiary)', padding: '6px 13px', borderRadius: '6px', fontSize: '12px', cursor: 'pointer' }}>
              Ouvrir dans Radarr
            </button>
          </div>
        </div>
      </div>

      {/* Cast */}
      <div style={{ marginBottom: '24px' }}>
        <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', marginBottom: '12px' }}>
          Cast principal
        </h2>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(80px, 1fr))', gap: '10px', background: 'var(--color-background-secondary)', border: '0.5px solid var(--color-border-secondary)', borderRadius: '8px', padding: '14px 16px' }}>
          {movie.cast?.slice(0, 5).map((c) => (
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

      {/* MediaInfo Table */}
      {movie.mediaInfo && (
        <div>
          <h2 style={{ fontSize: '11px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase', letterSpacing: '0.07em', marginBottom: '12px' }}>
            Métadonnées du fichier
          </h2>
          <div style={{ background: 'var(--color-background-tertiary)', border: '0.5px solid var(--color-border-tertiary)', borderRadius: '8px', overflow: 'hidden' }}>
            {/* Video */}
            {movie.mediaInfo?.videoTracks.map((videoTrack, index) => (
              <Fragment key={index}>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                    <rect x="2.5" y="5.5" width="11" height="7" rx="1.2" />
                    <path d="M2.5 5.5l1.5-3 2 3 1.5-3 2 3 1.5-3 2 3" />
                  </svg>
                  Vidéo — piste {index + 1}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Codec
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.codec}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Résolution
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.resolution}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Bitrate
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.bitrate}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        FPS
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.fps}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        HDR
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.hdr}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Espace colorimétrique
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {videoTrack.colorSpace}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </Fragment>
            ))}

            {/* Audio */}
            {movie.mediaInfo?.audioTracks.map((audioTrack, index) => (
              <Fragment key={index}>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-secondary)', textTransform: 'uppercase', display: 'flex', alignItems: 'center', gap: '6px' }}>
                  <svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ display: 'inline', verticalAlign: 'middle', opacity: 0.75 }}>
                    <path d="M5 4L3 6H1.5v1.5H3l2 2zM8 4.5a2.5 2.5 0 010 3"></path>
                  </svg>
                  Audio — piste {index + 1}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Codec
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.codec}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Canaux
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.channels}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Bitrate
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.bitrate}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Langue
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.language}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Fréquence d'échantillonnage
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {audioTrack.sampleRate}
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
                  Sous-titres — piste {index + 1}
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Langue
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {subtitleTrack.language}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Format
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {subtitleTrack.format}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Forcé
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {subtitleTrack.forced ? 'Oui' : 'Non'}
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
