import { Series } from '../types';
import comStyles from '../styles/components.module.css';

interface SeriesCardListProps {
  series: Series;
  onClick: () => void;
}

export const SeriesCardList = ({ series, onClick }: SeriesCardListProps) => {
  const initials = series.title
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const statusColor = series.status === 'complete' ? '#1D9E75' : series.status === 'ongoing' ? '#EF9F27' : '#E24B4A';

  return (
    <div className={comStyles['card-list']} onClick={onClick}>
      <div className={comStyles['card-list-poster']}>
        <div className={comStyles['card-list-poster-initial']}>
          {initials}
        </div>
        <div className={comStyles['card-list-poster-title']}>
          {series.title}
        </div>
        <div className={comStyles['card-list-poster-status']} style={{ background: statusColor }} />
      </div>

      <div className={comStyles['card-list-content']}>
        <div className={comStyles['card-list-title']}>
          {series.title}
        </div>
        <div className={comStyles['card-list-meta']}>
          <span>{series.yearStart}{series.yearEnd ? `-${series.yearEnd}` : '+'}</span>
          <span>·</span>
          <span>{series.seasonCount} saison{series.seasonCount > 1 ? 's' : ''}</span>
          <span>·</span>
          <span>{series.episodeCount} épisodes</span>
          <span>·</span>
          <span style={{ color: statusColor, fontWeight: 500 }}>
            {series.status === 'complete' ? 'Complète' : series.status === 'ongoing' ? 'En cours' : 'Partielle'}
          </span>
        </div>
        <div className={comStyles['card-list-badges']}>
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('x2160') && (
            <span className={comStyles['badge-4k']}>4K</span>
          )}
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('Dolby') && (
            <span className={comStyles['badge-dv']}>DV</span>
          )}
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.hdr.includes('HDR10+') && (
            <span className={comStyles['badge-hdr']}>HDR10+</span>
          )}
          {series.status === 'partial' && (
            <span className={comStyles['badge-codec']}>
              {series.seasons?.[0]?.episodes?.length ?? 0} ep. manq.
            </span>
          )}
        </div>
      </div>
    </div>
  );
};
