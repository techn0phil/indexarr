import { useEffect, useRef } from 'react';
import { ThemeToggle } from './ThemeToggle';
import styles from '../styles/topbar.module.css';

interface TopbarProps {
  showBack: boolean;
  breadcrumb: string;
  onBack: () => void;
  searchQuery?: string;
  onSearchChange?: (query: string) => void;
}

export const Topbar = ({ showBack, breadcrumb, onBack, searchQuery = '', onSearchChange }: TopbarProps) => {
  const searchInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Focus search input when "/" is pressed
      if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes((e.target as HTMLElement).tagName)) {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: '12px', padding: '0 20px', height: '56px', background: 'var(--color-background-primary)', borderBottom: '0.5px solid var(--color-border-tertiary)' }}>
      {showBack && (
        <>
          <button className={styles['back-btn']} onClick={onBack}>
            <svg className={styles['back-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M10 12L6 8l4-4" />
            </svg>
            Retour
          </button>
          <div className={styles.separator} />
        </>
      )}

      {breadcrumb && (
        <div className={styles.breadcrumb}>
          {breadcrumb}
        </div>
      )}

      <div className={styles['search-container']}>
        <svg className={styles['search-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <circle cx="7" cy="7" r="4.5" />
          <path d="M10.5 10.5l2.5 2.5" />
        </svg>
        <input
          ref={searchInputRef}
          type="text"
          className={styles['search-input']}
          placeholder="Rechercher…"
          value={searchQuery}
          onChange={(e) => onSearchChange?.(e.target.value)}
          onFocus={(e) => {
            const shortcut = e.currentTarget.parentElement?.querySelector('[data-shortcut]') as HTMLElement;
            if (shortcut) shortcut.style.display = 'none';
          }}
          onBlur={(e) => {
            const shortcut = e.currentTarget.parentElement?.querySelector('[data-shortcut]') as HTMLElement;
            if (shortcut && !e.target.value) shortcut.style.display = 'block';
          }}
        />
        {!searchQuery && (
          <span className={styles.shortcut} data-shortcut>
            /
          </span>
        )}
      </div>

      <ThemeToggle />
    </div>
  );
};
