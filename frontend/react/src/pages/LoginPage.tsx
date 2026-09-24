import { useState, FormEvent } from 'react';
import { useAppContext } from '../hooks/useAppContext';
import { ThemeToggle } from '../components/ThemeToggle';
import styles from '../styles/login.module.css';
import { LanguageToggle } from '../components/LanguageToggle';
import { useTranslation } from 'react-i18next';

export const LoginPage = () => {
  const { t } = useTranslation('login');
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
      setError(`errors.${result.error}` || 'errors.invalidCredentials');
    }
    
    setLoading(false);
  };

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles['logo-container']}>
            <div className={styles.logo}>
              <div className={styles.glass}></div>
              <div className={styles.handle}></div>
              <div className={styles.index}>
                <div className={styles.item}></div>
                <div className={styles.item}></div>
                <div className={styles.item}></div>
              </div>
            </div>

            <span className={styles['logo-name']}>Index<span style={{ color: "#1d9e75" }}>arr</span></span>
          </div>
          <div className={styles.themeToggle}>
            <ThemeToggle />
            <LanguageToggle />
          </div>
        </div>

        <h1 className={styles.title}>{t('title')}</h1>

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="username" className={styles.label}>
              {t('fields.username')}
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
              {t('fields.password')}
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
              {t(error)}
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
              t('buttons.submit')
            )}
          </button>
        </form>
      </div>
    </div>
  );
};
