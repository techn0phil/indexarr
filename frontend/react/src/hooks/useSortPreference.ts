import { useCallback, useState } from 'react';

export type SortOrder = 'asc' | 'desc';

export interface SortState {
  key: string;
  order: SortOrder;
}

export const toSortParam = (sort: SortState): string => (sort.order === 'desc' ? `-${sort.key}` : sort.key);

const loadSort = (storageKey: string, validKeys: string[], defaultKey: string): SortState => {
  const fallback: SortState = { key: defaultKey, order: 'asc' };
  try {
    const saved = JSON.parse(localStorage.getItem(storageKey) ?? 'null');
    if (saved && validKeys.includes(saved.key) && (saved.order === 'asc' || saved.order === 'desc')) {
      return { key: saved.key, order: saved.order };
    }
  } catch {
    // Corrupted value: use default.
  }
  return fallback;
};

// Parses a "?sort=-key" URL param; returns null when absent or not a valid key.
const loadUrlSort = (validKeys: string[]): SortState | null => {
  const raw = new URLSearchParams(window.location.search).get('sort');
  if (!raw) return null;
  const order: SortOrder = raw.startsWith('-') ? 'desc' : 'asc';
  const key = raw.replace(/^-/, '');
  return validKeys.includes(key) ? { key, order } : null;
};

const clearUrlSort = () => {
  const url = new URL(window.location.href);
  if (!url.searchParams.has('sort')) return;
  url.searchParams.delete('sort');
  window.history.replaceState(window.history.state, '', `${url.pathname}${url.search}${url.hash}`);
};

export function useSortPreference(storageKey: string, validKeys: string[], defaultKey: string) {
  // A valid ?sort= URL param overrides the saved preference for this visit only (it is not persisted).
  const [sort, setSort] = useState<SortState>(
    () => loadUrlSort(validKeys) ?? loadSort(storageKey, validKeys, defaultKey)
  );

  // Selecting the active option inverts the order; a new option starts ascending.
  const selectSort = useCallback(
    (key: string) => {
      clearUrlSort();
      setSort((prev) => {
        const next: SortState =
          prev.key === key ? { key, order: prev.order === 'asc' ? 'desc' : 'asc' } : { key, order: 'asc' };
        localStorage.setItem(storageKey, JSON.stringify(next));
        return next;
      });
    },
    [storageKey]
  );

  return { sort, selectSort, isDefault: sort.key === defaultKey && sort.order === 'asc' };
}
