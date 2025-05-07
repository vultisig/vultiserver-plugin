/// <reference types="vitest" />
import react from "@vitejs/plugin-react";
import svgr from "vite-plugin-svgr";
import removeAttribute from "react-remove-attr";
import tsconfigPaths from "vite-tsconfig-paths";
import { defineConfig } from "vite";

const IS_PRODUCTION = process.env.NODE_ENV == "production";

export default defineConfig({
  plugins: [
    tsconfigPaths(),
    svgr(),
    IS_PRODUCTION
      ? removeAttribute({
          extensions: ["tsx"],
          attributes: ["data-testid"],
        })
      : null,
    react(),
  ],
  test: {
    globals: true,
    environment: "jsdom",
    setupFiles: "./vitest.setup.ts",
  },
});
