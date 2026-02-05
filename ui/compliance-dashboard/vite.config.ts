import { defineConfig } from 'vite';
import { viteSingleFile } from 'vite-plugin-singlefile';

export default defineConfig({
  plugins: [viteSingleFile()],
  build: {
    outDir: '../../internal/resources/dist',
    emptyOutDir: true,
    minify: true,
    target: 'es2020',
  },
});
