import globals from "globals";
import { defineConfig } from "eslint/config";

export default defineConfig([
  { ignores: ["coverage/**"] },
  { files: ["**/*.{js,mjs,cjs}"], languageOptions: { globals: globals.browser } },
]);
