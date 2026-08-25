import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    maxWorkers: 1,
    minWorkers: 1,
    fileParallelism: false,
    setupFiles: ['./src/test/setup.ts'],
    css: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'json'],
      include: ['src/**/*.{ts,tsx}'],
      exclude: [
        'node_modules/',
        '.vite/',
        'src/test/',
        'src/**/*.test.{ts,tsx}',
        '**/*.d.ts',
        '**/*.config.*',
      ],
    },
  },
});
