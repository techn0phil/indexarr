import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { fireEvent, render, screen } from '../test/testUtils';
import { Topbar } from './Topbar';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('./ThemeToggle', () => ({ ThemeToggle: () => <div>theme-toggle</div> }));
vi.mock('./LanguageToggle', () => ({ LanguageToggle: () => <div>language-toggle</div> }));
vi.mock('./UserMenu', () => ({ UserMenu: () => <div>user-menu</div> }));

describe('Topbar', () => {
  it('renders back button and calls onBack', async () => {
    const user = userEvent.setup();
    const onBack = vi.fn();

    render(
      <Topbar
        showBack={true}
        breadcrumb="Movies / Item"
        onBack={onBack}
        searchQuery=""
        onSearchChange={vi.fn()}
      />
    );

    await user.click(screen.getByRole('button', { name: 'button.back' }));
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  it('calls onSearchChange on input and clear', async () => {
    const user = userEvent.setup();
    const onSearchChange = vi.fn();

    const { rerender } = render(
      <Topbar
        showBack={false}
        breadcrumb=""
        onBack={vi.fn()}
        searchQuery=""
        onSearchChange={onSearchChange}
      />
    );

    await user.type(screen.getByPlaceholderText('input.filter.placeholder'), 'abc');
    expect(onSearchChange).toHaveBeenCalled();

    rerender(
      <Topbar
        showBack={false}
        breadcrumb=""
        onBack={vi.fn()}
        searchQuery="abc"
        onSearchChange={onSearchChange}
      />
    );

    await user.click(screen.getByRole('button', { name: 'button.clear' }));
    expect(onSearchChange).toHaveBeenLastCalledWith('');
  });

  it('focuses search input when slash key is pressed', () => {
    render(
      <Topbar
        showBack={false}
        breadcrumb=""
        onBack={vi.fn()}
        searchQuery=""
        onSearchChange={vi.fn()}
      />
    );

    const input = screen.getByPlaceholderText('input.filter.placeholder');
    fireEvent.keyDown(window, { key: '/' });

    expect(document.activeElement).toBe(input);
  });
});
