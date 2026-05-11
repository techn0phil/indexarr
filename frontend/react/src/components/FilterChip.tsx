import styles from '../styles/components.module.css';

interface FilterChipProps {
  icon?: React.ReactNode;
  label: string;
  active: boolean;
  count?: number;
  onClick: () => void;
}

export const FilterChip = ({ icon = <></>, label, active, count, onClick }: FilterChipProps) => {
  return (
    <div
      className={`${styles['filter-chip']} ${active ? styles['filter-chip-active'] : ''}`}
      onClick={onClick}
    >
      {icon}
      {label}
      {count !== undefined && count > 0 && (
        <span className={styles['filter-chip-badge']}>{count}</span>
      )}
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M3 4.5l3 3 3-3"></path></svg>
    </div>
  );
};
