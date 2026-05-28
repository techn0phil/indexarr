import { useState, FormEvent } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import { ThemeToggle } from '../components/ThemeToggle';
import styles from '../styles/login.module.css';

export const LoginPage = () => {
  const { login } = useAppContext();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

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

          {error && (
            <div className={styles.error}>
              <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" className={styles.errorIcon}>
                <circle cx="8" cy="8" r="6" />
                <path d="M8 5v3M8 10v1" />
              </svg>
              {error}
            </div>
          )}

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
      </div>
    </div>
  );
};
