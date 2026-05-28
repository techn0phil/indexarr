import { useState, FormEvent, useEffect } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import { apiClient } from '../api/client';
import { ThemeToggle } from '../components/ThemeToggle';
import styles from '../styles/login.module.css';

export const LoginPage = () => {
  const { login, authMode } = useAppContext();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [oidcLoading, setOidcLoading] = useState(false);

  // Check for OIDC error in URL on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const authError = params.get('auth_error');
    if (authError) {
      setError(decodeURIComponent(authError));
      // Clean up URL
      window.history.replaceState({}, '', window.location.pathname);
    }
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    const result = await login(username, password);
    
    if (!result.success) {
      setError(result.error || 'Identifiants invalides');
    }
    
    setLoading(false);
  };

  const handleOIDCLogin = async () => {
    setError(null);
    setOidcLoading(true);

    try {
      const result = await apiClient.getOIDCLoginUrl();
      if (result.success && result.authUrl) {
        // Redirect to OIDC provider
        window.location.href = result.authUrl;
      } else {
        setError(result.error || 'Erreur lors de la connexion SSO');
        setOidcLoading(false);
      }
    } catch (err) {
      setError('Erreur de connexion au serveur');
      setOidcLoading(false);
    }
  };

  const isOIDCMode = authMode === 'oidc';

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles.logo}>
            <div className={styles.logoMark}>
              <svg viewBox="0 0 14 14" style={{ width: '13px', height: '13px', fill: 'white' }}>
                <path d="M2 11L7 3L12 11Z" />
              </svg>
            </div>
            <span className={styles.logoText}>Index<span style={{ color: "#1d9e75" }}>arr</span></span>
          </div>
          <div className={styles.themeToggle}>
            <ThemeToggle />
          </div>
        </div>

        <h1 className={styles.title}>Connexion</h1>

        {error && (
          <div className={styles.error}>
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.errorIcon}>
              <circle cx="8" cy="8" r="6" />
              <path d="M8 5v3M8 10v1" />
            </svg>
            {error}
          </div>
        )}

        {isOIDCMode ? (
          // OIDC Login Mode
          <div className={styles.oidcContainer}>
            <p className={styles.oidcText}>
              Utilisez votre compte d'entreprise pour vous connecter.
            </p>
            <button 
              type="button"
              className={styles.oidcButton}
              onClick={handleOIDCLogin}
              disabled={oidcLoading}
            >
              {oidcLoading ? (
                <span className={styles.spinner}></span>
              ) : (
                <>
                  <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.oidcIcon}>
                    <rect x="2" y="4" width="12" height="8" rx="1" />
                    <path d="M2 6l6 4 6-4" />
                  </svg>
                  Se connecter avec SSO
                </>
              )}
            </button>
          </div>
        ) : (
          // Simple Auth Mode (username/password)
          <form onSubmit={handleSubmit} className={styles.form}>
            <div className={styles.field}>
              <label htmlFor="username" className={styles.label}>
                Nom d'utilisateur
              </label>
              <input
                type="text"
                id="username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                className={styles.input}
                placeholder="admin"
                required
                autoComplete="username"
                autoFocus
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="password" className={styles.label}>
                Mot de passe
              </label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className={styles.input}
                placeholder="••••••••"
                required
                autoComplete="current-password"
              />
            </div>

            <button 
              type="submit" 
              className={styles.button}
              disabled={loading || !username || !password}
            >
              {loading ? (
                <span className={styles.spinner}></span>
              ) : (
                'Se connecter'
              )}
            </button>
          </form>
        )}
      </div>
    </div>
  );
};
