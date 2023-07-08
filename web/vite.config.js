import fs from "node:fs";
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import * as esbuild from "esbuild";

const sourceJSPattern = /\/src\/.*\.js$/;
const rollupPlugin = (matchers) => ({
  name: "js-in-jsx",
  load(id) {
    if (matchers.some(matcher => matcher.test(id))) {
      const file = fs.readFileSync(id, { encoding: "utf-8" });
      return esbuild.transformSync(file, { loader: "jsx" });
    }
  }
});

export default defineConfig({
  plugins: [
    react()
  ],
  build: {
    outDir: "./build",
    rollupOptions: {
      plugins: [
        rollupPlugin([sourceJSPattern])
      ],
    },
    commonjsOptions: {
      transformMixedEsModules: true,
    },
  },
  optimizeDeps: {
    esbuildOptions: {
      loader: {
        ".js": "jsx",
      },
    },
  },
  esbuild: {
    loader: "jsx",
    include: [sourceJSPattern],
    exclude: [],
  },
});
