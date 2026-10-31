import { useState, useRef, useEffect } from 'react';
import { SortState } from '../hooks/useSortPreference';
import styles from '../styles/usermenu.module.css';

interface SortOption {
  value: string;
  label: string;
}

interface SortDropdownProps {
  options: SortOption[];
  sort: SortState;
  onSelect: (key: string) => void;
  ariaLabel: string;
}

const SortArrow = ({ order, testId }: { order: SortState['order']; testId?: string }) => (
  <svg
    viewBox="0 0 16 16"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    width="14"
    height="14"
    data-testid={testId}
  >
    {order === 'asc' ? <path d="M8 13V3M4 7l4-4 4 4" /> : <path d="M8 3v10M4 9l4 4 4-4" />}
  </svg>
);

export const SortDropdown = ({ options, sort, onSelect, ariaLabel }: SortDropdownProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const current = options.find((o) => o.value === sort.key) || options[0];

  useEffect(() => {
    if (!isOpen) return;
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen]);

  const handleSelect = (key: string) => {
    onSelect(key);
    setIsOpen(false);
  };

  return (
    <div className={styles.container} ref={dropdownRef}>
      <button
        className={styles['trigger']}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-label={ariaLabel}
      >
        <SortArrow order={sort.order} testId={sort.order === 'asc' ? 'sort-icon-asc' : 'sort-icon-desc'} />
        <span>{current.label}</span>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.chevron}>
          <path d="M4 6l4 4 4-4" />
        </svg>
      </button>

      {isOpen && (
        <div className={styles.dropdown}>
          {options.map((option) => (
            <button className={styles.menuItem} key={option.value} onClick={() => handleSelect(option.value)}>
              <span>{option.label}</span>
              {sort.key === option.value && (
                <span style={{ marginLeft: 'auto', display: 'flex' }}>
                  <SortArrow order={sort.order} />
                </span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
