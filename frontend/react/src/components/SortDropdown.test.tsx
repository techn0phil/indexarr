import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen } from '../test/testUtils';
import { SortDropdown } from './SortDropdown';

const options = [
  { value: 'title', label: 'Title' },
  { value: 'year', label: 'Year' },
];

describe('SortDropdown', () => {
  it('shows the current option with an up arrow for ascending order', () => {
    render(<SortDropdown options={options} sort={{ key: 'year', order: 'asc' }} onSelect={vi.fn()} ariaLabel="Sort by" />);

    expect(screen.getByRole('button', { name: 'Sort by' })).toHaveTextContent('Year');
    expect(screen.getByTestId('sort-icon-asc')).toBeInTheDocument();
    expect(screen.queryByTestId('sort-icon-desc')).not.toBeInTheDocument();
  });

  it('shows a down arrow for descending order', () => {
    render(<SortDropdown options={options} sort={{ key: 'title', order: 'desc' }} onSelect={vi.fn()} ariaLabel="Sort by" />);

    expect(screen.getByTestId('sort-icon-desc')).toBeInTheDocument();
  });

  it('calls onSelect and closes the menu when an option is chosen', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<SortDropdown options={options} sort={{ key: 'title', order: 'asc' }} onSelect={onSelect} ariaLabel="Sort by" />);

    await user.click(screen.getByRole('button', { name: 'Sort by' }));
    await user.click(screen.getByText('Year'));

    expect(onSelect).toHaveBeenCalledWith('year');
    expect(screen.getByRole('button', { name: 'Sort by' })).toHaveAttribute('aria-expanded', 'false');
  });

  it('closes the menu on outside click', async () => {
    const user = userEvent.setup();
    render(<SortDropdown options={options} sort={{ key: 'title', order: 'asc' }} onSelect={vi.fn()} ariaLabel="Sort by" />);

    await user.click(screen.getByRole('button', { name: 'Sort by' }));
    expect(screen.getByRole('button', { name: 'Sort by' })).toHaveAttribute('aria-expanded', 'true');

    await user.click(document.body);
    expect(screen.getByRole('button', { name: 'Sort by' })).toHaveAttribute('aria-expanded', 'false');
  });
});
