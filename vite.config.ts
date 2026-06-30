import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'cmd/open-todo/web-dist',
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:18080',
    },
  },
});
