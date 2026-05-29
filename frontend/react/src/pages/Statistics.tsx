import { useEffect, useMemo, useState } from 'react';
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from 'recharts';
import { useAppContext } from '../hooks/useAppContext';
import { StatsDistributionItem } from '../types';
import styles from '../styles/stats.module.css';

type DonutDatum = {
  name: string;
  count: number;
  color: string;
};

const VIDEO_CODEC_COLORS = ['#1D9E75', '#3B82F6', '#8B5CF6', '#14B8A6', '#F59E0B', '#EF4444'];
const RESOLUTION_COLORS = ['#1D9E75', '#3B82F6', '#8B5CF6', '#F59E0B'];
const HDR_COLORS = ['#8B5CF6', '#3B82F6', '#1D9E75', '#F59E0B'];
const AUDIO_COLORS = ['#1D9E75', '#14B8A6', '#3B82F6', '#F59E0B', '#8B5CF6', '#EF4444'];

const buildDonutData = (items: StatsDistributionItem[], colors: string[]): DonutDatum[] => {
  return items
    .filter((item) => item.count > 0)
    .map((item, index) => ({
      name: item.name,
      count: item.count,
      color: colors[index % colors.length],
    }));
};

const formatPct = (value: number, total: number): string => {
  if (total <= 0) return '0%';
  return `${Math.round((value / total) * 100)}%`;
};

const formatStorage = (diskSpaceGB: number): string => {
  const toValue = diskSpaceGB / 1024;
  if (toValue >= 1) {
    return `${toValue.toFixed(1).replace('.', ',')} To`;
  }
  return `${diskSpaceGB.toFixed(1).replace('.', ',')} Go`;
};

const renderKpiIcon = (type: 'movies' | 'series' | 'storage' | 'problems') => {
  if (type === 'movies') {
    return (
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="2" y="3" width="12" height="10" rx="1.5" />
        <path d="M5 3v10M11 3v10M2 7h12" />
      </svg>
    );
  }
  if (type === 'series') {
    return (
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="2" y="2" width="12" height="12" rx="1.5" />
        <path d="M2 6h12M6 6v8" />
      </svg>
    );
  }
  if (type === 'storage') {
    return (
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
        <rect x="2" y="3" width="12" height="10" rx="1.5" />
        <path d="M5 7l2 2 4-4" />
      </svg>
    );
  }

  return (
    <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
      <circle cx="8" cy="8" r="6" />
      <path d="M8 5v4M8 11h.01" />
    </svg>
  );
};

interface DonutCardProps {
  title: string;
  data: DonutDatum[];
}

