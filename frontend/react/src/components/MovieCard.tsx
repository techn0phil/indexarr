import { Movie } from '../types';
import comStyles from '../styles/components.module.css';

interface MovieCardProps {
  movie: Movie;
  onClick: () => void;
}

export const MovieCard = ({ movie, onClick }: MovieCardProps) => {
  const initials = movie.title
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const statusColor = movie.status === 'available' ? '#1D9E75' : movie.status === 'missing' ? '#E24B4A' : '#EF9F27';

  return (
    <div className={comStyles['movie-card']} onClick={onClick}>
      <div
        style={{
          width: '100%',
          aspectRatio: '2/3', // Typical poster ratio
          background: 'var(--color-background-secondary)',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          gap: '4px',
        }}
      >
        {movie.poster ? (
          <img
            src={movie.poster}
            alt={movie.title}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'contain',
              background: 'var(--color-background-secondary)',
              borderRadius: 0,
              display: 'block',
              objectPosition: 'center',
            }}
          />
        ) : (
          <>
            <div style={{ fontSize: '30px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
              {initials}
            </div>
            <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', opacity: 0.4, textAlign: 'center', maxWidth: '90%' }}>
              {movie.title}
            </div>
          </>
        )}
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: '3px', background: statusColor }} />
      </div>
      <div style={{ padding: '9px 10px' }}>
        <div style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', marginBottom: '4px' }}>
          {movie.title}
        </div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '3px' }}>
          {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840') && <span className={comStyles['badge-4k']}>4K</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('1920') && <span className={comStyles['badge-1080p']}>1080p</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && <span className={comStyles['badge-dv']}>DV</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && <span className={comStyles['badge-hdr']}>HDR10+</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && !movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && <span className={comStyles['badge-hdr']}>HDR10</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'TrueHD') && <span className={comStyles['badge-truehd']}>TrueHD</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'E-AC-3') && <span className={comStyles['badge-ddplus']}>DD+</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && <span className={comStyles['badge-atmos']}>Atmos</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && <span className={comStyles['badge-dts']}>DTS</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.codec && <span className={comStyles['badge-codec']}>{movie.mediaInfo.videoTracks?.[0]?.codec}</span>}
          {movie.status === 'missing' && <span className={comStyles['badge-missing']}>Manquant</span>}
        </div>
      </div>
    </div>
  );
};
