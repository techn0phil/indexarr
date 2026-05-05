import styles from '../styles/components.module.css';

interface ViewToggleProps {
  view: 'grid' | 'list';
  onViewChange: (view: 'grid' | 'list') => void;
}

export const ViewToggle = ({ view, onViewChange }: ViewToggleProps) => {
  return (
    <div className={styles['view-toggle']}>
      <button
        className={`${styles['view-toggle-btn']} ${view === 'grid' ? styles['view-toggle-btn-active'] : ''}`}
        onClick={() => onViewChange('grid')}
        aria-label="Vue grille"
      >
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <rect x="2" y="2" width="5" height="5" rx="1" />
          <rect x="9" y="2" width="5" height="5" rx="1" />
          <rect x="2" y="9" width="5" height="5" rx="1" />
          <rect x="9" y="9" width="5" height="5" rx="1" />
        </svg>
      </button>
      <button
        className={`${styles['view-toggle-btn']} ${view === 'list' ? styles['view-toggle-btn-active'] : ''}`}
        onClick={() => onViewChange('list')}
        aria-label="Vue liste"
      >
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <line x1="5" y1="4" x2="14" y2="4" />
          <line x1="5" y1="8" x2="14" y2="8" />
          <line x1="5" y1="12" x2="14" y2="12" />
          <rect x="2" y="3" width="2" height="2" rx="0.5" />
          <rect x="2" y="7" width="2" height="2" rx="0.5" />
          <rect x="2" y="11" width="2" height="2" rx="0.5" />
        </svg>
      </button>
    </div>
  );
};
