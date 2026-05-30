import { useEffect, useMemo, useState } from 'react';
import { apiClient } from '../api/client';
import { MediaIssue, Movie, IssueLevel, IssueType } from '../types';
import { useAppContext } from '../hooks/useAppContext';
import styles from '../styles/problems.module.css';

type FilterTab = 'all' | 'missing' | 'duplicate' | 'encoding' | 'metadata';

const ISSUE_TABS: { key: FilterTab; label: string; matches: IssueType[] }[] = [
  {
    key: 'all',
    label: 'Tous',
    matches: [
      'missing_file',
      'corrupted_file',
      'invalid_metadata',
      'duplicate_movies',
      'low_bitrate',
      'missing_episodes',
      'missing_subtitles',
      'missing_poster',
      'unusual_codec',
      'incorrect_encoding',
      'duration_mismatch',
      'missing_minor_metadata',
    ],
  },
  {
    key: 'missing',
    label: 'Manquants',
    matches: ['missing_file', 'missing_episodes'],
  },
  {
    key: 'duplicate',
    label: 'Doublons',
    matches: ['duplicate_movies'],
  },
  {
    key: 'encoding',
    label: 'Encodage',
    matches: ['incorrect_encoding', 'low_bitrate', 'unusual_codec'],
  },
  {
    key: 'metadata',
    label: 'Métadonnées',
    matches: ['invalid_metadata', 'missing_subtitles', 'missing_poster', 'duration_mismatch', 'missing_minor_metadata'],
  },
];

const levelClassName: Record<IssueLevel, string> = {
  error: styles.issueError,
  warning: styles.issueWarning,
  info: styles.issueInfo,
  notice: styles.issueNotice,
};

const levelLabel: Record<IssueLevel, string> = {
  error: 'Erreur',
  warning: 'Avertissement',
  info: 'Info',
  notice: 'Notice',
};

const typeLabel: Record<IssueType, string> = {
  missing_file: 'Fichier manquant',
  corrupted_file: 'Fichier corrompu',
  invalid_metadata: 'Métadonnées invalides',
  duplicate_movies: 'Doublon film',
  low_bitrate: 'Bitrate faible',
  missing_episodes: 'Épisodes manquants',
  missing_subtitles: 'Sous-titres manquants',
  missing_poster: 'Poster manquant',
  unusual_codec: 'Codec inhabituel',
  incorrect_encoding: 'Encodage incorrect',
  duration_mismatch: 'Durée incohérente',
  missing_minor_metadata: 'Métadonnées mineures manquantes',
};

const codecFromMovie = (movie: Movie): string => {
  return movie.mediaInfo?.videoTracks?.[0]?.codec || '';
};

const bitrateFromMovie = (movie: Movie): number | null => {
  const bitrate = movie.mediaInfo?.videoTracks?.[0]?.bitrate || '';
  const match = bitrate.replace(',', '.').match(/([0-9]+(?:\.[0-9]+)?)/);
  if (!match) return null;
  const parsed = Number(match[1]);
  return Number.isFinite(parsed) ? parsed : null;
};

const resolutionFromMovie = (movie: Movie): string => {
  return movie.mediaInfo?.videoTracks?.[0]?.resolution || '';
};

