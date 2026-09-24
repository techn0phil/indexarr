import type { ReactElement } from 'react';
import { render, type RenderOptions } from '@testing-library/react';

export const renderWithProviders = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => render(ui, options);

export * from '@testing-library/react';
export { renderWithProviders as render };
