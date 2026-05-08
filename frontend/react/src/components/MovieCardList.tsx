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
    <div className={comStyles['card-list']} onClick={onClick}>
      <div className={comStyles['card-list-poster']}>
        <div className={comStyles['card-list-poster-initial']}>
          {initials}
        </div>
        <div className={comStyles['card-list-poster-title']}>
          {movie.title}
        </div>
        <div className={comStyles['card-list-poster-status']} style={{ background: statusColor }} />
      </div>

      <div className={comStyles['card-list-content']}>
        <div className={comStyles['card-list-title']}>
          {movie.title}
        </div>
        <div className={comStyles['card-list-meta']}>
          <span>{movie.year}</span>
          <span>·</span>
          <span>{movie.duration} min</span>
          <span>·</span>
          <span>{movie.genres}</span>
          <span>·</span>
          <span style={{ color: statusColor, fontWeight: 500 }}>
            {movie.status === 'available' ? 'Disponible' : movie.status === 'missing' ? 'Manquant' : 'Problème'}
          </span>
        </div>
        <div className={comStyles['card-list-badges']}>
          {movie.mediaInfo?.videoTracks?.[0]?.resolution.includes('x2160') && <span className={comStyles['badge-4k']}>4K</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && <span className={comStyles['badge-dv']}>DV</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && <span className={comStyles['badge-hdr']}>HDR10+</span>}
          {movie.mediaInfo?.videoTracks?.[0]?.codec && <span className={comStyles['badge-codec']}>{movie.mediaInfo.videoTracks?.[0]?.codec}</span>}
          {movie.status === 'missing' && <span className={comStyles['badge-missing']}>Manquant</span>}
        </div>
      </div>
    </div>
  );
};