const buildMovieIssues = (movies: Movie[]): MediaIssue[] => {
  const issues: MediaIssue[] = [];

  const byTitleYear = new Map<string, Movie[]>();
  movies.forEach((movie) => {
    const key = `${movie.title.toLowerCase().trim()}::${movie.year}`;
    const list = byTitleYear.get(key) || [];
    list.push(movie);
    byTitleYear.set(key, list);
  });

  movies.forEach((movie) => {
    if (movie.status === 'missing') {
      issues.push({
        id: `missing-file-${movie.id}`,
        type: 'missing_file',
        level: 'error',
        title: `${movie.title} (${movie.year}) — fichier introuvable`,
        description: 'Le film est référencé dans la librairie mais le fichier n\'est pas accessible.',
        path: movie.filePath,
        tags: ['Manquant'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Rechercher', secondary: 'Ignorer' },
      });
      return;
    }

    const duplicates = byTitleYear.get(`${movie.title.toLowerCase().trim()}::${movie.year}`) || [];
    if (duplicates.length > 1) {
      issues.push({
        id: `duplicate-${movie.id}`,
        type: 'duplicate_movies',
        level: 'warning',
        title: `${movie.title} — doublon détecté`,
        description: 'Plusieurs entrées semblent correspondre au même film (heuristique titre + année).',
        path: movie.filePath,
        tags: ['Doublon'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Garder meilleure version', secondary: 'Ignorer' },
      });
    }

    if (!movie.poster) {
      issues.push({
        id: `poster-${movie.id}`,
        type: 'missing_poster',
        level: 'info',
        title: `${movie.title} — poster manquant`,
        description: 'Aucune illustration n\'est disponible pour ce média.',
        path: movie.filePath,
        tags: ['Artwork'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Rafraîchir artwork', secondary: 'Ignorer' },
      });
    }

    const codec = codecFromMovie(movie).toUpperCase();
    if (codec && !codec.includes('H.264') && !codec.includes('H.265') && !codec.includes('AV1') && !codec.includes('HEVC')) {
      issues.push({
        id: `codec-${movie.id}`,
        type: 'unusual_codec',
        level: 'info',
        title: `${movie.title} — codec inhabituel`,
        description: `Codec détecté: ${codecFromMovie(movie)}. Compatibilité potentiellement limitée selon les clients.`,
        path: movie.filePath,
        tags: ['Compatibilité'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Recommander transcodage', secondary: 'Ignorer' },
      });
    }

    if (codec.includes('H.264')) {
      issues.push({
        id: `encoding-${movie.id}`,
        type: 'incorrect_encoding',
        level: 'notice',
        title: `${movie.title} — encodage non optimal`,
        description: 'Le profil d\'encodage diffère du standard préféré de la librairie.',
        path: movie.filePath,
        tags: ['Optimisation'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Planifier ré-encodage', secondary: 'Ignorer' },
      });
    }

    const bitrate = bitrateFromMovie(movie);
    const resolution = resolutionFromMovie(movie);
    if (bitrate !== null && resolution.includes('3840') && bitrate < 3) {
      issues.push({
        id: `bitrate-${movie.id}`,
        type: 'low_bitrate',
        level: 'warning',
        title: `${movie.title} — bitrate faible`,
        description: 'Le bitrate semble bas pour une source 4K, risque de qualité visuelle réduite.',
        path: movie.filePath,
        tags: ['Qualité'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Rechercher meilleure source', secondary: 'Ignorer' },
      });
    }

    const subtitleCount = movie.mediaInfo?.subtitleTracks?.length || 0;
    if (subtitleCount === 0) {
      issues.push({
        id: `subs-${movie.id}`,
        type: 'missing_subtitles',
        level: 'info',
        title: `${movie.title} — sous-titres manquants`,
        description: 'Aucune piste de sous-titres détectée pour ce fichier.',
        path: movie.filePath,
        tags: ['Sous-titres'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Récupérer sous-titres', secondary: 'Ignorer' },
      });
    }

    if (!movie.synopsis || !movie.genres || !movie.imdbId) {
      issues.push({
        id: `minor-meta-${movie.id}`,
        type: 'missing_minor_metadata',
        level: 'notice',
        title: `${movie.title} — métadonnées incomplètes`,
        description: 'Des champs non bloquants sont absents (synopsis, genres ou identifiant externe).',
        path: movie.filePath,
        tags: ['Complétude'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Rafraîchir métadonnées', secondary: 'Ignorer' },
      });
    }

    if ((movie.duration || 0) <= 0) {
      issues.push({
        id: `invalid-meta-${movie.id}`,
        type: 'invalid_metadata',
        level: 'error',
        title: `${movie.title} — métadonnées invalides`,
        description: 'Durée ou champs techniques invalides. Une nouvelle analyse est requise.',
        path: movie.filePath,
        tags: ['Metadata error'],
        mediaType: 'movie',
        mediaId: movie.id,
        actions: { primary: 'Re-scan metadata', secondary: 'Ignorer' },
      });
    }
  });

  return issues;
};

const buildSeriesIssues = (series: Array<{ id: number; title: string; status: string; episodeCount: number; seasonCount: number }>, missingEpisodes: number): MediaIssue[] => {
  const issues: MediaIssue[] = [];

  if (missingEpisodes > 0) {
    const partialSeries = series.filter((item) => item.status === 'partial');
    if (partialSeries.length > 0) {
      const totalPerSeries = Math.max(1, Math.ceil(missingEpisodes / partialSeries.length));
      partialSeries.forEach((item) => {
        issues.push({
          id: `missing-episodes-${item.id}`,
          type: 'missing_episodes',
          level: 'warning',
          title: `${item.title} — épisodes manquants`,
          description: `La série semble incomplète. Environ ${totalPerSeries} épisode(s) manquant(s) sur la base des stats actuelles.`,
          tags: [`${totalPerSeries} épisodes`],
          mediaType: 'series',
          mediaId: item.id,
          actions: { primary: 'Rechercher épisodes', secondary: 'Ignorer' },
        });
      });
    }
  }

  return issues;
};

const clampIssues = (issues: MediaIssue[], maxByType: number): MediaIssue[] => {
  const counters = new Map<IssueType, number>();
  const result: MediaIssue[] = [];

  issues.forEach((issue) => {
    const count = counters.get(issue.type) || 0;
    if (count >= maxByType) return;
    counters.set(issue.type, count + 1);
    result.push(issue);
  });

  return result;
};

// Keep it for now, in case we want to reintroduce type-based coloring later
// const issueTypeClass = (type: IssueType): string => {
//   if (type === 'missing_file' || type === 'missing_episodes') return styles.typeMissing;
//   if (type === 'duplicate_movies' || type === 'low_bitrate') return styles.typeWarning;
//   if (type === 'incorrect_encoding') return styles.typeNotice;
//   return styles.typeInfo;
// };

const levelTypeClass = (level: IssueLevel): string => {
  if (level === 'error') return styles.levelError;
  if (level === 'warning') return styles.levelWarning;
  if (level === 'info') return styles.levelInfo;
  return styles.levelNotice;
};

const kpiCounts = (issues: MediaIssue[]) => {
  const errors = issues.filter((i) => i.level === 'error').length;
  const warnings = issues.filter((i) => i.level === 'warning').length;
  const infos = issues.filter((i) => i.level === 'info').length;
  const notices = issues.filter((i) => i.level === 'notice').length;
  return { errors, warnings, infos, notices };
};

export const Problems = () => {
  const context = useAppContext();
  const [tab, setTab] = useState<FilterTab>('all');
  const [issues, setIssues] = useState<MediaIssue[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showErrors, setShowErrors] = useState(true);
  const [showWarnings, setShowWarnings] = useState(true);
  const [showInfos, setShowInfos] = useState(true);
  const [showNotices, setShowNotices] = useState(true);

  useEffect(() => {
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError(null);

      try {
        const [missingMoviesRes, moviesRes, seriesRes] = await Promise.all([
          apiClient.getMovies(1, 200, { status: 'missing' }),
          apiClient.getMovies(1, 200, {}),
          apiClient.getSeries(1, 120, {}),
        ]);

        if (!missingMoviesRes.success || !moviesRes.success || !seriesRes.success) {
          throw new Error('Impossible de charger les données des problèmes.');
        }

        const missingMovies = missingMoviesRes.data || [];
        const fullMovies = moviesRes.data || [];
        const allMovies = [...missingMovies, ...fullMovies].filter(
          (movie, index, arr) => arr.findIndex((x) => x.id === movie.id) === index
        );

        const movieIssues = buildMovieIssues(allMovies);
        const seriesIssues = buildSeriesIssues(seriesRes.data || [], context.stats?.missingEpisodes || 0);

        const synthesized = clampIssues([...movieIssues, ...seriesIssues], 8);

        if (!cancelled) {
          setIssues(synthesized);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Erreur inconnue');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    load();

    return () => {
      cancelled = true;
    };
  }, [context.stats?.missingEpisodes]);

  const counts = useMemo(() => kpiCounts(issues), [issues]);

  const filteredIssues = useMemo(() => {
    const currentTab = ISSUE_TABS.find((item) => item.key === tab);
    if (!currentTab || tab === 'all') {
        return issues.filter((issue) => {
          if (issue.level === 'error' && !showErrors) return false;
          if (issue.level === 'warning' && !showWarnings) return false;
          if (issue.level === 'info' && !showInfos) return false;
          if (issue.level === 'notice' && !showNotices) return false;
          return true;
        });
    }
    return issues
        .filter((issue) => currentTab.matches.includes(issue.type))
        .filter((issue) => {
          if (issue.level === 'error' && !showErrors) return false;
          if (issue.level === 'warning' && !showWarnings) return false;
          if (issue.level === 'info' && !showInfos) return false;
          if (issue.level === 'notice' && !showNotices) return false;
          return true;
        });
  }, [issues, tab, showErrors, showWarnings, showInfos, showNotices]);

  const tabCount = (tabKey: FilterTab): number => {
    const entry = ISSUE_TABS.find((item) => item.key === tabKey);
    if (!entry) return 0;
    if (tabKey === 'all') {
        return issues.filter((issue) => {
          if (issue.level === 'error' && !showErrors) return false;
          if (issue.level === 'warning' && !showWarnings) return false;
          if (issue.level === 'info' && !showInfos) return false;
          if (issue.level === 'notice' && !showNotices) return false;
          return true;
        }).length;
    }
    return issues
        .filter((issue) => entry.matches.includes(issue.type))
        .filter((issue) => {
          if (issue.level === 'error' && !showErrors) return false;
          if (issue.level === 'warning' && !showWarnings) return false;
          if (issue.level === 'info' && !showInfos) return false;
          if (issue.level === 'notice' && !showNotices) return false;
          return true;
        }).length;
  };

  return (
    <div className={styles.page}>
      {/* <div className={styles.topbar}>
        <span className={styles.topbarTitle}>{issues.length} problèmes détectés</span>
        <button type="button" className={styles.scanButton}>
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M2 7h10M8 4l3 3-3 3" />
          </svg>
          Lancer un scan
        </button>
      </div> */}

      <div className={styles.summary}>
        <div className={`${styles.kpi} ${styles.kpiGray}`}>
          <input
            className={styles.switchCheckbox}
            type="checkbox"
            name="error-filter"
            checked={showErrors}
            onChange={() => setShowErrors(!showErrors)}
          />

          <div className={styles.kpiLabel}>Erreurs</div>
          <div className={`${styles.kpiValue} ${styles.kpiRed}`}>{counts.errors}</div>
        </div>
        <div className={`${styles.kpi} ${styles.kpiGray}`}>
          <input
            type="checkbox"
            className={styles.switchCheckbox}
            name="warning-filter"
            checked={showWarnings}
            onChange={() => setShowWarnings(!showWarnings)}
          />

          <div className={styles.kpiLabel}>Avertissements</div>
          <div className={`${styles.kpiValue} ${styles.kpiOrange}`}>{counts.warnings}</div>
        </div>
        <div className={`${styles.kpi} ${styles.kpiGray}`}>
          <input
            type="checkbox"
            className={styles.switchCheckbox}
            name="info-filter"
            checked={showInfos}
            onChange={() => setShowInfos(!showInfos)}
          />

          <div className={styles.kpiLabel}>Informations</div>
          <div className={`${styles.kpiValue} ${styles.kpiBlue}`}>{counts.infos}</div>
        </div>
        <div className={`${styles.kpi} ${styles.kpiGray}`}>
          <input
            type="checkbox"
            className={styles.switchCheckbox}
            name="notice-filter"
            checked={showNotices}
            onChange={() => setShowNotices(!showNotices)}
          />

          <div className={styles.kpiLabel}>Remarques</div>
          <div className={`${styles.kpiValue} ${styles.kpiGray}`}>{counts.notices}</div>
        </div>
      </div>

      <div className={styles.tabs}>
        {ISSUE_TABS.map((item) => (
          <button
            key={item.key}
            type="button"
            className={`${styles.tab} ${tab === item.key ? styles.tabActive : ''}`}
            onClick={() => setTab(item.key)}
          >
            {item.label} ({tabCount(item.key)})
          </button>
        ))}
      </div>

      <div className={styles.list}>
        {loading ? (
          <div className={styles.state}>Chargement des problèmes...</div>
        ) : error ? (
          <div className={styles.state}>Erreur: {error}</div>
        ) : filteredIssues.length === 0 ? (
          <div className={styles.state}>Aucun problème pour ce filtre.</div>
        ) : (
          filteredIssues.map((issue) => (
            <article key={issue.id} className={`${styles.card} ${levelClassName[issue.level]}`}>
              <div className={styles.severityIcon}>
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M3 13L8 3l5 10H3z" />
                  <path d="M8 7v3M8 12h.01" />
                </svg>
              </div>

              <div className={styles.body}>
                <div className={styles.title}>{issue.title}</div>
                <div className={styles.description}>{issue.description}</div>
                {issue.path && <div className={styles.path}>{issue.path}</div>}
                <div className={styles.tags}>
                  <span className={`${styles.tag} ${levelTypeClass(issue.level)}`}>{typeLabel[issue.type]}</span>
                  <span className={styles.tagNeutral}>{levelLabel[issue.level]}</span>
                  {issue.tags.map((tag) => (
                    <span key={`${issue.id}-${tag}`} className={styles.tagNeutral}>{tag}</span>
                  ))}
                </div>
              </div>

              <div className={styles.actions}>
                {/* <button type="button" className={`${styles.actionButton} ${styles.actionPrimary}`}>
                  {issue.actions.primary}
                </button>
                <button type="button" className={styles.actionButton}>
                  {issue.actions.secondary}
                </button> */}
                <button type="button" className={styles.actionButton} onClick={() => context.goToPage(issue.mediaType === 'movie' ? 'detail-movie' : 'detail-series', issue.mediaId)}>
                  Voir
                </button>
              </div>
            </article>
          ))
        )}
      </div>
    </div>
  );
};
