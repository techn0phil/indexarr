import { useState } from 'react';
import styles from '../styles/sidebar.module.css';
import { Page } from '../hooks/useAppContext';
import { useAppContext } from '../hooks/useAppContext';
import { apiClient } from '../api/client';

interface SidebarProps {
  activeNav: string;
  onNavClick: (page: Page, id?: number) => void;
}

export const Sidebar = ({ activeNav, onNavClick }: SidebarProps) => {
  const [showPurgeConfirm, setShowPurgeConfirm] = useState(false);
  const [isPurging, setIsPurging] = useState(false);
  const context = useAppContext();

  const handlePurge = async () => {
    setIsPurging(true);
    try {
      const result = await apiClient.purgeDatabase();
      if (result.success) {
        setShowPurgeConfirm(false);
        // Reload the page to refresh the UI
        window.location.reload();
      }
    } catch (error) {
      console.error('Purge failed:', error);
    } finally {
      setIsPurging(false);
    }
  };

  return (
    <div className={styles.sidebar}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '0 14px' }}>
        <div className={styles['logo-mark']}>
          <svg viewBox="0 0 14 14" style={{ width: '13px', height: '13px', fill: 'white' }}>
            <path d="M2 11L7 3L12 11Z" />
          </svg>
        </div>
        <span className={styles['logo-name']}>Indexarr</span>
      </div>

      <nav className={styles.nav}>
        <div className={styles['nav-group']}>Librairie</div>

        <div
          className={`${styles['nav-item']} ${activeNav === 'list-films' ? styles.active : ''}`}
          onClick={() => onNavClick('list-films')}
        >
          <svg className={styles['nav-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <rect x="2" y="3" width="12" height="10" rx="1.5" />
            <path d="M5 3v10M11 3v10M2 7h12" />
          </svg>
          Films
          <span className={styles['nav-badge']}>{context?.stats?.totalMovies ?? 0}</span>
        </div>

        <div
          className={`${styles['nav-item']} ${activeNav === 'list-series' ? styles.active : ''}`}
          onClick={() => onNavClick('list-series')}
        >
          <svg className={styles['nav-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <rect x="2" y="2" width="12" height="12" rx="1.5" />
            <path d="M2 6h12M6 6v8" />
          </svg>
          Séries
          <span className={styles['nav-badge']}>{context?.stats?.totalSeries ?? 0}</span>
        </div>

        <div className={styles['nav-item']}>
          <svg className={styles['nav-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="8" cy="8" r="5" />
            <path d="M8 5v3l2 2" />
          </svg>
          Récents
        </div>

        <div className={styles['nav-group']} style={{ marginTop: '6px' }}>
          Analyse
        </div>

        <div className={styles['nav-item']}>
          <svg className={styles['nav-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M2 12l4-5 3 3 3-4 2 2" />
          </svg>
          Statistiques
        </div>

        <div className={styles['nav-item']}>
          <svg className={styles['nav-icon']} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="8" cy="8" r="6" />
            <path d="M8 5v4M8 11h.01" />
          </svg>
          Problèmes
          <span className={styles['nav-badge']} style={{ background: 'var(--color-badge-problem)', color: 'var(--color-badge-problem-text)', borderColor: 'var(--color-border-secondary)' }}>
            {context?.stats?.problemsCount ?? 0}
          </span>
        </div>
      </nav>

      <div style={{ padding: '6px 18px', marginTop: 'auto' }}>
        <button
          onClick={() => setShowPurgeConfirm(true)}
          style={{
            marginTop: '8px',
            padding: '6px 12px',
            fontSize: '11px',
            background: 'var(--color-badge-problem)',
            color: 'var(--color-badge-problem-text)',
            border: '0.5px solid var(--color-border-secondary)',
            borderRadius: '4px',
            cursor: 'pointer',
            width: '100%',
            fontWeight: 500,
          }}
        >
          Purger DB
        </button>

        {showPurgeConfirm && (
          <div
            style={{
              position: 'fixed',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: 'rgba(0, 0, 0, 0.5)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              zIndex: 1000,
            }}
            onClick={() => !isPurging && setShowPurgeConfirm(false)}
          >
            <div
              style={{
                background: 'var(--color-background-primary)',
                padding: '20px',
                borderRadius: '8px',
                border: '0.5px solid var(--color-border-tertiary)',
                maxWidth: '300px',
                boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
              }}
              onClick={(e) => e.stopPropagation()}
            >
              <h3 style={{ margin: '0 0 12px 0', fontSize: '14px', fontWeight: 500 }}>
                Confirmer la suppression
              </h3>
              <p style={{ margin: '0 0 16px 0', fontSize: '12px', color: 'var(--color-text-secondary)' }}>
                Cette action supprimera tous les films et séries de la base de données. Cette opération ne peut pas être annulée.
              </p>
              <div style={{ display: 'flex', gap: '8px' }}>
                <button
                  onClick={() => setShowPurgeConfirm(false)}
                  disabled={isPurging}
                  style={{
                    flex: 1,
                    padding: '8px',
                    fontSize: '11px',
                    background: 'var(--color-background-secondary)',
                    color: 'var(--color-text-primary)',
                    border: '0.5px solid var(--color-border-tertiary)',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    fontWeight: 500,
                  }}
                >
                  Annuler
                </button>
                <button
                  onClick={handlePurge}
                  disabled={isPurging}
                  style={{
                    flex: 1,
                    padding: '8px',
                    fontSize: '11px',
                    background: '#E24B4A',
                    color: 'white',
                    border: 'none',
                    borderRadius: '4px',
                    cursor: isPurging ? 'not-allowed' : 'pointer',
                    fontWeight: 500,
                    opacity: isPurging ? 0.6 : 1,
                  }}
                >
                  {isPurging ? 'Suppression...' : 'Supprimer'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
      <div className={styles.footer}>
        <div className={styles['status-dot']} />
        <span>Système opérationnel</span>
      </div>
    </div>
  );
};
