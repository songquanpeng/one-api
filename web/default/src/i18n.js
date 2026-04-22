import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import zhTranslation from './locales/zh/translation.json';
import enTranslation from './locales/en/translation.json';
import deTranslation from './locales/de/translation.json';
import srTranslation from './locales/sr/translation.json';

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: 'en',
    supportedLngs: ['zh', 'en', 'de', 'sr'],
    nonExplicitSupportedLngs: true,
    debug: process.env.NODE_ENV === 'development',

    interpolation: {
      escapeValue: false,
    },

      resources: {
          zh: {
              translation: zhTranslation
          },
          en: {
              translation: enTranslation
          },
          de: {
              translation: deTranslation
          },
          sr: {
              translation: srTranslation
          }
      }
  });

export default i18n;
