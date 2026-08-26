import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { UserMenu } from './UserMenu';
import { apiClient } from '../api/client';

const logoutMock = vi.hoisted(() => vi.fn());
const appContextState = vi.hoisted(() => ({
  user: { id: 1, username: 'admin', role: 'admin' as 'admin' | 'guest' },
  authMode: 'simple' as 'none' | 'simple' | 'oidc',
  logout: logoutMock,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => appContextState,
}));

const getInputFromFieldLabel = (labelKey: string) => {
  const label = screen.getByText(labelKey);
  const field = label.closest('div');
  const input = field?.querySelector('input') as HTMLInputElement | null;
  if (!input) {
    throw new Error(`Could not find input for label ${labelKey}`);
  }
  return input;
};

describe('UserMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    appContextState.authMode = 'simple';
    appContextState.user = { id: 1, username: 'admin', role: 'admin' };
    logoutMock.mockResolvedValue(undefined);
  });

  it('does not render when auth mode is none', () => {
    appContextState.authMode = 'none';

    const { container } = render(<UserMenu />);

    expect(container.firstChild).toBeNull();
  });

  it('opens menu and logs out user', async () => {
    const user = userEvent.setup();
    render(<UserMenu />);

    await user.click(screen.getByRole('button', { name: /admin/i }));
    await user.click(screen.getByRole('button', { name: 'usermenu.logout' }));

    await waitFor(() => {
      expect(logoutMock).toHaveBeenCalledTimes(1);
    });
  });

  it('shows role in dropdown', async () => {
    const user = userEvent.setup();
    render(<UserMenu />);

    await user.click(screen.getByRole('button', { name: /admin/i }));

    expect(screen.getByText('role.admin')).toBeInTheDocument();
  });

  it('validates password change fields before submit', async () => {
    const user = userEvent.setup();
    render(<UserMenu />);

    await user.click(screen.getByRole('button', { name: /admin/i }));
    await user.click(screen.getByRole('button', { name: 'usermenu.updatePassword' }));
    await user.click(screen.getByRole('button', { name: 'popup.updatePassword.button.update' }));

    expect(screen.getByText('popup.updatePassword.message.requiredFields')).toBeInTheDocument();
  });

  it('submits password change and shows success state', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'changePassword').mockResolvedValue({ success: true } as any);

    render(<UserMenu />);

    await user.click(screen.getByRole('button', { name: /admin/i }));
    await user.click(screen.getByRole('button', { name: 'usermenu.updatePassword' }));

    await user.type(getInputFromFieldLabel('popup.updatePassword.input.currentPassword'), 'old-pass');
    await user.type(getInputFromFieldLabel('popup.updatePassword.input.newPassword'), 'new-pass');
    await user.type(getInputFromFieldLabel('popup.updatePassword.input.confirmPassword'), 'new-pass');
    await user.click(screen.getByRole('button', { name: 'popup.updatePassword.button.update' }));

    await waitFor(() => {
      expect(apiClient.changePassword).toHaveBeenCalledWith('old-pass', 'new-pass');
      expect(screen.getByText('popup.updatePassword.message.passwordUpdated')).toBeInTheDocument();
    });
  });
});
