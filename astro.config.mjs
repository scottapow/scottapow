// @ts-check
import { defineConfig } from 'astro/config';
import { loadEnv } from "vite";
import node from '@astrojs/node';

const { PORT } = loadEnv(process.env.PORT ?? '', process.cwd(), "");

// https://astro.build/config
export default defineConfig({
  adapter: node({
    mode: 'standalone',
  }),
  server: {
    port: Number(PORT ?? 4321)
  }
});