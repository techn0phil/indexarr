import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen } from '../test/testUtils';
import { ThemeToggle } from './ThemeToggle';
import { AppContext } from '../hooks/useAppContext';

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({
      t: (key: string) => key,
    }),
  };
});

describe('ThemeToggle', () => {
  it('renders nothing without context', () => {
    const { container } = render(<ThemeToggle />);
    expect(container.firstChild).toBeNull();
  });

  it('renders moon state and toggles theme', async () => {
    const user = userEvent.setup();
    const toggleTheme = vi.fn();

    render(
      <AppContext.Provider
        value={{
          authMode: 'none',
          user: null,
          isAuthenticated: true,
          authLoading: false,
          login: vi.fn(),
          logout: vi.fn(),
          isDark: false,
          toggleTheme,
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
        }}
      >
        <ThemeToggle />
      </AppContext.Provider>
    );

    const button = screen.getByRole('button', { name: 'button.toggleTheme' });
    expect(button).toHaveAttribute('title', 'button.darkModeSwitch');

    await user.click(button);
    expect(toggleTheme).toHaveBeenCalledTimes(1);
  });

  it('renders sun state when dark mode is active', () => {
    render(
      <AppContext.Provider
        value={{
          authMode: 'none',
          user: null,
          isAuthenticated: true,
          authLoading: false,
          login: vi.fn(),
          logout: vi.fn(),
          isDark: true,
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
        }}
      >
        <ThemeToggle />
      </AppContext.Provider>
    );

    expect(screen.getByRole('button', { name: 'button.toggleTheme' })).toHaveAttribute('title', 'button.lightModeSwitch');
  });
});
