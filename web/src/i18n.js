import { createI18n } from 'vue-i18n'
import en from './locales/en.json'
import fr from './locales/fr.json'

export const defaultLocale = 'en'

function getInitialLocale() {
  try {
    const stored = localStorage.getItem('_flocale')
    if (stored === 'en' || stored === 'fr') return stored
  } catch (_) {}
  return defaultLocale
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLocale(),
  fallbackLocale: defaultLocale,
  messages: {
    en,
    fr,
  },
})
