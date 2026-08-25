import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { LoginPage } from './LoginPage';

const loginMock = vi.fn();

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => ({
    login: loginMock,
  }),
}));

vi.mock('../components/ThemeToggle', () => ({
  ThemeToggle: () => <button type="button">theme</button>,
}));

describe('LoginPage', () => {
  beforeEach(() => {
    loginMock.mockReset();
  });

  it('renders login form fields and submit button', () => {
    render(<LoginPage />);

    expect(screen.getByLabelText("Nom d'utilisateur")).toBeInTheDocument();
    expect(screen.getByLabelText('Mot de passe')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Se connecter' })).toBeInTheDocument();
  });

  it('submits credentials via context login function', async () => {
    const user = userEvent.setup();
    loginMock.mockResolvedValueOnce({ success: true });

    render(<LoginPage />);

    await user.type(screen.getByLabelText("Nom d'utilisateur"), 'admin');
    await user.type(screen.getByLabelText('Mot de passe'), 'secret');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledWith('admin', 'secret');
    });
  });

  it('shows backend error on failed login', async () => {
    const user = userEvent.setup();
    loginMock.mockResolvedValueOnce({ success: false, error: 'Invalid credentials' });

    render(<LoginPage />);

    await user.type(screen.getByLabelText("Nom d'utilisateur"), 'admin');
    await user.type(screen.getByLabelText('Mot de passe'), 'wrong');
    await user.click(screen.getByRole('button', { name: 'Se connecter' }));

    expect(await screen.findByText('Invalid credentials')).toBeInTheDocument();
  });

  it('disables submit button when fields are empty', () => {
    render(<LoginPage />);

    const submit = screen.getByRole('button', { name: 'Se connecter' });
    expect(submit).toBeDisabled();
  });
});
