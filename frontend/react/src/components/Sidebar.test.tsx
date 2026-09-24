import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { Sidebar } from './Sidebar';
import { apiClient } from '../api/client';

const appContextState = vi.hoisted(() => ({
  stats: { totalMovies: 42, totalSeries: 11 },
  authMode: 'simple' as 'none' | 'simple' | 'oidc',
  user: { id: 1, username: 'admin', role: 'admin' as 'admin' | 'guest' },
  wsConnected: true,
  wsReconnecting: false,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => appContextState,
}));

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    appContextState.authMode = 'simple';
    appContextState.user = { id: 1, username: 'admin', role: 'admin' };
    appContextState.wsConnected = true;
    appContextState.wsReconnecting = false;
  });

  it('renders movie and series nav items with counts', () => {
    render(<Sidebar activeNav="movies" onNavClick={vi.fn()} />);

    expect(screen.getByText('nav.movies')).toBeInTheDocument();
    expect(screen.getByText('nav.series')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('11')).toBeInTheDocument();
  });

  it('shows users nav for admin when auth mode is simple', () => {
    render(<Sidebar activeNav="users" onNavClick={vi.fn()} />);

    expect(screen.getByText('nav.users')).toBeInTheDocument();
  });

  it('hides users nav for non-admin when auth mode is simple', () => {
    appContextState.user = { id: 2, username: 'guest', role: 'guest' };

    render(<Sidebar activeNav="movies" onNavClick={vi.fn()} />);

    expect(screen.queryByText('nav.users')).not.toBeInTheDocument();
  });

  it('triggers purge flow and calls API', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'purgeDatabase').mockResolvedValue({ success: false } as any);

    render(<Sidebar activeNav="movies" onNavClick={vi.fn()} />);

    await user.click(screen.getByRole('button', { name: 'button.purge' }));
    await user.click(screen.getByRole('button', { name: 'popup.purge.delete' }));

    await waitFor(() => {
      expect(apiClient.purgeDatabase).toHaveBeenCalledTimes(1);
    });
  });

  it('shows reconnecting status when websocket is reconnecting', () => {
    appContextState.wsConnected = false;
    appContextState.wsReconnecting = true;

    render(<Sidebar activeNav="movies" onNavClick={vi.fn()} />);

    expect(screen.getByText('status.reconnecting')).toBeInTheDocument();
  });
});
