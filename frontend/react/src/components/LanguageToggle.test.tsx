import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { fireEvent, render, screen } from '../test/testUtils';
import { LanguageToggle } from './LanguageToggle';

const setLocaleMock = vi.hoisted(() => vi.fn());
const appContextState = vi.hoisted(() => ({
  locale: 'fr',
  setLocale: setLocaleMock,
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => appContextState,
}));

describe('LanguageToggle', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    appContextState.locale = 'fr';
  });

  it('renders current locale flag', () => {
    render(<LanguageToggle />);

    expect(screen.getByRole('button', { name: 'Select language' })).toHaveTextContent('🇫🇷');
  });

  it('opens dropdown and selects language', async () => {
    const user = userEvent.setup();
    render(<LanguageToggle />);

    await user.click(screen.getByRole('button', { name: 'Select language' }));
    await user.click(screen.getByRole('button', { name: /English/i }));

    expect(setLocaleMock).toHaveBeenCalledWith('en');
  });

  it('closes dropdown on outside click', async () => {
    const user = userEvent.setup();
    render(<LanguageToggle />);

    await user.click(screen.getByRole('button', { name: 'Select language' }));
    expect(screen.getByRole('button', { name: /English/i })).toBeInTheDocument();

    fireEvent.mouseDown(document.body);

    expect(screen.queryByRole('button', { name: /English/i })).not.toBeInTheDocument();
  });
});
