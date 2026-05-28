import { useState, useRef, useEffect } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import { apiClient } from '../api/client';
import styles from '../styles/usermenu.module.css';

export const UserMenu = () => {
  const { user, authMode, logout } = useAppContext();
  const [isOpen, setIsOpen] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordSuccess, setPasswordSuccess] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Don't render if no auth or no user
  if (authMode === 'disabled' || !user) {
    return null;
  }

  // Check if user can change password (not env admin)
  const canChangePassword = authMode === 'local' && user.id !== undefined && user.id !== 0;

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

  const openPasswordModal = () => {
    setIsOpen(false);
    setCurrentPassword('');
    setNewPassword('');
    setConfirmPassword('');
    setPasswordError(null);
    setPasswordSuccess(false);
    setShowPasswordModal(true);
  };

  const handleChangePassword = async () => {
    if (!currentPassword || !newPassword || !confirmPassword) {
      setPasswordError('Tous les champs sont requis');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('Les mots de passe ne correspondent pas');
      return;
    }

    if (newPassword.length < 4) {
      setPasswordError('Le mot de passe doit contenir au moins 4 caractères');
      return;
    }

    setPasswordLoading(true);
    setPasswordError(null);

    try {
      const response = await apiClient.changePassword(currentPassword, newPassword);
      if (response.success) {
        setPasswordSuccess(true);
        setTimeout(() => {
          setShowPasswordModal(false);
        }, 1500);
      } else {
        setPasswordError(response.error || 'Erreur lors du changement de mot de passe');
      }
    } catch (err) {
      setPasswordError('Erreur de connexion');
    } finally {
      setPasswordLoading(false);
    }
  };

  return (
    <>
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
            {canChangePassword && (
              <button className={styles.menuItem} onClick={openPasswordModal}>
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.menuIcon}>
                  <rect x="3" y="7" width="10" height="7" rx="1" />
                  <path d="M5 7V5a3 3 0 016 0v2" />
                </svg>
                Modifier le mot de passe
              </button>
            )}
            <button className={styles.menuItem} onClick={handleLogout}>
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.menuIcon}>
                <path d="M6 2H3a1 1 0 00-1 1v10a1 1 0 001 1h3M11 11l3-3-3-3M14 8H6" />
              </svg>
              Déconnexion
            </button>
          </div>
        )}
      </div>

      {/* Change Password Modal */}
      {showPasswordModal && (
        <div className={styles.modalOverlay} onClick={() => !passwordLoading && setShowPasswordModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Modifier le mot de passe</h2>

            {passwordSuccess ? (
              <div className={styles.successMessage}>
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.successIcon}>
                  <circle cx="8" cy="8" r="6" />
                  <path d="M5 8l2 2 4-4" />
                </svg>
                Mot de passe modifié avec succès
              </div>
            ) : (
              <>
                <div className={styles.formField}>
                  <label className={styles.label}>Mot de passe actuel</label>
                  <input
                    type="password"
                    className={styles.input}
                    value={currentPassword}
                    onChange={(e) => setCurrentPassword(e.target.value)}
                    placeholder="••••••••"
                    autoFocus
                  />
                </div>

                <div className={styles.formField}>
                  <label className={styles.label}>Nouveau mot de passe</label>
                  <input
                    type="password"
                    className={styles.input}
                    value={newPassword}
                    onChange={(e) => setNewPassword(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>

                <div className={styles.formField}>
                  <label className={styles.label}>Confirmer le mot de passe</label>
                  <input
                    type="password"
                    className={styles.input}
                    value={confirmPassword}
                    onChange={(e) => setConfirmPassword(e.target.value)}
                    placeholder="••••••••"
                  />
                </div>

                {passwordError && <div className={styles.error}>{passwordError}</div>}

                <div className={styles.modalActions}>
                  <button
                    className={styles.cancelButton}
                    onClick={() => setShowPasswordModal(false)}
                    disabled={passwordLoading}
                  >
                    Annuler
                  </button>
                  <button
                    className={styles.submitButton}
                    onClick={handleChangePassword}
                    disabled={passwordLoading}
                  >
                    {passwordLoading ? 'Modification...' : 'Modifier'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </>
  );
};
