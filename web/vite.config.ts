/// <reference types="vitest/config" />
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vite';
import { VitePWA } from 'vite-plugin-pwa';
import packageJson from './package.json' with { type: 'json' };

// In sviluppo il frontend gira sul server di Vite e le chiamate all'API vanno al backend
// Go, che gira per conto suo. In produzione è il backend a servire entrambi.
const devBackend = process.env.SPESTUBE_DEV_BACKEND ?? 'http://localhost:8080';

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['icon.svg', 'apple-touch-icon.png'],
      manifest: {
        name: 'SpesTube',
        short_name: 'SpesTube',
        description: 'YouTube, solo quello che hai scelto di guardare.',
        lang: 'it',
        start_url: '/',
        scope: '/',
        display: 'standalone',
        orientation: 'portrait',
        theme_color: '#2557C7',
        background_color: '#F3F4F7',
        icons: [
          { src: 'icon-192.png', sizes: '192x192', type: 'image/png' },
          { src: 'icon-512.png', sizes: '512x512', type: 'image/png' },
          {
            src: 'icon-maskable-512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
      },
      workbox: {
        // Il service worker tiene solo l'involucro dell'app. Le risposte dell'API non si
        // mettono in cache: un feed vecchio mostrato come quello di oggi sarebbe peggio di
        // un errore di rete.
        globPatterns: ['**/*.{js,css,html,svg,png,woff2}'],
        navigateFallback: '/index.html',
        navigateFallbackDenylist: [/^\/api\//, /^\/healthz$/],
      },
    }),
  ],
  // Gli alias non si ripetono qui: Vite li legge dai `paths` di tsconfig.app.json.
  resolve: { tsconfigPaths: true },
  base: '/',
  css: {
    modules: { localsConvention: 'camelCaseOnly' },
  },
  define: { __APP_VERSION__: JSON.stringify(packageJson.version) },
  server: {
    proxy: { '/api': devBackend, '/healthz': devBackend },
  },
  test: {
    environment: 'jsdom',
    include: ['tests/**/*.test.{ts,tsx}'],
    setupFiles: ['tests/setup.ts'],
  },
});
