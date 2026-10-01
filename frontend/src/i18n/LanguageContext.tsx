import React, { createContext, useContext, useState, useEffect } from 'react';
import { Language, Translations, translations } from './translations';

interface LanguageContextProps {
  language: Language;
  setLanguage: (lang: Language) => void;
  toggleLanguage: () => void;
  t: Translations;
}

const LanguageContext = createContext<LanguageContextProps | null>(null);

const detectLanguage = (): Language => {
  try {
    const saved = localStorage.getItem('dbridge_language');
    if (saved === 'en' || saved === 'es') {
      return saved;
    }
  } catch {}

  // Check system / OS / browser locale
  const candidateLangs: string[] = [];
  if (navigator.languages && navigator.languages.length > 0) {
    candidateLangs.push(...navigator.languages);
  }
  if (navigator.language) {
    candidateLangs.push(navigator.language);
  }

  for (const l of candidateLangs) {
    if (l && l.toLowerCase().startsWith('es')) {
      return 'es';
    }
  }

  // Default is English as requested
  return 'en';
};

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>(detectLanguage);

  const setLanguage = (lang: Language) => {
    setLanguageState(lang);
    try {
      localStorage.setItem('dbridge_language', lang);
    } catch {}
  };

  const toggleLanguage = () => {
    setLanguage(language === 'en' ? 'es' : 'en');
  };

  const t = translations[language];

  return (
    <LanguageContext.Provider value={{ language, setLanguage, toggleLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useI18n = (): LanguageContextProps => {
  const ctx = useContext(LanguageContext);
  if (!ctx) {
    throw new Error('useI18n must be used within a LanguageProvider');
  }
  return ctx;
};
