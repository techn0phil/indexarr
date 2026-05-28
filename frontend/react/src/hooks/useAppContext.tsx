import { createContext, useState, useEffect, useContext, ReactNode, useRef, useCallback } from 'react';
import { apiClient } from '../api/client';
import { StatsResponse, ScanStatus, AuthMode, User } from '../types';

export type Page = 'list-films' | 'list-series' | 'detail-movie' | 'detail-series' | 'admin-users';

interface AppConfig {
  radarrUrl?: string;
  sonarrUrl?: string;
}

interface WSMessage {
  type: 'scan_start' | 'scan_progress' | 'scan_complete' | 'scan_error' | 'scan_stopped' | 'scan_idle';
  filesFound?: number;
  filesProcessed?: number;
  startedAt?: string;
  error?: string;
  moviesAdded?: number;
  episodesAdded?: number;
}

interface AppContextType {
  // Auth state
  authMode: AuthMode;
  user: User | null;
  isAuthenticated: boolean;
  authLoading: boolean;
  login: (username: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  
  // Navigation state
  currentPage: Page;
  selectedId: number | null;
  goToPage: (page: Page, id?: number) => void;
  goBack: () => void;
  history: Page[];
  
  // Theme state
  isDark: boolean;
  toggleTheme: () => void;
  
  // App data state
  config: AppConfig | null;
  configLoading: boolean;
  stats: StatsResponse | null;
  statsLoading: boolean;
  refreshStats: () => Promise<void>;
  scanStatus: ScanStatus | null;
  wsConnected: boolean;
  wsReconnecting: boolean;
}

export const AppContext = createContext<AppContextType | undefined>(undefined);

interface AppContextProviderProps {
  children: ReactNode;
}

export const AppContextProvider = ({ children }: AppContextProviderProps) => {
  // Auth state
  const [authMode, setAuthMode] = useState<AuthMode>('none');
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [authLoading, setAuthLoading] = useState(true);
  
  // Navigation state
  const [currentPage, setCurrentPage] = useState<Page>('list-films');
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [history, setHistory] = useState<Page[]>(['list-films']);
  
  // App data state
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [stats, setStats] = useState<StatsResponse | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);
  
  // WebSocket state
  const [scanStatus, setScanStatus] = useState<ScanStatus | null>(null);
  const [wsConnected, setWsConnected] = useState(false);
  const [wsReconnecting, setWsReconnecting] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const unmountedRef = useRef(false);