const DonutCard = ({ title, data }: DonutCardProps) => {
  const [hiddenLabels, setHiddenLabels] = useState<Set<string>>(new Set());

  const visibleData = useMemo(() => data.filter((item) => !hiddenLabels.has(item.name)), [data, hiddenLabels]);
  const total = useMemo(() => visibleData.reduce((sum, item) => sum + item.count, 0), [visibleData]);
  const dominantItem = visibleData[0];

  const toggleLabel = (label: string) => {
    setHiddenLabels((prev) => {
      const next = new Set(prev);
      if (next.has(label)) {
        next.delete(label);
      } else {
        next.add(label);
      }
      return next;
    });
  };

  return (
    <div className={styles.card}>
      <div className={styles.cardTitle}>{title}</div>
      {data.length === 0 ? (
        <div className={styles.state}>Aucune donnée disponible.</div>
      ) : (
        <div className={styles.donutWrap}>
          <div style={{ width: 150, height: 150 }}>
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  animationDuration={1000}
                  data={visibleData}
                  dataKey="count"
                  nameKey="name"
                  innerRadius={38}
                  outerRadius={60}
                  stroke="none"
                  paddingAngle={2}
                >
                  {visibleData.map((entry) => (
                    <Cell key={entry.name} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value, name) => [
                    `${Number(value ?? 0)} (${formatPct(Number(value ?? 0), total)})`,
                    String(name),
                  ]}
                />
                <text x="50%" y="47%" textAnchor="middle" className={styles.donutCenter}>
                  {dominantItem ? formatPct(dominantItem.count, total) : '0%'}
                </text>
                <text x="50%" y="57%" textAnchor="middle" className={styles.donutCenterSub}>
                  {dominantItem?.name ?? '—'}
                </text>
              </PieChart>
            </ResponsiveContainer>
          </div>

          <div className={styles.legend}>
            {data.map((item) => {
              const isHidden = hiddenLabels.has(item.name);
              return (
                <button
                  key={item.name}
                  type="button"
                  onClick={() => toggleLabel(item.name)}
                  className={`${styles.legendItem} ${isHidden ? styles.legendItemMuted : ''}`}
                  title={isHidden ? 'Afficher cette catégorie' : 'Masquer cette catégorie'}
                >
                  <span className={styles.legendDot} style={{ background: item.color }} />
                  <span className={styles.legendLabel}>{item.name}</span>
                  <span className={styles.legendPct}>{formatPct(item.count, total || 1)}</span>
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};

interface ProgressCardProps {
  title: string;
  items: DonutDatum[];
}

const ProgressCard = ({ title, items }: ProgressCardProps) => {
  const total = items.reduce((sum, item) => sum + item.count, 0);
  const [animateBars, setAnimateBars] = useState(false);

  useEffect(() => {
    setAnimateBars(false);
    const rafId = requestAnimationFrame(() => {
      setAnimateBars(true);
    });

    return () => cancelAnimationFrame(rafId);
  }, [items]);

  return (
    <div className={styles.card}>
      <div className={styles.cardTitle}>{title}</div>
      {items.length === 0 ? (
        <div className={styles.state}>Aucune donnée disponible.</div>
      ) : (
        <div className={styles.progressList}>
          {items.map((item, index) => {
            const pct = total > 0 ? (item.count / total) * 100 : 0;
            return (
              <div className={styles.progressItem} key={item.name}>
                <div className={styles.progressLabelRow}>
                  <span className={styles.progressLabel}>{item.name}</span>
                  <span className={styles.progressValue}>{`${Math.round(pct)}%`}</span>
                </div>
                <div className={styles.progressTrack} title={`${item.name}: ${item.count}`}>
                  <div
                    className={styles.progressFill}
                    style={{
                      width: animateBars ? `${pct}%` : '0%',
                      background: item.color,
                      transitionDelay: `${index * 60}ms`,
                    }}
                  />
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export const Statistics = () => {
  const { stats, statsLoading } = useAppContext();

  const videoData = useMemo(
    () => buildDonutData(stats?.videoCodecDistribution ?? [], VIDEO_CODEC_COLORS),
    [stats?.videoCodecDistribution]
  );
  const resolutionData = useMemo(
    () => buildDonutData(stats?.resolutionDistribution ?? [], RESOLUTION_COLORS),
    [stats?.resolutionDistribution]
  );
  const hdrData = useMemo(() => buildDonutData(stats?.hdrDistribution ?? [], HDR_COLORS), [stats?.hdrDistribution]);
  const audioFormatData = useMemo(
    () => buildDonutData(stats?.audioFormatDistribution ?? [], AUDIO_COLORS),
    [stats?.audioFormatDistribution]
  );
  const audioLanguageData = useMemo(
    () => buildDonutData(stats?.audioLanguageDistribution ?? [], AUDIO_COLORS),
    [stats?.audioLanguageDistribution]
  );

  if (statsLoading) {
    return <div className={styles.page}><div className={styles.state}>Chargement des statistiques...</div></div>;
  }

  if (!stats) {
    return <div className={styles.page}><div className={styles.state}>Impossible de charger les statistiques.</div></div>;
  }

  return (
    <div className={styles.page}>
      <div className={styles.kpiGrid}>
        <div className={styles.kpiCard}>
          <div className={`${styles.kpiIcon} ${styles.iconGreen}`}>{renderKpiIcon('movies')}</div>
          <div className={styles.kpiValue}>{stats.totalMovies}</div>
          <div className={styles.kpiLabel}>Films</div>
          <div className={`${styles.kpiDelta} ${styles.kpiDeltaUp}`}>{`${stats.availMovies} disponibles`}</div>
        </div>

        <div className={styles.kpiCard}>
          <div className={`${styles.kpiIcon} ${styles.iconBlue}`}>{renderKpiIcon('series')}</div>
          <div className={styles.kpiValue}>{stats.totalSeries}</div>
          <div className={styles.kpiLabel}>{`Séries · ${stats.totalEpisodes} épisodes`}</div>
          <div className={`${styles.kpiDelta} ${styles.kpiDeltaUp}`}>{`${stats.availEpisodes} épisodes disponibles`}</div>
        </div>

        <div className={styles.kpiCard}>
          <div className={`${styles.kpiIcon} ${styles.iconPurple}`}>{renderKpiIcon('storage')}</div>
          <div className={styles.kpiValue}>{formatStorage(stats.diskSpaceGB)}</div>
          <div className={styles.kpiLabel}>To stockage total</div>
          <div className={styles.kpiDelta}>{`${stats.fourKCount} titres 4K · ${Math.round(stats.fourKPercent)}% films`}</div>
        </div>

        <div className={styles.kpiCard}>
          <div className={`${styles.kpiIcon} ${styles.iconAmber}`}>{renderKpiIcon('problems')}</div>
          <div className={`${styles.kpiValue} ${styles.kpiError}`}>{stats.problemsCount}</div>
          <div className={styles.kpiLabel}>Problèmes actifs</div>
          <div className={styles.kpiDelta}>{`${stats.missingMovies} manquants · ${stats.missingEpisodes} ep. manquants`}</div>
        </div>
      </div>

      <div className={styles.rowThree}>
        <DonutCard title="Codec vidéo — distribution" data={videoData} />
        <DonutCard title="Résolution — répartition" data={resolutionData} />
        <DonutCard title="HDR — couverture" data={hdrData} />
      </div>

      <div className={styles.rowTwo}>
        <ProgressCard title="Audio — formats" items={audioFormatData} />
        <ProgressCard title="Audio — langues" items={audioLanguageData} />
      </div>
    </div>
  );
};
