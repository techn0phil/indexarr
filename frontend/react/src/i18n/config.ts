import i18next, { InitOptions } from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import HttpBackend from 'i18next-http-backend';

const SUPPORTED_LANGUAGES = ['en', 'fr', 'de', 'it', 'es'];

const options: InitOptions = {
  fallbackLng: 'en',
  defaultNS: 'common',
  ns: ['common', 'sidebar'],
  detection: {
    order: ['localStorage', 'navigator'],
    caches: ['localStorage'],
    lookupLocalStorage: 'locale-preference',
  },
  backend: {
    loadPath: '/locales/{{lng}}/{{ns}}.json',
  },
  supportedLngs: SUPPORTED_LANGUAGES,
  load: 'languageOnly',
  interpolation: {
    escapeValue: false,
  },
  react: {
    useSuspense: false,
  },
};

i18next
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init(options);

export default i18next;
export { SUPPORTED_LANGUAGES };