  // WebSocket URL generator
  const getWebSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    return `${protocol}//${host}/api/scan/ws`;
  }, []);

  // Update scan status from WebSocket message
  const updateStatusFromMessage = useCallback((msg: WSMessage) => {
    setScanStatus((prevStatus) => {
      const newStatus: ScanStatus = {
        id: prevStatus?.id || 1,
        status: 'idle',
        filesFound: msg.filesFound || prevStatus?.filesFound || 0,
        filesProcessed: msg.filesProcessed || prevStatus?.filesProcessed || 0,
        startedAt: msg.startedAt || prevStatus?.startedAt,
        completedAt: prevStatus?.completedAt,
        errorMessage: msg.error || prevStatus?.errorMessage,
      };

      switch (msg.type) {
        case 'scan_start':
          newStatus.status = 'running';
          newStatus.startedAt = msg.startedAt;
          newStatus.filesFound = 0;
          newStatus.filesProcessed = 0;
          newStatus.completedAt = undefined;
          newStatus.errorMessage = undefined;
          break;
        case 'scan_progress':
          newStatus.status = 'running';
          break;
        case 'scan_complete':
          newStatus.status = 'completed';
          newStatus.completedAt = new Date().toISOString();
          break;
        case 'scan_error':
          newStatus.status = 'error';
          newStatus.errorMessage = msg.error;
          newStatus.completedAt = new Date().toISOString();
          break;
        case 'scan_stopped':
          newStatus.status = 'stopped';
          newStatus.completedAt = new Date().toISOString();
          break;
        case 'scan_idle':
          newStatus.status = msg.filesProcessed ? 'completed' : 'idle';
          break;
      }

      return newStatus;
    });
  }, []);

  // Connect to WebSocket
  const connect = useCallback(() => {
    if (unmountedRef.current) {
      console.log('[WS] Skipping connection: component unmounted');
      return;
    }

    // Check if we already have an open connection (prevents StrictMode duplicates)
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('[WS] Connection already open, skipping duplicate');
      return;
    }

    const url = getWebSocketUrl();
    console.log('[WS] Creating connection at', new Date().toISOString(), 'to', url);

    try {
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('[WS] Connected successfully');
        setWsConnected(true);
        setWsReconnecting(false);
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const msg: WSMessage = JSON.parse(event.data);
          console.log('[WS] Message received:', msg.type);
          updateStatusFromMessage(msg);
        } catch (error) {
          console.error('[WS] Failed to parse message:', event.data, error);
        }
      };

      ws.onerror = (error) => {
        console.error('[WS] Error:', error);
      };

      ws.onclose = (event) => {
        console.log('[WS] Closed:', event.code, event.reason || '(no reason)');
        setWsConnected(false);
        wsRef.current = null;

        // Reconnect with exponential backoff if not unmounted and authenticated
        if (!unmountedRef.current && isAuthenticated) {
          setWsReconnecting(true);
          reconnectAttemptsRef.current++;
          const backoffTime = Math.min(
            1000 * Math.pow(2, reconnectAttemptsRef.current - 1),
            10000
          );
          console.log(`[WS] Reconnecting in ${backoffTime}ms (attempt ${reconnectAttemptsRef.current})`);

          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, backoffTime) as unknown as number;
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket:', error);
    }
  }, [getWebSocketUrl, updateStatusFromMessage, isAuthenticated]);

  // Initialize auth on mount - fetch auth config first (public endpoint)
  useEffect(() => {
    const initAuth = async () => {
      try {
        const authConfig = await apiClient.getAuthConfig();
        setAuthMode(authConfig.authMode);
        
        if (authConfig.authMode === 'none') {
          // No auth required, mark as authenticated
          setIsAuthenticated(true);
          setAuthLoading(false);
        } else if (authConfig.authMode === 'simple' || authConfig.authMode === 'oidc') {
          // Check if we have a valid session (both simple and OIDC use the same session mechanism)
          try {
            const response = await apiClient.getCurrentUser();
            if (response.success && response.user) {
              setUser(response.user);
              setIsAuthenticated(true);
            }
          } catch {
            // Not authenticated, will show login
            setIsAuthenticated(false);
          }
          setAuthLoading(false);
        } else {
          // Unknown auth mode, fallback to no auth
          setAuthMode('none');
          setIsAuthenticated(true);
          setAuthLoading(false);
        }
      } catch (error) {
        console.error('Failed to fetch auth config:', error);
        // Fallback to no auth mode
        setAuthMode('none');
        setIsAuthenticated(true);
        setAuthLoading(false);
      }
    };

    initAuth();
  }, []);

  // Load protected data when authenticated
  useEffect(() => {
    if (!isAuthenticated) return;

    // Connect to WebSocket
    console.log('[WS] Initializing connection...');
    unmountedRef.current = false;
    connect();

    // Fetch config
    const fetchConfig = async () => {
      try {
        const data = await apiClient.getConfig();
        setConfig(data);
      } catch (error) {
        console.error('Failed to fetch config:', error);
        setConfig({ radarrUrl: '', sonarrUrl: '' });
      } finally {
        setConfigLoading(false);
      }
    };
    fetchConfig();

    // Fetch stats
    refreshStats();

    return () => {
      console.log('[WS] Cleanup: Closing connection');
      unmountedRef.current = true;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, 'Component unmounting');
        wsRef.current = null;
      }
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isAuthenticated, connect]);

  // Fetch stats helper
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

  // Auth functions
  const login = async (username: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const response = await apiClient.login(username, password);
      if (response.success && response.user) {
        setUser(response.user);
        setIsAuthenticated(true);
        return { success: true };
      }
      return { success: false, error: response.error || 'Identifiants invalides' };
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, error: 'Erreur de connexion' };
    }
  };

  const logout = async (): Promise<void> => {
    try {
      await apiClient.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      setIsAuthenticated(false);
      // Reset app state
      setConfig(null);
      setConfigLoading(true);
      setStats(null);
      setStatsLoading(true);
      setScanStatus(null);
      // Close WebSocket
      if (wsRef.current) {
        wsRef.current.close(1000, 'User logged out');
        wsRef.current = null;
      }
      setWsConnected(false);
    }
  };

  return (
    <AppContext.Provider value={{ 
      // Auth
      authMode, user, isAuthenticated, authLoading, login, logout,
      // Navigation
      currentPage, selectedId, goToPage, goBack, history, 
      // Theme
      isDark, toggleTheme, 
      // App data
      config, configLoading, stats, statsLoading, refreshStats, scanStatus, wsConnected, wsReconnecting 
    }}>
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
