# Locales

User-facing text is stored here for translation. English is the default and currently the only language.

## Adding a language

1. Copy `en.json` to a new file (e.g. `de.json`, `fr.json`).
2. Translate all values; keep the same keys.
3. In `src/i18n.js`, import the new locale and add it to `messages`:

   ```js
   import en from './locales/en.json'
   import fr from './locales/fr.json'

   export const i18n = createI18n({
     legacy: false,
     locale: 'en',
     fallbackLocale: 'en',
     messages: { en, fr },
   })
   ```

4. To switch locale at runtime, use `i18n.global.locale = 'fr'` (or expose a locale selector that sets it). The app can persist the choice in localStorage if needed.
