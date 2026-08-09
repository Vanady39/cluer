import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'
import { defineConfig, globalIgnores } from 'eslint/config'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      js.configs.recommended,
      tseslint.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
    },
    rules: {
      // TODO: raise back to 'error' once the existing violations are fixed.
      // These three are the React Compiler rules from eslint-plugin-react-hooks
      // v7. They flag real anti-patterns, but the current code trips them in
      // Search, Home, AddScenarios and UserMenu, and fixing those means
      // reworking how those components hold state — too much to bundle into
      // turning the linter on. Warnings keep them visible without blocking CI.
      'react-hooks/set-state-in-effect': 'warn',
      'react-hooks/immutability': 'warn',
      'react-hooks/purity': 'warn',
    },
  },
])
