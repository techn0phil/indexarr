import { createContext, useState, useEffect, useContext, ReactNode } from 'react';
import { apiClient } from '../api/client';
import { StatsResponse } from '../types';

export type Page = 'list-films' | 'list-series' | 'detail-movie' | 'detail-series';

interface AppConfig {
  radarrUrl: string;
}

interface AppContextType {
  currentPage: Page;
  selectedId: number | null;
  goToPage: (page: Page, id?: number) => void;
  goBack: () => void;
  history: Page[];
  isDark: boolean;
  toggleTheme: () => void;
  config: AppConfig | null;
  configLoading: boolean;
  stats: StatsResponse | null;
  statsLoading: boolean;
  refreshStats: () => Promise<void>;
}

export const AppContext = createContext<AppContextType | undefined>(undefined);

interface AppContextProviderProps {
  children: ReactNode;
}

export const AppContextProvider = ({ children }: AppContextProviderProps) => {
  const [currentPage, setCurrentPage] = useState<Page>('list-films');
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [history, setHistory] = useState<Page[]>(['list-films']);
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [stats, setStats] = useState<StatsResponse | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  // Fetch config on mount
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        const data = await apiClient.getConfig();
        setConfig(data);
      } catch (error) {
        console.error('Failed to fetch config:', error);
        // Fallback to default config
        setConfig({ radarrUrl: '' });
      } finally {
        setConfigLoading(false);
      }
    };

    fetchConfig();
  }, []);

  // Fetch stats on mount
  const refreshStats = async () => {
    setStatsLoading(true);
    try {
      const data = await apiClient.getStats();
      setStats(data);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
      setStats(null);
    } finally {
      setStatsLoading(false);
    }
  };

  useEffect(() => {
    refreshStats();
  }, []);

  // Initialize theme from localStorage or system preference
  const [isDark, setIsDark] = useState(() => {
    const saved = localStorage.getItem('theme-preference');
    if (saved === 'dark' || saved === 'light') {
      return saved === 'dark';
    }
    // Fallback to system preference
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  });

  // Apply theme on mount and when isDark changes
  useEffect(() => {
    const theme = isDark ? 'dark' : 'light';
    document.documentElement.setAttribute('data-theme', theme);
    document.documentElement.style.colorScheme = theme;
  }, [isDark]);

  // Listen to system preference changes only if user hasn't set a manual preference
  useEffect(() => {
    const saved = localStorage.getItem('theme-preference');
    if (saved) return; // User has manual preference, don't listen to system

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
      setIsDark(e.matches);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  const goToPage = (page: Page, id?: number) => {
    setCurrentPage(page);
    if (id) setSelectedId(id);
    setHistory([...history, page]);
  };

  const goBack = () => {
    if (history.length > 1) {
      const newHistory = history.slice(0, -1);
      setHistory(newHistory);
      setCurrentPage(newHistory[newHistory.length - 1]);
    }
  };

  const toggleTheme = () => {
    const newTheme = !isDark;
    setIsDark(newTheme);
    localStorage.setItem('theme-preference', newTheme ? 'dark' : 'light');
  };

  return (
    <AppContext.Provider value={{ currentPage, selectedId, goToPage, goBack, history, isDark, toggleTheme, config, configLoading, stats, statsLoading, refreshStats }}>
      {children}
    </AppContext.Provider>
  );
};

export const useAppContext = () => {
  const context = useContext(AppContext);
  if (!context) {
    throw new Error('useAppContext must be used within AppContextProvider');
  }
  return context;
};
