import { defineConfig } from "@pandacss/dev";

export default defineConfig({
  preflight: true,
  include: ["js/**/*.c.ts"],
  syntax: 'object-literal',
  theme: {
    extend: {}
  },
  outdir: "js/styled-system",
  outExtension: 'js',
  validation: 'error'
});
