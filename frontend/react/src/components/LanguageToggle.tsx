import { useState, useRef, useEffect } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import styles from '../styles/usermenu.module.css';

const LANGUAGE_OPTIONS = [
  { code: 'en', label: 'English', flag: '🇬🇧' },
  { code: 'fr', label: 'Français', flag: '🇫🇷' },
  { code: 'de', label: 'Deutsch', flag: '🇩🇪' },
  { code: 'it', label: 'Italiano', flag: '🇮🇹' },
  { code: 'es', label: 'Español', flag: '🇪🇸' },
];

export const LanguageToggle = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { locale, setLocale } = useAppContext();
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Find current language label
  const currentLanguage = LANGUAGE_OPTIONS.find(lang => lang.code === locale) || LANGUAGE_OPTIONS[0];

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen]);

  const handleSelectLanguage = (code: string) => {
    setLocale(code);
    setIsOpen(false);
  };

  return (
    <div className={styles.container} ref={dropdownRef}>
      <button
        className={styles['trigger']}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-label="Select language"
        title={currentLanguage.label}
      >
        <span>{currentLanguage.flag}</span>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.chevron}>
          <path d="M4 6l4 4 4-4" />
        </svg>
      </button>

      {isOpen && (
        <div className={styles.dropdown}>
          {LANGUAGE_OPTIONS.map((lang) => (
            <button
              className={styles.menuItem}
              key={lang.code}
              onClick={() => handleSelectLanguage(lang.code)}
            >
              <span style={{ fontSize: '16px' }}>{lang.flag}</span>
              <span>{lang.label}</span>
              {locale === lang.code && (
                <span style={{ marginLeft: 'auto', fontSize: '12px', fontWeight: 500 }}>✓</span>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};
