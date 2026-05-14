import { Movie } from '../types';
import comStyles from '../styles/components.module.css';

interface MovieCardListProps {
  movie: Movie;
  onClick: () => void;
}

export const MovieCardList = ({ movie, onClick }: MovieCardListProps) => {
  const initials = movie.title
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const statusColor = movie.status === 'available' ? '#1D9E75' : movie.status === 'missing' ? '#E24B4A' : '#EF9F27';

  return (
    <div className={comStyles['card-list']} onClick={onClick} style={{ display: 'flex', gap: '16px' }}>
      {movie.poster ? (
        <div className={comStyles['card-list-poster']}>
          <img
            src={movie.poster}
            alt={movie.title}
            width="100%"
            height="100%"
            style={{ objectFit: 'contain' }}
            className={comStyles['card-list-poster-img']}
          />
          <div className={comStyles['card-list-poster-status']} style={{ background: statusColor }} />
        </div>
      ) : (
        <div className={comStyles['card-list-poster']}>
          <div className={comStyles['card-list-poster-initial']}>
            {initials}
          </div>
          <div className={comStyles['card-list-poster-title']}>
            {movie.title}
          </div>
          <div className={comStyles['card-list-poster-status']} style={{ background: statusColor }} />
        </div>
      )}

      <div className={comStyles['card-list-content']}>
        <div className={comStyles['card-list-title']} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {movie.title}
          {typeof movie.rating === 'number' && (
            <span style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '3px',
              background: 'var(--color-badge-rating)',
              color: 'var(--color-badge-rating-text)',
              borderRadius: '99px',
              fontSize: '10px',
              fontWeight: 500,
              padding: '1px 6px 1px 4px',
              border: 'none',
              lineHeight: 1,
              minWidth: '32px',
              height: '16px',
              flexShrink: 0,
            }}>
              <svg width="10" height="10" viewBox="0 0 12 12" fill="var(--color-badge-rating-text)" style={{ marginRight: '1px', flexShrink: 0 }} aria-hidden="true"><path d="M6 1l1.4 3h3.1l-2.5 1.9 1 3L6 7.2l-3 1.7 1-3L1.5 4H4.6z"></path></svg>
              <span style={{ fontSize: '10px', fontWeight: 500, color: 'var(--color-badge-rating-text)', lineHeight: 1 }}>{movie.rating?.toFixed(1)}</span>
            </span>
          )}
        </div>
        <div className={comStyles['card-list-meta']}>
          <span>{movie.year}</span>
          <span>·</span>
          <span>{Math.floor(movie.duration / 60)}h {movie.duration % 60}min</span>
          <span>·</span>
          <span>{movie.genres}</span>
          <span>·</span>
          <span style={{ color: statusColor, fontWeight: 500 }}>
            {movie.status === 'available' ? 'Disponible' : movie.status === 'missing' ? 'Manquant' : 'Problème'}
          </span>
        </div>
        <div className={comStyles['card-list-badges']}>
          {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840') && <span className={comStyles['badge-4k']}>4K</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('1920') && <span className={comStyles['badge-1080p']}>1080p</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && <span className={comStyles['badge-dv']}>Dolby Vision</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && <span className={comStyles['badge-hdr']}>HDR10+</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10') && !movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && <span className={comStyles['badge-hdr']}>HDR10</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'TrueHD') && <span className={comStyles['badge-truehd']}>TrueHD</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'E-AC-3') && <span className={comStyles['badge-ddplus']}>Dolby Digital Plus</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec.includes('Atmos')) && <span className={comStyles['badge-atmos']}>Atmos</span>}
          {(movie.mediaInfo?.audioTracks ?? []).find((track) => track.codec === 'DTS') && <span className={comStyles['badge-dts']}>DTS</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.codec && <span className={comStyles['badge-codec']}>{movie.mediaInfo.videoTracks?.[0]?.codec}</span>}
          {movie.status === 'missing' && <span className={comStyles['badge-missing']}>Manquant</span>}
        </div>
      </div>

      <div className={comStyles['card-list-synopsis']}>
      {/* <div style={{ flex: '1 1 auto', display: 'flex', alignItems: 'center', paddingRight: '8px', maxWidth: '50%' }}> */}
        <p style={{ fontSize: '12px', color: 'var(--color-text-secondary)', lineHeight: 1.5, margin: 0, textAlign: 'justify', display: '-webkit-box', WebkitLineClamp: 4, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
          {movie.synopsis}
        </p>
      </div>
    </div>
  );
};
