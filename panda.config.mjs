import { defineConfig } from "@pandacss/dev";

export default defineConfig({
  preflight: true,
  include: ["js/**/*.c.ts"],
  syntax: 'object-literal',
  theme: {
    tokens: {
    }
  },
  outdir: "js/styled-system",
  outExtension: 'js'
});
