import { useState, useEffect } from 'react';
import { apiClient } from '../api/client';
import { UserDetails, CreateUserRequest } from '../types';
import styles from '../styles/users.module.css';

export const UsersPage = () => {
  const [users, setUsers] = useState<UserDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Modal states
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserDetails | null>(null);
  
  // Form states
  const [formUsername, setFormUsername] = useState('');
  const [formPassword, setFormPassword] = useState('');
  const [formRole, setFormRole] = useState<'admin' | 'guest'>('guest');
  const [formEnabled, setFormEnabled] = useState(true);
  const [formLoading, setFormLoading] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const fetchUsers = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await apiClient.getUsers();
      if (response.success && response.data) {
        setUsers(response.data);
      } else {
        setError(response.error || 'Erreur lors du chargement des utilisateurs');
      }
    } catch (err) {
      setError('Erreur de connexion');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const openCreateModal = () => {
    setFormUsername('');
    setFormPassword('');
    setFormRole('guest');
    setFormError(null);
    setShowCreateModal(true);
  };

  const openEditModal = (user: UserDetails) => {
    setSelectedUser(user);
    setFormUsername(user.username);
    setFormRole(user.role);
    setFormEnabled(user.enabled);
    setFormError(null);
    setShowEditModal(true);
  };

  const openDeleteModal = (user: UserDetails) => {
    setSelectedUser(user);
    setShowDeleteModal(true);
  };

  const openPasswordModal = (user: UserDetails) => {
    setSelectedUser(user);
    setFormPassword('');
    setFormError(null);
    setShowPasswordModal(true);
  };

  const handleCreate = async () => {
    if (!formUsername || !formPassword) {
      setFormError('Tous les champs sont requis');
      return;
    }
    
    setFormLoading(true);
    setFormError(null);
    
    try {
      const data: CreateUserRequest = {
        username: formUsername,
        password: formPassword,
        role: formRole,
      };
      const response = await apiClient.createUser(data);
      if (response.success) {
        setShowCreateModal(false);
        fetchUsers();
      } else {
        setFormError(response.error || 'Erreur lors de la création');
      }
    } catch (err) {
      setFormError('Erreur de connexion');
    } finally {
      setFormLoading(false);
    }
  };

  const handleUpdate = async () => {
    if (!selectedUser) return;
    
    setFormLoading(true);
    setFormError(null);
    
    try {
      const response = await apiClient.updateUser(selectedUser.id, {
        username: formUsername !== selectedUser.username ? formUsername : undefined,
        role: formRole !== selectedUser.role ? formRole : undefined,
        enabled: formEnabled !== selectedUser.enabled ? formEnabled : undefined,
      });
      if (response.success) {
        setShowEditModal(false);
        fetchUsers();
      } else {
        setFormError(response.error || 'Erreur lors de la mise à jour');
      }
    } catch (err) {
      setFormError('Erreur de connexion');
    } finally {
      setFormLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedUser) return;
    
    setFormLoading(true);
    
    try {
      const response = await apiClient.deleteUser(selectedUser.id);
      if (response.success) {
        setShowDeleteModal(false);
        fetchUsers();
      } else {
        setFormError(response.error || 'Erreur lors de la suppression');
      }
    } catch (err) {
      setFormError('Erreur de connexion');
    } finally {
      setFormLoading(false);
    }
  };

  const handleSetPassword = async () => {
    if (!selectedUser || !formPassword) {
      setFormError('Le mot de passe est requis');
      return;
    }
    
    setFormLoading(true);
    setFormError(null);
    
    try {
      const response = await apiClient.adminSetPassword(selectedUser.id, formPassword);
      if (response.success) {
        setShowPasswordModal(false);
      } else {
        setFormError(response.error || 'Erreur lors de la modification');
      }
    } catch (err) {
      setFormError('Erreur de connexion');
    } finally {
      setFormLoading(false);
    }
  };

  const toggleUserEnabled = async (user: UserDetails) => {
    try {
      const response = await apiClient.updateUser(user.id, { enabled: !user.enabled });
      if (response.success) {
        fetchUsers();
      }
    } catch (err) {
      console.error('Failed to toggle user:', err);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('fr-FR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className={styles.container}>
        <div className={styles.loading}>Chargement...</div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Gestion des utilisateurs</h1>
        <button className={styles.createButton} onClick={openCreateModal}>
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.buttonIcon}>
            <path d="M8 3v10M3 8h10" />
          </svg>
          Nouvel utilisateur
        </button>
      </div>

      {error && <div className={styles.error}>{error}</div>}

      {users.length === 0 ? (
        <div className={styles.empty}>
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.emptyIcon}>
            <circle cx="6" cy="5" r="2.5" />
            <path d="M1 13c0-2.5 2-4 5-4s5 1.5 5 4" />
            <circle cx="12" cy="6" r="2" />
            <path d="M15 13c0-2 -1.5-3-3.5-3" />
          </svg>
          <p>Aucun utilisateur créé</p>
          <p className={styles.emptySubtext}>Créez votre premier utilisateur pour commencer</p>
        </div>
      ) : (
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            <div className={styles.colUsername}>Nom d'utilisateur</div>
            <div className={styles.colRole}>Rôle</div>
            <div className={styles.colStatus}>Statut</div>
            <div className={styles.colDate}>Créé le</div>
            <div className={styles.colActions}>Actions</div>
          </div>
          {users.map((user) => (
            <div key={user.id} className={styles.tableRow}>
              <div className={styles.colUsername}>
                <span className={styles.username}>{user.username}</span>
              </div>
              <div className={styles.colRole}>
                <span className={`${styles.roleBadge} ${user.role === 'admin' ? styles.roleAdmin : styles.roleGuest}`}>
                  {user.role === 'admin' ? 'Administrateur' : 'Invité'}
                </span>
              </div>
              <div className={styles.colStatus}>
                <button
                  className={`${styles.statusToggle} ${user.enabled ? styles.enabled : styles.disabled}`}
                  onClick={() => toggleUserEnabled(user)}
                  title={user.enabled ? 'Cliquer pour désactiver' : 'Cliquer pour activer'}
                >
                  <span className={styles.statusDot} />
                  {user.enabled ? 'Actif' : 'Désactivé'}
                </button>
              </div>
              <div className={styles.colDate}>{formatDate(user.createdAt)}</div>
              <div className={styles.colActions}>
                <button className={styles.actionButton} onClick={() => openEditModal(user)} title="Modifier">
                  <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M11.5 2.5l2 2-8 8H3.5v-2l8-8z" />
                  </svg>
                </button>
                <button className={styles.actionButton} onClick={() => openPasswordModal(user)} title="Changer le mot de passe">
                  <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <rect x="3" y="7" width="10" height="7" rx="1" />
                    <path d="M5 7V5a3 3 0 016 0v2" />
                  </svg>
                </button>
                <button className={`${styles.actionButton} ${styles.deleteButton}`} onClick={() => openDeleteModal(user)} title="Supprimer">
                  <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                    <path d="M3 4h10M5.5 4V3a1 1 0 011-1h3a1 1 0 011 1v1M6 7v5M10 7v5M4 4l.8 9a1 1 0 001 .9h4.4a1 1 0 001-.9l.8-9" />
                  </svg>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <div className={styles.modalOverlay} onClick={() => setShowCreateModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Nouvel utilisateur</h2>
            
            <div className={styles.formField}>
              <label className={styles.label}>Nom d'utilisateur</label>
              <input
                type="text"
                className={styles.input}
                value={formUsername}
                onChange={(e) => setFormUsername(e.target.value)}
                placeholder="john.doe"
                autoFocus
              />
            </div>
            
            <div className={styles.formField}>
              <label className={styles.label}>Mot de passe</label>
              <input
                type="password"
                className={styles.input}
                value={formPassword}
                onChange={(e) => setFormPassword(e.target.value)}
                placeholder="••••••••"
              />
            </div>
            
            <div className={styles.formField}>
              <label className={styles.label}>Rôle</label>
              <select
                className={styles.select}
                value={formRole}
                onChange={(e) => setFormRole(e.target.value as 'admin' | 'guest')}
              >
                <option value="guest">Invité</option>
                <option value="admin">Administrateur</option>
              </select>
            </div>
            
            {formError && <div className={styles.formError}>{formError}</div>}
            
            <div className={styles.modalActions}>
              <button className={styles.cancelButton} onClick={() => setShowCreateModal(false)} disabled={formLoading}>
                Annuler
              </button>
              <button className={styles.submitButton} onClick={handleCreate} disabled={formLoading}>
                {formLoading ? 'Création...' : 'Créer'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {showEditModal && selectedUser && (
        <div className={styles.modalOverlay} onClick={() => setShowEditModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Modifier l'utilisateur</h2>
            
            <div className={styles.formField}>
              <label className={styles.label}>Nom d'utilisateur</label>
              <input
                type="text"
                className={styles.input}
                value={formUsername}
                onChange={(e) => setFormUsername(e.target.value)}
                autoFocus
              />
            </div>
            
            <div className={styles.formField}>
              <label className={styles.label}>Rôle</label>
              <select
                className={styles.select}
                value={formRole}
                onChange={(e) => setFormRole(e.target.value as 'admin' | 'guest')}
              >
                <option value="guest">Invité</option>
                <option value="admin">Administrateur</option>
              </select>
            </div>
            
            <div className={styles.formField}>
              <label className={styles.checkboxLabel}>
                <input
                  type="checkbox"
                  checked={formEnabled}
                  onChange={(e) => setFormEnabled(e.target.checked)}
                />
                Compte actif
              </label>
            </div>
            
            {formError && <div className={styles.formError}>{formError}</div>}
            
            <div className={styles.modalActions}>
              <button className={styles.cancelButton} onClick={() => setShowEditModal(false)} disabled={formLoading}>
                Annuler
              </button>
              <button className={styles.submitButton} onClick={handleUpdate} disabled={formLoading}>
                {formLoading ? 'Enregistrement...' : 'Enregistrer'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Modal */}
      {showDeleteModal && selectedUser && (
        <div className={styles.modalOverlay} onClick={() => setShowDeleteModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Supprimer l'utilisateur</h2>
            
            <p className={styles.modalText}>
              Êtes-vous sûr de vouloir supprimer l'utilisateur <strong>{selectedUser.username}</strong> ?
              Cette action est irréversible.
            </p>
            
            {formError && <div className={styles.formError}>{formError}</div>}
            
            <div className={styles.modalActions}>
              <button className={styles.cancelButton} onClick={() => setShowDeleteModal(false)} disabled={formLoading}>
                Annuler
              </button>
              <button className={styles.deleteSubmitButton} onClick={handleDelete} disabled={formLoading}>
                {formLoading ? 'Suppression...' : 'Supprimer'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Password Modal */}
      {showPasswordModal && selectedUser && (
        <div className={styles.modalOverlay} onClick={() => setShowPasswordModal(false)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <h2 className={styles.modalTitle}>Changer le mot de passe</h2>
            
            <p className={styles.modalSubtext}>
              Définir un nouveau mot de passe pour <strong>{selectedUser.username}</strong>
            </p>
            
            <div className={styles.formField}>
              <label className={styles.label}>Nouveau mot de passe</label>
              <input
                type="password"
                className={styles.input}
                value={formPassword}
                onChange={(e) => setFormPassword(e.target.value)}
                placeholder="••••••••"
                autoFocus
              />
            </div>
            
            {formError && <div className={styles.formError}>{formError}</div>}
            
            <div className={styles.modalActions}>
              <button className={styles.cancelButton} onClick={() => setShowPasswordModal(false)} disabled={formLoading}>
                Annuler
              </button>
              <button className={styles.submitButton} onClick={handleSetPassword} disabled={formLoading}>
                {formLoading ? 'Enregistrement...' : 'Enregistrer'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
