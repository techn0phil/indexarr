import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { UsersPage } from './UsersPage';
import { apiClient } from '../api/client';

const getInputFromFieldLabel = (labelKey: string) => {
  const label = screen.getByText(labelKey);
  const field = label.closest('div');
  const input = field?.querySelector('input,select') as HTMLInputElement | HTMLSelectElement | null;
  if (!input) {
    throw new Error(`Could not find input for label ${labelKey}`);
  }
  return input;
};

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const baseUsers = [
  {
    id: 1,
    username: 'admin',
    role: 'admin',
    enabled: true,
    createdAt: '2026-01-01T10:00:00Z',
  },
];

describe('UsersPage CRUD flows', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    vi.spyOn(apiClient, 'getUsers').mockResolvedValue({
      success: true,
      data: baseUsers as any,
    });
  });

  it('loads and displays users list', async () => {
    render(<UsersPage />);

    expect(await screen.findByText('admin')).toBeInTheDocument();
    expect(apiClient.getUsers).toHaveBeenCalledTimes(1);
  });

  it('creates a new user from create modal', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'createUser').mockResolvedValue({ success: true, data: baseUsers[0] as any });

    render(<UsersPage />);

    await screen.findByText('admin');
    await user.click(screen.getByRole('button', { name: 'button.add' }));

    await user.type(getInputFromFieldLabel('popup.new.input.username') as HTMLInputElement, 'new-user');
    await user.type(getInputFromFieldLabel('popup.new.input.password') as HTMLInputElement, 'new-pass');
    await user.selectOptions(getInputFromFieldLabel('popup.new.input.role') as HTMLSelectElement, 'admin');
    await user.click(screen.getByRole('button', { name: 'button.create' }));

    await waitFor(() => {
      expect(apiClient.createUser).toHaveBeenCalledWith({
        username: 'new-user',
        password: 'new-pass',
        role: 'admin',
      });
      expect(apiClient.getUsers).toHaveBeenCalledTimes(2);
    });
  });

  it('updates an existing user', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'updateUser').mockResolvedValue({ success: true, data: baseUsers[0] as any });

    render(<UsersPage />);

    await screen.findByText('admin');
    await user.click(screen.getByTitle('button.edit'));

    const usernameInput = getInputFromFieldLabel('popup.edit.input.username') as HTMLInputElement;
    await user.clear(usernameInput);
    await user.type(usernameInput, 'admin2');
    await user.click(screen.getByRole('button', { name: 'button.save' }));

    await waitFor(() => {
      expect(apiClient.updateUser).toHaveBeenCalledWith(1, {
        username: 'admin2',
        role: undefined,
        enabled: undefined,
      });
    });
  });

  it('deletes a user from delete modal', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'deleteUser').mockResolvedValue({ success: true });

    render(<UsersPage />);

    await screen.findByText('admin');
    await user.click(screen.getByTitle('button.delete'));
    await user.click(screen.getAllByRole('button', { name: 'button.delete' })[1]);

    await waitFor(() => {
      expect(apiClient.deleteUser).toHaveBeenCalledWith(1);
    });
  });

  it('sets user password from password modal', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'adminSetPassword').mockResolvedValue({ success: true });

    render(<UsersPage />);

    await screen.findByText('admin');
    await user.click(screen.getByTitle('button.changePassword'));
    await user.type(getInputFromFieldLabel('popup.changePassword.input.newPassword') as HTMLInputElement, 'new-secret');
    await user.click(screen.getByRole('button', { name: 'button.save' }));

    await waitFor(() => {
      expect(apiClient.adminSetPassword).toHaveBeenCalledWith(1, 'new-secret');
    });
  });
});
