import { beforeEach, describe, expect, it } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import { toSortParam, useSortPreference } from './useSortPreference';

const KEYS = ['title', 'year'];

describe('useSortPreference', () => {
  beforeEach(() => localStorage.clear());

  it('defaults to the default key in ascending order', () => {
    const { result } = renderHook(() => useSortPreference('test-sort', KEYS, 'title'));

    expect(result.current.sort).toEqual({ key: 'title', order: 'asc' });
    expect(result.current.isDefault).toBe(true);
  });

  it('starts ascending on a new option and inverts when selecting it again', () => {
    const { result } = renderHook(() => useSortPreference('test-sort', KEYS, 'title'));

    act(() => result.current.selectSort('year'));
    expect(result.current.sort).toEqual({ key: 'year', order: 'asc' });

    act(() => result.current.selectSort('year'));
    expect(result.current.sort).toEqual({ key: 'year', order: 'desc' });
    expect(JSON.parse(localStorage.getItem('test-sort') as string)).toEqual({ key: 'year', order: 'desc' });

    act(() => result.current.selectSort('title'));
    expect(result.current.sort).toEqual({ key: 'title', order: 'asc' });
  });

  it('restores a valid saved value and ignores invalid ones', () => {
    localStorage.setItem('test-sort', JSON.stringify({ key: 'year', order: 'desc' }));
    expect(renderHook(() => useSortPreference('test-sort', KEYS, 'title')).result.current.sort).toEqual({
      key: 'year',
      order: 'desc',
    });

    localStorage.setItem('test-sort', JSON.stringify({ key: 'bogus', order: 'desc' }));
    expect(renderHook(() => useSortPreference('test-sort', KEYS, 'title')).result.current.sort).toEqual({
      key: 'title',
      order: 'asc',
    });

    localStorage.setItem('test-sort', 'not json');
    expect(renderHook(() => useSortPreference('test-sort', KEYS, 'title')).result.current.sort).toEqual({
      key: 'title',
      order: 'asc',
    });
  });

  it('serializes descending order with a minus prefix', () => {
    expect(toSortParam({ key: 'year', order: 'asc' })).toBe('year');
    expect(toSortParam({ key: 'year', order: 'desc' })).toBe('-year');
  });
});
