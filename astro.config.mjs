// @ts-check
import { defineConfig } from 'astro/config';
import { loadEnv } from "vite";

const { PORT } = loadEnv(process.env.PORT ?? '', process.cwd(), "");

// https://astro.build/config
export default defineConfig({
  server: {
    port: Number(PORT ?? 4321)
  }
});