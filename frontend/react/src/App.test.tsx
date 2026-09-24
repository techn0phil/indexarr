import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from './test/testUtils';
import App from './App';

const appContextState = vi.hoisted(() => ({
  authMode: 'none',
  user: null,
  isAuthenticated: true,
  authLoading: false,
  login: vi.fn(),
  logout: vi.fn(),
  isDark: false,
  toggleTheme: vi.fn(),
  locale: 'fr',
  availableLanguages: ['fr'],
  setLocale: vi.fn(),
  config: null,
  configLoading: false,
  stats: null,
  statsLoading: false,
  refreshStats: vi.fn(),
  scanStatus: null,
  wsConnected: false,
  wsReconnecting: false,
}));

vi.mock('./hooks/useAppContext.tsx', () => {
  const AppContext = React.createContext(appContextState as any);

  return {
    AppContext,
    AppContextProvider: ({ children }: { children: React.ReactNode }) => (
      <AppContext.Provider value={appContextState as any}>{children}</AppContext.Provider>
    ),
  };
});

vi.mock('./components/Sidebar', () => ({
  Sidebar: () => <div>sidebar</div>,
}));

vi.mock('./components/Topbar', () => ({
  Topbar: () => <div>topbar</div>,
}));

vi.mock('./pages/ListFilms', () => ({
  ListFilms: () => <div>movies-page</div>,
}));

vi.mock('./pages/ListSeries', () => ({
  ListSeries: () => <div>series-page</div>,
}));

vi.mock('./pages/MovieDetail', () => ({
  MovieDetail: () => <div>movie-detail-page</div>,
}));

vi.mock('./pages/SeriesDetail', () => ({
  SeriesDetail: () => <div>series-detail-page</div>,
}));

vi.mock('./pages/UsersPage', () => ({
  UsersPage: () => <div>users-page</div>,
}));

vi.mock('./pages/LoginPage', () => ({
  LoginPage: () => <div>login-page</div>,
}));

describe('App auth and routing gates', () => {
  beforeEach(() => {
    appContextState.authMode = 'none';
    appContextState.isAuthenticated = true;
    appContextState.authLoading = false;
    window.history.pushState({}, '', '/');
  });

  it('shows login page when simple auth is enabled and user is not authenticated', () => {
    appContextState.authMode = 'simple';
    appContextState.isAuthenticated = false;

    render(<App />);

    expect(screen.getByText('login-page')).toBeInTheDocument();
    expect(screen.queryByText('movies-page')).not.toBeInTheDocument();
  });

  it('renders app content when auth mode is none', () => {
    render(<App />);

    expect(screen.getByText('movies-page')).toBeInTheDocument();
    expect(screen.queryByText('login-page')).not.toBeInTheDocument();
  });

  it('redirects unknown route to movies', () => {
    window.history.pushState({}, '', '/unexpected-path');

    render(<App />);

    expect(screen.getByText('movies-page')).toBeInTheDocument();
  });

  it('shows auth loading gate before login or app content', () => {
    appContextState.authLoading = true;

    render(<App />);

    expect(screen.queryByText('login-page')).not.toBeInTheDocument();
    expect(screen.queryByText('movies-page')).not.toBeInTheDocument();
  });
});
