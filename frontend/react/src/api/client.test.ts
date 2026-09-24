import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from './client';

describe('apiClient', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('includes credentials in getMovies request with filters', async () => {
    const fetchMock = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      json: async () => ({ success: true, data: [], total: 0, page: 1, pageSize: 50 }),
    });

    await apiClient.getMovies(1, 50, { status: 'available', resolution: '3840' });

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toContain('/api/movies?');
    expect(url).toContain('page=1');
    expect(url).toContain('page_size=50');
    expect(url).toContain('status=available');
    expect(url).toContain('resolution=3840');
    expect(options).toMatchObject({ credentials: 'include' });
  });

  it('sends login payload as JSON with credentials', async () => {
    const fetchMock = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      json: async () => ({ success: true, user: { id: 1, username: 'admin', role: 'admin' } }),
    });

    await apiClient.login('admin', 'secret');

    const [url, options] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/auth/login');
    expect(options).toMatchObject({
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
    });
    expect(options.body).toBe(JSON.stringify({ username: 'admin', password: 'secret' }));
  });

  it('throws from getCurrentUser when response is not ok', async () => {
    const fetchMock = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValueOnce({
      ok: false,
      json: async () => ({}),
    });

    await expect(apiClient.getCurrentUser()).rejects.toThrow('Not authenticated');
  });

  it('calls scan endpoints with POST', async () => {
    const fetchMock = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValue({
      json: async () => ({ success: true }),
    });

    await apiClient.triggerScan();
    await apiClient.stopScan();

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/scan', {
      method: 'POST',
      credentials: 'include',
    });
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/scan/stop', {
      method: 'POST',
      credentials: 'include',
    });
  });

  it('calls admin user endpoints with proper method and body', async () => {
    const fetchMock = globalThis.fetch as unknown as ReturnType<typeof vi.fn>;
    fetchMock.mockResolvedValue({
      json: async () => ({ success: true }),
    });

    await apiClient.createUser({ username: 'john', password: 'pw', role: 'guest' });
    await apiClient.updateUser(7, { role: 'admin', enabled: true });
    await apiClient.deleteUser(7);

    expect(fetchMock).toHaveBeenNthCalledWith(1, '/api/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'john', password: 'pw', role: 'guest' }),
      credentials: 'include',
    });
    expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/users/7', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ role: 'admin', enabled: true }),
      credentials: 'include',
    });
    expect(fetchMock).toHaveBeenNthCalledWith(3, '/api/users/7', {
      method: 'DELETE',
      credentials: 'include',
    });
  });
});
