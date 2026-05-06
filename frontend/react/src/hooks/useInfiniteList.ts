import { useState, useCallback, useEffect, useRef } from 'react';
import { PaginatedResponse } from '../types';

interface UseInfiniteListOptions<T> {
  fetchFn: (page: number, pageSize: number, filters: Record<string, string>) => Promise<PaginatedResponse<T>>;
  pageSize?: number;
  filters?: Record<string, string>;
}

interface UseInfiniteListReturn<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  loadMore: () => void;
  reset: () => void;
}

export function useInfiniteList<T>({
  fetchFn,
  pageSize = 50,
  filters = {},
}: UseInfiniteListOptions<T>): UseInfiniteListReturn<T> {
  const [items, setItems] = useState<T[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isInitialMount = useRef(true);

  // Derive hasMore from items length vs total
  const hasMore = items.length < total;

  // Load a page of data
  const loadPage = useCallback(
    async (pageNum: number, append: boolean = true) => {
      setLoading(true);
      setError(null);
      try {
        const response = await fetchFn(pageNum, pageSize, filters);
        
        if (!response.success) {
          throw new Error(response.error || 'Failed to fetch data');
        }

        setTotal(response.total);
        
        if (append && pageNum > 1) {
          // Append to existing items
          setItems((prev) => [...prev, ...(response.data || [])]);
        } else {
          // Replace items (initial load or reset)
          setItems(response.data || []);
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Unknown error';
        setError(errorMessage);
        console.error('Failed to load page:', err);
      } finally {
        setLoading(false);
      }
    },
    [fetchFn, pageSize, filters]
  );

  // Load next page
  const loadMore = useCallback(() => {
    if (!loading && hasMore) {
      const nextPage = page + 1;
      setPage(nextPage);
      loadPage(nextPage, true);
    }
  }, [loading, hasMore, page, loadPage]);

  // Reset to page 1 and clear items
  const reset = useCallback(() => {
    setItems([]);
    setPage(1);
    setTotal(0);
    setError(null);
    loadPage(1, false);
  }, [loadPage]);

  // Initial load on mount or when filters change
  useEffect(() => {
    if (isInitialMount.current) {
      isInitialMount.current = false;
      loadPage(1, false);
    } else {
      // Filters changed, reset to page 1
      reset();
    }
  }, [filters, reset, loadPage]);

  return {
    items,
    page,
    pageSize,
    total,
    loading,
    error,
    hasMore,
    loadMore,
    reset,
  };
}
