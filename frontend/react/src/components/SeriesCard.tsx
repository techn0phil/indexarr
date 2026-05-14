import { Series } from '../types';
import comStyles from '../styles/components.module.css';

interface SeriesCardProps {
  series: Series;
  onClick: () => void;
}

export const SeriesCard = ({ series, onClick }: SeriesCardProps) => {
  const initials = series.title
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);

  const statusColor =
    series.status === 'complete' ? '#1D9E75' : series.status === 'ongoing' ? '#EF9F27' : '#E24B4A';

  return (
    <div className={comStyles['tv-show-card']} onClick={onClick}>
      <div
        style={{
          width: '100%',
          aspectRatio: '2 / 3',
          background: 'var(--color-background-secondary)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
          gap: '4px',
        }}
      >
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
            <div style={{ fontSize: '30px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
              {initials}
            </div>
            <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', opacity: 0.4, textAlign: 'center', maxWidth: '90%' }}>
              {series.title}
            </div>
            <div style={{ fontSize: '14px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
              {`${series.yearStart || '?'} - ${series.yearEnd || '?'}`}
            </div>
          </>
        )}
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: '3px', background: statusColor }} />
      </div>
      <div style={{ padding: '9px 10px' }}>
        <div style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', marginBottom: '4px' }}>
          {series.title}
        </div>
        {/* As we don't have seasons loaded for performance reasons, we won't show badges for now.
        We could consider fetching seasons/episodes on demand when hovering or clicking on a series card in the future. */}
        {/* <div style={{ display: 'flex', flexWrap: 'wrap', gap: '3px' }}>
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks[0]?.resolution.includes('3840') && (
            <span className={comStyles['badge-4k']}>4K</span>
          )}
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks[0]?.hdr.includes('Dolby') && (
            <span className={comStyles['badge-dv']}>DV</span>
          )}
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks[0]?.hdr.includes('HDR10+') && (
            <span className={comStyles['badge-hdr']}>HDR10+</span>
          )}
          {series.status === 'partial' && (
            <span className={comStyles['badge-codec']}>
              {series.seasonCount * 10 - series.episodeCount} ep. manq.
            </span>
          )}
        </div> */}
      </div>
    </div>
  );
};
