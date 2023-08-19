import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import envCompatible from 'vite-plugin-env-compatible'

export default defineConfig({
  plugins: [
    react(),
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
