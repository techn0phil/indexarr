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

export function useSortPreference(storageKey: string, validKeys: string[], defaultKey: string) {
  const [sort, setSort] = useState<SortState>(() => loadSort(storageKey, validKeys, defaultKey));

  // Selecting the active option inverts the order; a new option starts ascending.
  const selectSort = useCallback(
    (key: string) => {
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
