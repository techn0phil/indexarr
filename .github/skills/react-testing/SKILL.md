---
name: react-testing
description: "Create and implement React component tests for Indexarr frontend: component testing with React Testing Library, custom hook testing, API mocking, integration tests for pages, user interaction testing. Use when: writing frontend tests, testing React components, testing hooks, mocking API calls, testing user interactions."
---

# React Testing Skill — Indexarr Frontend

Generate comprehensive tests for Indexarr's React frontend using Vitest and React Testing Library.

## When to Use

- Creating test files for React components
- Testing custom hooks (useInfiniteList, useAppContext)
- Mocking API calls in tests
- Testing user interactions (clicks, filters, infinite scroll)
- Integration testing for full pages
- Testing context providers and state management

## Project Context

**Test Status**: No testing infrastructure currently exists. This skill bootstraps the testing setup.

**Tech Stack**:
- React 19 with TypeScript
- Vite (build tool)
- Custom hooks: `useInfiniteList`, `useAppContext`
- API client: `src/api/client.ts`
- CSS Modules for styling

**Key Testing Targets**:
- Components: MovieCard, FilterChip, Sidebar, ScanStatusCard
- Pages: ListFilms, ListSeries, MovieDetail, SeriesDetail
- Hooks: useInfiniteList (pagination), useAppContext (navigation state)
- API client: All fetch operations
- User interactions: Filter selection, infinite scroll, search

## Setup Testing Infrastructure

### Step 1: Install Dependencies

```bash
cd frontend/react
npm install -D vitest @vitest/ui @testing-library/react @testing-library/jest-dom @testing-library/user-event jsdom
```

**Packages**:
- `vitest` — Test runner (Vite-native, faster than Jest)
- `@vitest/ui` — Visual test UI
- `@testing-library/react` — React component testing utilities
- `@testing-library/jest-dom` — Custom matchers (toBeInTheDocument, etc.)
- `@testing-library/user-event` — User interaction simulation
- `jsdom` — DOM environment for tests

### Step 2: Configure Vitest

Create `vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: true, // Enable CSS modules in tests
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'json'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData',
      ],
    },
  },
});
```

### Step 3: Create Test Setup File

Create `src/test/setup.ts`:

```typescript
import { expect, afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import * as matchers from '@testing-library/jest-dom/matchers';

// Extend Vitest matchers
expect.extend(matchers);

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock window.matchMedia (for responsive components)
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock IntersectionObserver (for infinite scroll)
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  takeRecords() {
    return [];
  }
  unobserve() {}
} as any;
```

### Step 4: Add Test Scripts to package.json

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage"
  }
}
```

## Testing Patterns

### 1. Component Testing — Basic

Test component rendering and props:

```typescript
// src/components/MovieCard.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MovieCard } from './MovieCard';
import { Movie } from '../types';

