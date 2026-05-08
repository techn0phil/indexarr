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
    <div style={{ background: 'var(--color-background-primary)', border: '0.5px solid var(--color-border-tertiary)', borderRadius: '8px', overflow: 'hidden', cursor: 'pointer', transition: 'all 0.15s' }} onClick={onClick}>
      <div style={{ height: '180px', background: 'var(--color-background-secondary)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', position: 'relative', gap: '4px' }}>
        <div style={{ fontSize: '26px', fontWeight: 500, color: 'var(--color-text-tertiary)', opacity: 0.18 }}>
          {initials}
        </div>
        <div style={{ fontSize: '10px', color: 'var(--color-text-tertiary)', opacity: 0.4, textAlign: 'center', maxWidth: '90%' }}>
          {series.title}
        </div>
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: '3px', background: statusColor }} />
      </div>
      <div style={{ padding: '9px 10px' }}>
        <div style={{ fontSize: '12px', fontWeight: 500, color: 'var(--color-text-primary)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', marginBottom: '4px' }}>
          {series.title}
        </div>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '3px' }}>
          {series.seasons && series.seasons[0]?.episodes[0]?.mediaInfo?.videoTracks[0]?.resolution.includes('x2160') && (
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
        </div>
      </div>
    </div>
  );
};
