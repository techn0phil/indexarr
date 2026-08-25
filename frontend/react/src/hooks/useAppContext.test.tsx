import { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { act, render, screen, waitFor } from '../test/testUtils';
import { AppContextProvider, useAppContext } from './useAppContext';
import { apiClient } from '../api/client';

const changeLanguageMock = vi.fn();

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({
      i18n: {
        language: 'fr',
        changeLanguage: changeLanguageMock,
      },
    }),
  };
});

class MockWebSocket {
  static OPEN = 1;
  readyState = 1;
  onopen: ((ev: Event) => any) | null = null;
  onmessage: ((ev: MessageEvent) => any) | null = null;
  onclose: ((ev: CloseEvent) => any) | null = null;
  onerror: ((ev: Event) => any) | null = null;

  constructor(_url: string) {}

  close() {
    this.readyState = 3;
  }
}

const Consumer = () => {
  const ctx = useAppContext();

  return (
    <div>
      <div data-testid="auth-mode">{ctx.authMode}</div>
      <div data-testid="auth-loading">{String(ctx.authLoading)}</div>
      <div data-testid="is-authenticated">{String(ctx.isAuthenticated)}</div>
      <div data-testid="locale">{ctx.locale}</div>
      <button type="button" onClick={() => ctx.setLocale('en')}>set-locale-en</button>
      <button type="button" onClick={ctx.toggleTheme}>toggle-theme</button>
      <button type="button" onClick={() => ctx.login('admin', 'secret')}>login</button>
    </div>
  );
};

const renderWithProvider = (children: ReactNode) => {
  return render(<AppContextProvider>{children}</AppContextProvider>);
};

describe('useAppContext/AppContextProvider', () => {
  const never = new Promise<never>(() => undefined);

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    Object.defineProperty(window, 'WebSocket', { value: MockWebSocket, writable: true });

    vi.spyOn(apiClient, 'getAuthConfig').mockResolvedValue({ authMode: 'none' });
    vi.spyOn(apiClient, 'getConfig').mockImplementation(() => never as any);
    vi.spyOn(apiClient, 'getStats').mockImplementation(() => never as any);
    vi.spyOn(apiClient, 'login').mockResolvedValue({
      success: true,
      user: { id: 1, username: 'admin', role: 'admin' },
    });
  });

  it('throws when hook is used outside provider', () => {
    const BrokenConsumer = () => {
      useAppContext();
      return null;
    };

    expect(() => render(<BrokenConsumer />)).toThrow('useAppContext must be used within AppContextProvider');
  });

  it('initializes auth config and marks auth as ready', async () => {
    renderWithProvider(<Consumer />);

    await waitFor(() => {
      expect(screen.getByTestId('auth-mode')).toHaveTextContent('none');
      expect(screen.getByTestId('auth-loading')).toHaveTextContent('false');
      expect(screen.getByTestId('is-authenticated')).toHaveTextContent('true');
    });
  });

  it('updates locale and persists preference', async () => {
    renderWithProvider(<Consumer />);

    await screen.findByTestId('locale');

    act(() => {
      screen.getByRole('button', { name: 'set-locale-en' }).click();
    });

    expect(changeLanguageMock).toHaveBeenCalledWith('en');
    expect(localStorage.getItem('locale-preference')).toBe('en');
  });

  it('toggles theme and persists preference', async () => {
    renderWithProvider(<Consumer />);

    await screen.findByTestId('locale');

    act(() => {
      screen.getByRole('button', { name: 'toggle-theme' }).click();
    });

    expect(localStorage.getItem('theme-preference')).toBe('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('logs in user via context login helper', async () => {
    renderWithProvider(<Consumer />);

    await screen.findByTestId('locale');

    await act(async () => {
      screen.getByRole('button', { name: 'login' }).click();
    });

    expect(apiClient.login).toHaveBeenCalledWith('admin', 'secret');
  });
});
