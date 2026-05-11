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
        {series.poster ? (
          <img
            src={series.poster}
            alt={series.title}
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
            <div className={comStyles['card-list-poster-initial']}>
              {initials}
            </div>
            <div className={comStyles['card-list-poster-title']}>
              {series.title}
            </div>
          </>
        )}
        <div className={comStyles['card-list-poster-status']} style={{ background: statusColor }} />
      </div>

      <div className={comStyles['card-list-content']}>
        <div className={comStyles['card-list-title']} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          {series.title}
          {typeof series.rating === 'number' && (
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
              <span style={{ fontSize: '10px', fontWeight: 500, color: 'var(--color-badge-rating-text)', lineHeight: 1 }}>{series.rating?.toFixed(1)}</span>
            </span>
          )}
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
        <div className={comStyles['card-list-meta']}>
          <span>{series.genres}</span>
        </div>
        {/* <div className={comStyles['card-list-badges']}>
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks?.[0]?.resolution.includes('3840') && (
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
        </div> */}
      </div>

      <div className={comStyles['card-list-synopsis']}>
      {/* <div style={{ flex: '1 1 auto', display: 'flex', alignItems: 'center', paddingRight: '8px', maxWidth: '50%' }}> */}
        <p style={{ fontSize: '12px', color: 'var(--color-text-secondary)', lineHeight: 1.5, margin: 0, textAlign: 'justify', display: '-webkit-box', WebkitLineClamp: 4, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>
          {series.synopsis}
        </p>
      </div>
    </div>
  );
};
