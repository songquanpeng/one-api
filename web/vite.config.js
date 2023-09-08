import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import envCompatible from 'vite-plugin-env-compatible'
import { createHtmlPlugin } from 'vite-plugin-html'
import { ViteImageOptimizer } from 'vite-plugin-image-optimizer';
import sri from 'vite-plugin-sri';

export default defineConfig({
  plugins: [
    sri(),
    react(),
    ViteImageOptimizer(),
    createHtmlPlugin(),
    envCompatible({
      prefix: "REACT_APP_",
      mountedPath: "process.env",
    }),
  ],
  build: {
    outDir: "build"
  },
  esbuild: {
    loader: "tsx",
    include: [
      "src/**/*.js",
    ],
    exclude: [],
  }
})
