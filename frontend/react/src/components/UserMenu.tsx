import { useState, useRef, useEffect } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import styles from '../styles/usermenu.module.css';

export const UserMenu = () => {
  const { user, authMode, logout } = useAppContext();
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Don't render if no auth or no user
  if (authMode === 'none' || !user) {
    return null;
  }

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleLogout = async () => {
    setIsOpen(false);
    await logout();
  };

  return (
    <div className={styles.container} ref={menuRef}>
      <button 
        className={styles.trigger}
        onClick={() => setIsOpen(!isOpen)}
        aria-expanded={isOpen}
        aria-haspopup="true"
      >
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.icon}>
          <circle cx="8" cy="5" r="3" />
          <path d="M2 14c0-3 2.5-5 6-5s6 2 6 5" />
        </svg>
        <span className={styles.username}>{user.username}</span>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.chevron}>
          <path d="M4 6l4 4 4-4" />
        </svg>
      </button>

      {isOpen && (
        <div className={styles.dropdown}>
          <div className={styles.userInfo}>
            <span className={styles.userName}>{user.username}</span>
            <span className={styles.userRole}>{user.role === 'admin' ? 'Administrateur' : 'Invité'}</span>
          </div>
          <div className={styles.separator} />
          <button className={styles.menuItem} onClick={handleLogout}>
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.menuIcon}>
              <path d="M6 2H3a1 1 0 00-1 1v10a1 1 0 001 1h3M11 11l3-3-3-3M14 8H6" />
            </svg>
            Déconnexion
          </button>
        </div>
      )}
    </div>
  );
};