describe('MovieCard', () => {
  const mockMovie: Movie = {
    id: 1,
    title: 'Test Movie',
    year: 2024,
    status: 'available',
    rating: 8.5,
    poster: null,
  };

  it('renders movie title', () => {
    render(<MovieCard movie={mockMovie} onClick={vi.fn()} />);
    expect(screen.getByText('Test Movie')).toBeInTheDocument();
  });

  it('displays year and rating', () => {
    render(<MovieCard movie={mockMovie} onClick={vi.fn()} />);
    expect(screen.getByText('2024')).toBeInTheDocument();
    expect(screen.getByText('8.5')).toBeInTheDocument();
  });

  it('shows initials when poster is missing', () => {
    render(<MovieCard movie={mockMovie} onClick={vi.fn()} />);
    expect(screen.getByText('TM')).toBeInTheDocument(); // Test Movie -> TM
  });

  it('displays poster image when available', () => {
    const movieWithPoster = { ...mockMovie, poster: 'https://example.com/poster.jpg' };
    render(<MovieCard movie={movieWithPoster} onClick={vi.fn()} />);
    
    const img = screen.getByRole('img', { name: 'Test Movie' });
    expect(img).toHaveAttribute('src', 'https://example.com/poster.jpg');
  });

  it('calls onClick when card is clicked', async () => {
    const handleClick = vi.fn();
    const user = userEvent.setup();
    
    render(<MovieCard movie={mockMovie} onClick={handleClick} />);
    
    await user.click(screen.getByText('Test Movie'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('applies correct status color', () => {
    const { container } = render(<MovieCard movie={mockMovie} onClick={vi.fn()} />);
    // Check status bar has green color for 'available'
    const statusBar = container.querySelector('[style*="background: #1D9E75"]');
    expect(statusBar).toBeInTheDocument();
  });
});
```

### 2. Component Testing — With Context

Test components that use context:

```typescript
// src/components/Sidebar.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { AppProvider } from '../hooks/useAppContext';
import { Sidebar } from './Sidebar';

const renderWithContext = (component: React.ReactElement) => {
  return render(<AppProvider>{component}</AppProvider>);
};

describe('Sidebar', () => {
  it('renders all navigation items', () => {
    renderWithContext(<Sidebar />);
    
    expect(screen.getByText('Films')).toBeInTheDocument();
    expect(screen.getByText('Séries')).toBeInTheDocument();
    expect(screen.getByText('Statistiques')).toBeInTheDocument();
  });

  it('highlights active navigation item', () => {
    renderWithContext(<Sidebar />);
    
    const filmsLink = screen.getByText('Films').closest('div');
    expect(filmsLink).toHaveClass('nav-item-active');
  });

  it('displays badge counts', () => {
    renderWithContext(<Sidebar />);
    
    // Assuming badges are rendered as separate elements
    expect(screen.getByText('142')).toBeInTheDocument(); // Movie count
    expect(screen.getByText('58')).toBeInTheDocument();  // Series count
  });
});
```

### 3. Custom Hook Testing

Test custom hooks using `renderHook`:

```typescript
// src/hooks/useInfiniteList.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useInfiniteList } from './useInfiniteList';
import { PaginatedResponse, Movie } from '../types';

describe('useInfiniteList', () => {
  const mockMovies: Movie[] = [
    { id: 1, title: 'Movie 1', year: 2023, status: 'available' },
    { id: 2, title: 'Movie 2', year: 2024, status: 'available' },
  ];

  const mockFetchFn = vi.fn<[number, number, Record<string, string>], Promise<PaginatedResponse<Movie>>>();

  beforeEach(() => {
    mockFetchFn.mockClear();
  });

  it('loads initial page on mount', async () => {
    mockFetchFn.mockResolvedValueOnce({
      success: true,
      data: mockMovies,
      total: 2,
      page: 1,
      pageSize: 50,
    });

    const { result } = renderHook(() =>
      useInfiniteList({ fetchFn: mockFetchFn, pageSize: 50 })
    );

    expect(result.current.isInitialLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isInitialLoading).toBe(false);
    });

    expect(mockFetchFn).toHaveBeenCalledWith(1, 50, {});
    expect(result.current.items).toHaveLength(2);
    expect(result.current.total).toBe(2);
  });

  it('loads more items when loadMore is called', async () => {
    mockFetchFn
      .mockResolvedValueOnce({
        success: true,
        data: mockMovies.slice(0, 2),
        total: 4,
        page: 1,
        pageSize: 2,
      })
      .mockResolvedValueOnce({
        success: true,
        data: [
          { id: 3, title: 'Movie 3', year: 2024, status: 'available' },
          { id: 4, title: 'Movie 4', year: 2024, status: 'available' },
        ],
        total: 4,
        page: 2,
        pageSize: 2,
      });

    const { result } = renderHook(() =>
      useInfiniteList({ fetchFn: mockFetchFn, pageSize: 2 })
    );

    await waitFor(() => {
      expect(result.current.items).toHaveLength(2);
    });

    expect(result.current.hasMore).toBe(true);

    // Load more
    result.current.loadMore();

    await waitFor(() => {
      expect(result.current.items).toHaveLength(4);
    });

    expect(mockFetchFn).toHaveBeenCalledTimes(2);
    expect(result.current.hasMore).toBe(false);
  });

  it('handles fetch errors', async () => {
    mockFetchFn.mockResolvedValueOnce({
      success: false,
      error: 'Network error',
      data: [],
      total: 0,
      page: 1,
      pageSize: 50,
    });

    const { result } = renderHook(() =>
      useInfiniteList({ fetchFn: mockFetchFn })
    );

    await waitFor(() => {
      expect(result.current.error).toBe('Network error');
    });
  });

  it('resets list when reset is called', async () => {
    mockFetchFn.mockResolvedValue({
      success: true,
      data: mockMovies,
      total: 2,
      page: 1,
      pageSize: 50,
    });

    const { result } = renderHook(() =>
      useInfiniteList({ fetchFn: mockFetchFn })
    );

    await waitFor(() => {
      expect(result.current.items).toHaveLength(2);
    });

    result.current.reset();

    expect(result.current.items).toHaveLength(0);
    expect(result.current.page).toBe(1);
  });
});
```

### 4. API Client Testing

Test API client with fetch mocking:

```typescript
// src/api/client.test.ts
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { apiClient } from './client';

describe('apiClient', () => {
  beforeEach(() => {
    global.fetch = vi.fn();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getMovies', () => {
    it('fetches movies with correct parameters', async () => {
      const mockResponse = {
        success: true,
        data: [{ id: 1, title: 'Test Movie' }],
        total: 1,
        page: 1,
        pageSize: 50,
      };

      (global.fetch as any).mockResolvedValueOnce({
        json: async () => mockResponse,
      });

      const result = await apiClient.getMovies(1, 50, { status: 'available' });

      expect(global.fetch).toHaveBeenCalledWith(
        '/api/movies?page=1&page_size=50&status=available'
      );
      expect(result).toEqual(mockResponse);
    });

    it('handles multiple filter parameters', async () => {
      const mockResponse = { success: true, data: [], total: 0, page: 1, pageSize: 50 };
      (global.fetch as any).mockResolvedValueOnce({
        json: async () => mockResponse,
      });

      await apiClient.getMovies(1, 50, {
        status: 'available',
        resolution: '3840',
        codec: 'H.265',
      });

      const url = (global.fetch as any).mock.calls[0][0];
      expect(url).toContain('status=available');
      expect(url).toContain('resolution=3840');
      expect(url).toContain('codec=H.265');
    });
  });

  describe('triggerScan', () => {
    it('sends POST request to scan endpoint', async () => {
      const mockResponse = { success: true, message: 'Scan started' };
      (global.fetch as any).mockResolvedValueOnce({
        json: async () => mockResponse,
      });

      const result = await apiClient.triggerScan();

      expect(global.fetch).toHaveBeenCalledWith('/api/scan', { method: 'POST' });
      expect(result).toEqual(mockResponse);
    });
  });
});
```

### 5. Page Integration Testing

Test full page components with mocked API:

```typescript
// src/pages/ListFilms.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ListFilms } from './ListFilms';
import { AppProvider } from '../hooks/useAppContext';
import * as apiClient from '../api/client';

// Mock the API client
vi.mock('../api/client', () => ({
  apiClient: {
    getMovies: vi.fn(),
    getStats: vi.fn(),
  },
}));

const renderWithProviders = (component: React.ReactElement) => {
  return render(<AppProvider>{component}</AppProvider>);
};

describe('ListFilms', () => {
  beforeEach(() => {
    (apiClient.apiClient.getMovies as any).mockResolvedValue({
      success: true,
      data: [
        { id: 1, title: 'Movie 1', year: 2023, status: 'available', poster: null },
        { id: 2, title: 'Movie 2', year: 2024, status: 'available', poster: null },
      ],
      total: 2,
      page: 1,
      pageSize: 50,
    });

    (apiClient.apiClient.getStats as any).mockResolvedValue({
      success: true,
      totalMovies: 142,
      totalSeries: 58,
      totalDiskSpace: 5000000000000,
      fourKPercentage: 45,
      problems: 3,
    });
  });

  it('renders movie list', async () => {
    renderWithProviders(<ListFilms />);

    await waitFor(() => {
      expect(screen.getByText('Movie 1')).toBeInTheDocument();
      expect(screen.getByText('Movie 2')).toBeInTheDocument();
    });
  });

  it('displays statistics cards', async () => {
    renderWithProviders(<ListFilms />);

    await waitFor(() => {
      expect(screen.getByText('142')).toBeInTheDocument(); // Total movies
      expect(screen.getByText('45%')).toBeInTheDocument(); // 4K percentage
    });
  });

  it('filters movies when filter is applied', async () => {
    const user = userEvent.setup();
    renderWithProviders(<ListFilms />);

    await waitFor(() => {
      expect(screen.getByText('Movie 1')).toBeInTheDocument();
    });

    // Click resolution filter chip
    const resolutionChip = screen.getByText('Résolution');
    await user.click(resolutionChip);

    // Select 4K option in modal
    const option4K = screen.getByText('3840 x 2160 (4K)');
    await user.click(option4K);

    // Apply filter
    const applyButton = screen.getByText('Appliquer');
    await user.click(applyButton);

    // Verify API called with filter
    await waitFor(() => {
      expect(apiClient.apiClient.getMovies).toHaveBeenCalledWith(
        1,
        50,
        expect.objectContaining({ resolution: '3840' })
      );
    });
  });

  it('toggles between grid and list view', async () => {
    const user = userEvent.setup();
    renderWithProviders(<ListFilms />);

    await waitFor(() => {
      expect(screen.getByText('Movie 1')).toBeInTheDocument();
    });

    // Find view toggle button
    const listViewButton = screen.getByLabelText('List view');
    await user.click(listViewButton);

    // Check localStorage was updated
    expect(localStorage.getItem('films-view')).toBe('list');
  });
});
```

### 6. User Interaction Testing

Test complex user interactions:

```typescript
// src/components/FilterModal.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { FilterModal } from './FilterModal';

describe('FilterModal', () => {
  const mockOnApply = vi.fn();
  const mockOnClose = vi.fn();

  const options = [
    { value: '3840', label: '3840 x 2160 (4K)' },
    { value: '1920', label: '1920 x 1080 (1080p)' },
    { value: '1280', label: '1280 x 720 (720p)' },
  ];

  it('renders all options', () => {
    render(
      <FilterModal
        isOpen={true}
        title="Résolution"
        options={options}
        selectedValues={[]}
        onApply={mockOnApply}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText('3840 x 2160 (4K)')).toBeInTheDocument();
    expect(screen.getByText('1920 x 1080 (1080p)')).toBeInTheDocument();
    expect(screen.getByText('1280 x 720 (720p)')).toBeInTheDocument();
  });

  it('allows selecting multiple options', async () => {
    const user = userEvent.setup();
    
    render(
      <FilterModal
        isOpen={true}
        title="Résolution"
        options={options}
        selectedValues={[]}
        onApply={mockOnApply}
        onClose={mockOnClose}
      />
    );

    // Select two options
    await user.click(screen.getByText('3840 x 2160 (4K)'));
    await user.click(screen.getByText('1920 x 1080 (1080p)'));

    // Apply
    await user.click(screen.getByText('Appliquer'));

    expect(mockOnApply).toHaveBeenCalledWith(['3840', '1920']);
  });

  it('clears all selections when Clear is clicked', async () => {
    const user = userEvent.setup();
    
    render(
      <FilterModal
        isOpen={true}
        title="Résolution"
        options={options}
        selectedValues={['3840', '1920']}
        onApply={mockOnApply}
        onClose={mockOnClose}
      />
    );

    await user.click(screen.getByText('Effacer'));
    await user.click(screen.getByText('Appliquer'));

    expect(mockOnApply).toHaveBeenCalledWith([]);
  });

  it('closes modal when clicking outside', async () => {
    const user = userEvent.setup();
    
    const { container } = render(
      <FilterModal
        isOpen={true}
        title="Résolution"
        options={options}
        selectedValues={[]}
        onApply={mockOnApply}
        onClose={mockOnClose}
      />
    );

    // Click modal backdrop
    const backdrop = container.querySelector('.modal-backdrop');
    if (backdrop) {
      await user.click(backdrop);
      expect(mockOnClose).toHaveBeenCalled();
    }
  });
});
```

## Test File Organization

```
frontend/react/
├── src/
│   ├── components/
│   │   ├── MovieCard.tsx
│   │   ├── MovieCard.test.tsx
│   │   ├── FilterChip.tsx
│   │   ├── FilterChip.test.tsx
│   │   └── ...
│   ├── pages/
│   │   ├── ListFilms.tsx
│   │   ├── ListFilms.test.tsx
│   │   └── ...
│   ├── hooks/
│   │   ├── useInfiniteList.ts
│   │   ├── useInfiniteList.test.ts
│   │   └── ...
│   ├── api/
│   │   ├── client.ts
│   │   └── client.test.ts
│   └── test/
│       ├── setup.ts              # Test setup (matchers, mocks)
│       ├── mockData.ts           # Shared mock data
│       └── testUtils.tsx         # Custom render utilities
├── vitest.config.ts
└── package.json
```

## Running Tests

```bash
# Run all tests (watch mode)
npm test

# Run tests once (CI mode)
npm run test:run

# Run with UI
npm run test:ui

# Run specific test file
npm test MovieCard.test

# Generate coverage report
npm run test:coverage

# Run tests for changed files only
npm test --changed
```

## Testing Best Practices

1. **Test User Behavior, Not Implementation**: Focus on what users see and do
2. **Use Semantic Queries**: Prefer `getByRole`, `getByLabelText` over `getByTestId`
3. **Mock External Dependencies**: Mock API calls, WebSocket connections
4. **Test Accessibility**: Verify ARIA labels, keyboard navigation
5. **Keep Tests Fast**: Avoid unnecessary waits, use `waitFor` judiciously
6. **Use Shared Utilities**: Create `renderWithProviders` helper
7. **Test Error States**: Loading states, error messages, empty states
8. **Isolate Tests**: Each test should be independent

## Common Testing Gotchas

- **CSS Modules**: Ensure `css: true` in vitest.config.ts to resolve imports
- **IntersectionObserver**: Mock in setup.ts for infinite scroll components
- **LocalStorage**: Mock or clear between tests
- **Fetch API**: Mock globally or per-test basis
- **Context Providers**: Wrap components in providers for context-dependent tests
- **Async Operations**: Always use `waitFor` for async updates
- **Timers**: Use `vi.useFakeTimers()` for setTimeout/setInterval tests

## Helper Utilities

Create `src/test/testUtils.tsx`:

```typescript
import { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { AppProvider } from '../hooks/useAppContext';

// Custom render with providers
export const renderWithProviders = (
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) => {
  return render(ui, { wrapper: AppProvider, ...options });
};

// Re-export everything from testing-library
export * from '@testing-library/react';
export { renderWithProviders as render };
```

Create `src/test/mockData.ts`:

```typescript
import { Movie, Series } from '../types';

export const mockMovie: Movie = {
  id: 1,
  title: 'Test Movie',
  year: 2024,
  duration: 120,
  synopsis: 'A test movie',
  genres: 'Action, Drama',
  rating: 8.5,
  popularity: 100,
  status: 'available',
  fileSize: 5000000000,
  filePath: '/movies/test.mkv',
  container: 'mkv',
  dateAdded: '2024-01-01',
  tmdbId: 12345,
  imdbId: 'tt1234567',
  poster: null,
};

export const mockSeries: Series = {
  id: 1,
  title: 'Test Series',
  yearStart: 2020,
  yearEnd: 2024,
  seasonCount: 4,
  episodeCount: 40,
  synopsis: 'A test series',
  genres: 'Drama',
  rating: 9.0,
  status: 'complete',
  tvdbId: 54321,
  imdbId: 'tt7654321',
  poster: null,
};
```

## Next Steps After Creating Tests

1. **Add to CI/CD**: Run `npm run test:run` in GitHub Actions
2. **Coverage Goals**: Aim for >70% coverage on components and hooks
3. **E2E Tests**: Consider Playwright for end-to-end testing
4. **Visual Regression**: Add Storybook with Chromatic for visual testing
5. **Performance Tests**: Benchmark rendering performance for large lists
6. **Accessibility Tests**: Add `vitest-axe` for automated a11y testing
