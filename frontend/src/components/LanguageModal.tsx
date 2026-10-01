import React from 'react';
import { X, Languages, Check, Globe } from 'lucide-react';
import { useI18n } from '../i18n/LanguageContext';
import { Language } from '../i18n/translations';
import { BracketBadge } from './atoms/BracketBadge';

interface LanguageModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const LanguageModal: React.FC<LanguageModalProps> = ({ isOpen, onClose }) => {
  const { language, setLanguage, t } = useI18n();

  if (!isOpen) return null;

  const handleSelect = (lang: Language) => {
    setLanguage(lang);
    onClose();
  };

  const options: { code: Language; name: string; desc: string; tag: string }[] = [
    {
      code: 'en',
      name: t.languageModal.englishTitle,
      desc: t.languageModal.englishDesc,
      tag: 'EN / US-INTL',
    },
    {
      code: 'es',
      name: t.languageModal.spanishTitle,
      desc: t.languageModal.spanishDesc,
      tag: 'ES / LATAM-ES',
    },
  ];

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-xs p-4">
      <div className="brutal-card w-full max-w-md bg-[#0c0c0c] border-2 border-[#2e2e2e] text-[#f2efe6] shadow-[10px_10px_0_#050505] flex flex-col">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-[#2e2e2e] bg-[#050505] shrink-0">
          <div className="flex items-center gap-2">
            <Languages className="w-4 h-4 text-[#d9ff3f]" />
            <span className="eyebrow text-xs font-black">
              {t.languageModal.title}
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-[#a7a49c] hover:text-white transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="p-6 space-y-4 font-mono text-xs">
          <p className="text-[#a7a49c] text-[11px] leading-relaxed">
            {t.languageModal.description}
          </p>

          <div className="space-y-3">
            {options.map((opt) => {
              const isSelected = language === opt.code;
              return (
                <button
                  key={opt.code}
                  type="button"
                  onClick={() => handleSelect(opt.code)}
                  className={`w-full text-left p-4 border transition-all cursor-pointer flex items-center justify-between gap-3 ${
                    isSelected
                      ? 'border-[#d9ff3f] bg-[#141414] shadow-[4px_4px_0_#d9ff3f]'
                      : 'border-[#2e2e2e] bg-[#050505] hover:border-[#444] hover:bg-[#0c0c0c]'
                  }`}
                >
                  <div className="space-y-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-sm text-[#f2efe6] uppercase">
                        {opt.name}
                      </span>
                      <span className="text-[10px] text-[#706e68]">
                        [{opt.tag}]
                      </span>
                    </div>
                    <div className="text-[11px] text-[#a7a49c] truncate">
                      {opt.desc}
                    </div>
                  </div>

                  <div className="shrink-0 flex items-center">
                    {isSelected ? (
                      <div className="flex items-center gap-1.5 px-2 py-1 bg-[#d9ff3f] text-[#050505] font-black text-[10px]">
                        <Check className="w-3.5 h-3.5" />
                        <span>{t.languageModal.activeBadge}</span>
                      </div>
                    ) : (
                      <div className="w-4 h-4 border border-[#383838] flex items-center justify-center" />
                    )}
                  </div>
                </button>
              );
            })}
          </div>

          <div className="p-3 bg-[#050505] border border-[#1e1e1e] flex items-center gap-2 text-[#706e68] text-[10px]">
            <Globe className="w-3.5 h-3.5 shrink-0 text-[#d9ff3f]" />
            <span>{t.languageModal.autoDetectNotice}</span>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="flex justify-end px-5 py-3 border-t border-[#2e2e2e] bg-[#050505]">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-1.5 border border-[#2e2e2e] bg-[#141414] hover:bg-[#1f1f1f] text-[#f2efe6] font-mono font-bold text-xs uppercase cursor-pointer"
          >
            {t.languageModal.closeBtn}
          </button>
        </div>
      </div>
    </div>
  );
};
