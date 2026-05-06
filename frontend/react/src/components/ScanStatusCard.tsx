import { useCallback, useEffect, useRef, useState } from 'react';
import { ScanStatus as ScanStatusType } from '../types';
import { apiClient } from '../api/client';
import comStyles from '../styles/components.module.css';

interface ScanStatusProps {
  onScanComplete?: () => void;
}

export const ScanStatusCard = ({ onScanComplete }: ScanStatusProps) => {
  const [status, setStatus] = useState<ScanStatusType | null>(null);
  const [loading, setLoading] = useState(false);
  const [triggering, setTriggering] = useState(false);
  const onScanCompleteRef = useRef(onScanComplete);

  // Keep ref in sync with latest callback
  useEffect(() => {
    onScanCompleteRef.current = onScanComplete;
  }, [onScanComplete]);

  // Memoized fetch status function
  const fetchStatus = useCallback(async () => {
    try {
      const data = await apiClient.getScanStatus();
      setStatus(data);
      
      // Notify parent when scan completes
      if (data.status === 'completed' && onScanCompleteRef.current) {
        onScanCompleteRef.current();
      }
    } catch (error) {
      console.error('Failed to fetch scan status:', error);
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    
    // Poll while scanning
    const interval = setInterval(() => {
      fetchStatus();
    }, status?.status === 'running' ? 2000 : 3600000);

    return () => clearInterval(interval);
  }, [status?.status, fetchStatus]);

  const handleTriggerScan = async () => {
    setTriggering(true);
    try {
      await apiClient.triggerScan();
      fetchStatus();
    } catch (error) {
      console.error('Failed to trigger scan:', error);
    } finally {
      setTriggering(false);
    }
  };

  const handleStopScan = async () => {
    try {
      await apiClient.stopScan();
      fetchStatus();
    } catch (error) {
      console.error('Failed to stop scan:', error);
    }
  };

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Jamais';
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);

    if (diffMins < 1) return "À l'instant";
    if (diffMins < 60) return `Il y a ${diffMins} min`;
    if (diffHours < 24) return `Il y a ${diffHours}h`;
    return `Il y a ${diffDays}j`;
  };

  const getStatusLabel = () => {
    if (!status) return 'Chargement...';
    switch (status.status) {
      case 'running':
        return 'Scan en cours...';
      case 'completed':
        return 'Dernier scan terminé';
      case 'error':
        return 'Erreur lors du scan';
      case 'stopped':
        return 'Scan arrêté';
      default:
        return 'En attente';
    }
  };

  const getProgress = () => {
    if (!status || status.filesFound === 0) return 0;
    return Math.round((status.filesProcessed / status.filesFound) * 100);
  };

  return (
    <div className={comStyles.stat} style={{ position: 'relative' }}>
      <div className={comStyles['stat-label']}>Scan de la bibliothèque</div>
      
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginTop: '4px' }}>
        <div style={{ 
          width: '8px', 
          height: '8px', 
          borderRadius: '50%', 
          backgroundColor: status?.status === 'running' ? '#1D9E75' : 
                          status?.status === 'error' ? '#E24B4A' : 
                          'var(--color-text-tertiary)',
          animation: status?.status === 'running' ? 'pulse 1.5s infinite' : 'none'
        }} />
        <span style={{ fontSize: '13px', fontWeight: 500 }}>{getStatusLabel()}</span>
      </div>

      {status?.status === 'running' && (
        <div style={{ marginTop: '8px' }}>
          <div style={{ 
            height: '4px', 
            backgroundColor: 'var(--color-background-tertiary)', 
            borderRadius: '2px',
            overflow: 'hidden'
          }}>
            <div style={{ 
              height: '100%', 
              width: `${getProgress()}%`,
              backgroundColor: '#1D9E75',
              transition: 'width 0.3s ease'
            }} />
          </div>
          <div style={{ 
            fontSize: '10px', 
            color: 'var(--color-text-tertiary)', 
            marginTop: '4px' 
          }}>
            {status.filesProcessed} / {status.filesFound} fichiers ({getProgress()}%)
          </div>
        </div>
      )}

      {status?.status !== 'running' && status?.completedAt && (
        <div className={comStyles['stat-sub']} style={{ marginTop: '4px' }}>
          {formatDate(status.completedAt)}
        </div>
      )}

      {status?.errorMessage && (
        <div style={{ 
          fontSize: '10px', 
          color: '#E24B4A', 
          marginTop: '4px' 
        }}>
          {status.errorMessage}
        </div>
      )}

      <div style={{ marginTop: '10px', display: 'flex', gap: '6px' }}>
        {status?.status === 'running' ? (
          <button
            onClick={handleStopScan}
            style={{
              fontSize: '11px',
              padding: '4px 10px',
              borderRadius: '6px',
              border: '0.5px solid #E24B4A',
              backgroundColor: 'transparent',
              color: '#E24B4A',
              cursor: 'pointer',
            }}
          >
            Arrêter
          </button>
        ) : (
          <button
            onClick={handleTriggerScan}
            disabled={triggering}
            style={{
              fontSize: '11px',
              padding: '4px 10px',
              borderRadius: '6px',
              border: 'none',
              backgroundColor: '#1D9E75',
              color: 'white',
              cursor: triggering ? 'wait' : 'pointer',
              opacity: triggering ? 0.7 : 1,
            }}
          >
            {triggering ? 'Démarrage...' : 'Lancer un scan'}
          </button>
        )}
      </div>

      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.5; }
        }
      `}</style>
    </div>
  );
};
