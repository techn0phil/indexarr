import { beforeEach, describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '../test/testUtils';
import { ScanStatusCard } from './ScanStatusCard';
import { apiClient } from '../api/client';

const appContextState = vi.hoisted(() => ({
  scanStatus: { id: 1, status: 'idle', filesFound: 0, filesProcessed: 0 } as any,
  authMode: 'none' as 'none' | 'simple' | 'oidc',
  user: null as any,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, options?: { time?: string }) => {
      if (key === 'statCard.scan.label.ago') {
        return `ago ${options?.time || ''}`;
      }
      if (key === 'statCard.scan.label.now') {
        return 'now';
      }
      return key;
    },
  }),
}));

vi.mock('../hooks/useAppContext', () => ({
  useAppContext: () => appContextState,
}));

describe('ScanStatusCard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    appContextState.scanStatus = { id: 1, status: 'idle', filesFound: 0, filesProcessed: 0 };
    appContextState.authMode = 'none';
    appContextState.user = null;
  });

  it('renders start button and triggers scan', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'triggerScan').mockResolvedValue({ success: true } as any);

    render(<ScanStatusCard />);

    await user.click(screen.getByRole('button', { name: 'statCard.scan.button.start' }));

    expect(apiClient.triggerScan).toHaveBeenCalledTimes(1);
  });

  it('renders progress and stop action when running', async () => {
    const user = userEvent.setup();
    vi.spyOn(apiClient, 'stopScan').mockResolvedValue({ success: true } as any);

    appContextState.scanStatus = {
      id: 1,
      status: 'running',
      filesFound: 10,
      filesProcessed: 4,
      startedAt: new Date().toISOString(),
    };

    render(<ScanStatusCard />);

    expect(screen.getByText('4 / 10 statCard.scan.label.files (40%)')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'statCard.scan.button.stop' }));
    expect(apiClient.stopScan).toHaveBeenCalledTimes(1);
  });

  it('calls onScanComplete when status transitions from running to completed', async () => {
    const onScanComplete = vi.fn();

    appContextState.scanStatus = {
      id: 1,
      status: 'running',
      filesFound: 10,
      filesProcessed: 4,
      startedAt: new Date().toISOString(),
    };

    const { rerender } = render(<ScanStatusCard onScanComplete={onScanComplete} />);

    appContextState.scanStatus = {
      id: 1,
      status: 'completed',
      filesFound: 10,
      filesProcessed: 10,
      completedAt: new Date().toISOString(),
    };

    rerender(<ScanStatusCard onScanComplete={onScanComplete} />);

    await waitFor(() => {
      expect(onScanComplete).toHaveBeenCalledTimes(1);
    });
  });

  it('hides scan controls for guest user in simple mode', () => {
    appContextState.authMode = 'simple';
    appContextState.user = { id: 2, username: 'guest', role: 'guest' };

    render(<ScanStatusCard />);

    expect(screen.queryByRole('button', { name: 'statCard.scan.button.start' })).not.toBeInTheDocument();
  });
});
