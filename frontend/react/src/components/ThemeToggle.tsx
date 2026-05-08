import { useContext } from 'react';
import { AppContext } from '../hooks/useAppContext';
import styles from '../styles/topbar.module.css';

export const ThemeToggle = () => {
  const context = useContext(AppContext);
  
  if (!context) return null;

  const { isDark, toggleTheme } = context;

  return (
    <button 
      className={styles['theme-toggle']} 
      onClick={toggleTheme}
      aria-label="Toggle dark mode"
      title={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
    >
      {isDark ? (
        // Sun icon for light mode
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round">
          <circle cx="8" cy="8" r="3.5" />
          <line x1="8" y1="1" x2="8" y2="2.5" />
          <line x1="8" y1="13.5" x2="8" y2="15" />
          <line x1="1" y1="8" x2="2.5" y2="8" />
          <line x1="13.5" y1="8" x2="15" y2="8" />
          <line x1="3" y1="3" x2="4" y2="4" />
          <line x1="12" y1="12" x2="13" y2="13" />
          <line x1="3" y1="13" x2="4" y2="12" />
          <line x1="12" y1="4" x2="13" y2="3" />
        </svg>
      ) : (
        // Moon icon for dark mode
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <path d="M14 9a6 6 0 11-9-5.2A5.5 5.5 0 0014 9z" />
        </svg>
      )}
    </button>
  );
};
