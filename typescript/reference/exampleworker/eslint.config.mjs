import eslint from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["dist/**", "coverage/**", "node_modules/**"] },
  eslint.configs.recommended,
  {
    files: ["**/*.ts"],
    extends: [...tseslint.configs.strictTypeChecked, ...tseslint.configs.stylisticTypeChecked],
    languageOptions: {
      globals: globals.nodeBuiltin,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "@typescript-eslint/consistent-type-imports": "error",
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-non-null-assertion": "error",
    },
  },
  {
    files: ["src/core/**/*.ts"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            { name: "fastify", message: "core must not import HTTP" },
            { name: "pino", message: "core must not import logging" },
          ],
          patterns: [
            {
              group: ["../health/**", "../messaging/**", "../telemetry/**"],
              message: "dependency direction points toward core",
            },
          ],
        },
      ],
    },
  },
  {
    files: ["**/*.cjs"],
    languageOptions: { globals: globals.node },
  },
  {
    files: ["**/*.mjs"],
    languageOptions: { globals: globals.nodeBuiltin },
  },
  eslintConfigPrettier,
);
