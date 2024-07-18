import { defineConfig } from "@pandacss/dev";

export default defineConfig({
  preflight: true,
  include: ["./components/**/*.c.ts"],
  syntax: 'object-literal',
  theme: {
    tokens: {
    }
  },
  outdir: "styled-system",
});
