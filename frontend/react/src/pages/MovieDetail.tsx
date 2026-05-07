import { useEffect, useState } from 'react';
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
          <h1 style={{ fontSize: '22px', fontWeight: 500, color: 'var(--color-text-primary)', marginBottom: '4px', display: 'flex', alignItems: 'center', gap: '10px' }}>
            {movie.title}
            {typeof movie.rating === 'number' && (
              <span style={{
                display: 'inline-flex',
                alignItems: 'center',
                gap: '3px',
                background: '#FAF3E2',
                color: '#BA7517',
                borderRadius: '99px',
                fontSize: '12px',
                fontWeight: 500,
                padding: '2px 10px 2px 7px',
                border: 'none',
                lineHeight: 1,
                minWidth: '36px',
                height: '22px',
              }}>
                <svg width="11" height="11" viewBox="0 0 12 12" fill="#BA7517" style={{ marginRight: '2px', flexShrink: 0 }} aria-hidden="true"><path d="M6 1l1.4 3h3.1l-2.5 1.9 1 3L6 7.2l-3 1.7 1-3L1.5 4H4.6z"></path></svg>
                <span style={{ fontSize: '12px', fontWeight: 500, color: '#BA7517', lineHeight: 1 }}>{movie.rating?.toFixed(1)}</span>
              </span>
            )}
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
            {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840') && (
              <span className={comStyles['badge-4k']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                4K UHD
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
              <span className={comStyles['badge-dv']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                Dolby Vision
              </span>
            )}
            {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && (
              <span className={comStyles['badge-hdr']} style={{ fontSize: '10px', padding: '3px 8px' }}>
                HDR10
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
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(80px, 1fr))', gap: '10px', background: 'var(--color-background-primary)', border: '0.5px solid var(--color-border-tertiary)', borderRadius: '8px', padding: '14px 16px' }}>
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
                    background: 'var(--color-background-secondary)',
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
          <div style={{ background: 'var(--color-background-primary)', border: '0.5px solid var(--color-border-tertiary)', borderRadius: '8px', overflow: 'hidden' }}>
            {/* Video */}
            <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase' }}>
              Vidéo — piste 1
            </div>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <tbody>
                {movie.mediaInfo?.videoTracks?.[0] && (
                  <>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Codec
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.videoTracks?.[0]?.codec}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Résolution
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.videoTracks?.[0]?.resolution}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Bitrate
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.videoTracks?.[0]?.bitrate}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        HDR
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.videoTracks?.[0]?.hdr}
                      </td>
                    </tr>
                  </>
                )}
              </tbody>
            </table>

            {/* Audio */}
            {movie.mediaInfo?.audioTracks && movie.mediaInfo.audioTracks.length > 0 && (
              <>
                <div style={{ padding: '8px 8px 4px', background: 'var(--color-background-secondary)', fontSize: '10px', fontWeight: 500, color: 'var(--color-text-tertiary)', textTransform: 'uppercase' }}>
                  Audio — piste 1
                </div>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                  <tbody>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Codec
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.audioTracks?.[0]?.codec}
                      </td>
                    </tr>
                    <tr style={{ borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-tertiary)', padding: '7px 8px', width: '38%' }}>
                        Canaux
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--color-text-secondary)', padding: '7px 8px' }}>
                        {movie.mediaInfo?.audioTracks?.[0]?.channels}
                      </td>
                    </tr>
                  </tbody>
                </table>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
