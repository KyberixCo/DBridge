import React, { useState } from 'react';
import { Shield, CheckCircle2, XCircle, Plus, X, RotateCcw, Save, Lock } from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import { ValidateQueryTest } from '../../wailsjs/go/main/App';
import { useI18n } from '../i18n/LanguageContext';

interface SecurityTabProps {
  policy: models.SecurityPolicy;
  onSavePolicy: (policy: models.SecurityPolicy) => Promise<void>;
}

export const SecurityTab: React.FC<SecurityTabProps> = ({ policy, onSavePolicy }) => {
  const { t } = useI18n();
  const [currentPolicy, setCurrentPolicy] = useState<models.SecurityPolicy>(
    models.SecurityPolicy.createFrom({
      ...policy,
      blockedKeywords: [...(policy.blockedKeywords || [])],
    })
  );

  const [newKeyword, setNewKeyword] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Live validator playground state
  const [testQuery, setTestQuery] = useState(
    'SELECT id, name FROM employees WHERE status = \'ACTIVE\''
  );
  const [testResult, setTestResult] = useState<{ allowed: boolean; reason: string } | null>(null);

  const handleAddKeyword = () => {
    const kw = newKeyword.trim().toUpperCase();
    if (!kw) return;
    if (!currentPolicy.blockedKeywords.includes(kw)) {
      setCurrentPolicy((prev) =>
        models.SecurityPolicy.createFrom({
          ...prev,
          blockedKeywords: [...prev.blockedKeywords, kw],
        })
      );
    }
    setNewKeyword('');
  };

  const handleRemoveKeyword = (kwToRemove: string) => {
    setCurrentPolicy((prev) =>
      models.SecurityPolicy.createFrom({
        ...prev,
        blockedKeywords: prev.blockedKeywords.filter((k) => k !== kwToRemove),
      })
    );
  };

  const handleResetDefaults = () => {
    setCurrentPolicy(
      models.SecurityPolicy.createFrom({
        mode: 'read_only',
        blockedKeywords: [
          'INSERT', 'UPDATE', 'DELETE', 'DROP', 'ALTER',
          'TRUNCATE', 'CREATE', 'GRANT', 'REVOKE', 'MERGE',
          'RENAME', 'EXEC', 'EXECUTE',
        ],
        allowPlsql: false,
        maxRows: 500,
      })
    );
  };

  const handleSave = async () => {
    setIsSaving(true);
    setSaveSuccess(false);
    try {
      await onSavePolicy(currentPolicy);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err) {
      console.error(err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleRunValidationTest = async () => {
    try {
      const res = await ValidateQueryTest(testQuery);
      setTestResult({ allowed: res.allowed, reason: res.reason });
    } catch (err: any) {
      setTestResult({ allowed: false, reason: err?.message || String(err) });
    }
  };

  return (
    <div className="h-full flex flex-col p-6 overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between pb-4 border-b border-[#2e2e2e]">
        <div>
          <div className="flex items-center gap-2">
            <span className="eyebrow">{t.security.subsystem}</span>
            <span className="text-[10px] font-mono font-bold px-2 py-0.5 bg-[#d9ff3f] text-[#050505]">
              [KYBERIX FIREWALL ACTIVE]
            </span>
          </div>
          <h2 className="text-xl font-black font-mono text-[#f2efe6] tracking-wide mt-1">
            {t.security.title}
          </h2>
          <p className="text-xs text-[#a7a49c] font-mono mt-0.5">
            {t.security.description}
          </p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={handleResetDefaults}
            className="brutal-button brutal-button--ghost text-xs"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            <span>{t.security.resetDefaultsBtn}</span>
          </button>

          <button
            onClick={handleSave}
            disabled={isSaving}
            className="brutal-button brutal-button--acid text-xs"
          >
            <Save className="w-3.5 h-3.5" />
            <span>{isSaving ? t.security.savingPolicyBtn : t.security.savePolicyBtn}</span>
          </button>
        </div>
      </div>

      {saveSuccess && (
        <div className="mt-4 p-3 bg-[#d9ff3f]/15 border border-[#d9ff3f] text-[#d9ff3f] font-mono text-xs flex items-center gap-2">
          <CheckCircle2 className="w-4 h-4" />
          <span className="font-bold">{t.security.savedSuccess}</span>
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-6">
        {/* Left Column: Security Modes & PL/SQL toggles */}
        <div className="lg:col-span-2 space-y-6">
          {/* Policy Mode Selector Card */}
          <div className="brutal-card p-5">
            <span className="eyebrow">{t.security.policyModeTitle}</span>
            <h3 className="font-mono text-sm font-bold text-[#f2efe6] mt-1 mb-3">
              {t.security.policyModeDesc}
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {/* Option 1: Read-Only (Default) */}
              <div
                onClick={() =>
                  setCurrentPolicy((p) => models.SecurityPolicy.createFrom({ ...p, mode: 'read_only' }))
                }
                className={`p-3.5 border cursor-pointer transition-all ${
                  currentPolicy.mode === 'read_only'
                    ? 'border-[#d9ff3f] bg-[#0f1408] shadow-[3px_3px_0_#d9ff3f]'
                    : 'border-[#2e2e2e] bg-[#0c0c0c] hover:border-[#4a4a4a]'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-mono text-xs font-bold text-[#f2efe6]">{t.security.modeReadOnly}</span>
                  <span className="text-[9px] font-mono font-bold px-1.5 py-0.2 bg-[#d9ff3f] text-[#050505]">DEFAULT</span>
                </div>
                <p className="text-[11px] font-mono text-[#a7a49c] leading-relaxed">
                  {t.security.modeReadOnlyDesc}
                </p>
              </div>

              {/* Option 2: Permissive */}
              <div
                onClick={() =>
                  setCurrentPolicy((p) => models.SecurityPolicy.createFrom({ ...p, mode: 'permissive' }))
                }
                className={`p-3.5 border cursor-pointer transition-all ${
                  currentPolicy.mode === 'permissive' || currentPolicy.mode === 'full'
                    ? 'border-[#e11d48] bg-[#1a080c] shadow-[3px_3px_0_#e11d48]'
                    : 'border-[#2e2e2e] bg-[#0c0c0c] hover:border-[#4a4a4a]'
                }`}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="font-mono text-xs font-bold text-[#f2efe6]">{t.security.modePermissive}</span>
                  <span className="text-[9px] font-mono font-bold px-1.5 py-0.2 bg-[#e11d48] text-[#ffffff]">WRITE</span>
                </div>
                <p className="text-[11px] font-mono text-[#a7a49c] leading-relaxed">
                  {t.security.modePermissiveDesc}
                </p>
              </div>
            </div>
          </div>

          {/* Blocked Keywords Manager Card */}
          <div className="brutal-card p-5">
            <div className="flex items-center justify-between mb-2">
              <div>
                <span className="eyebrow">{t.security.blockedKeywordsTitle}</span>
                <h3 className="font-mono text-sm font-bold text-[#f2efe6] mt-1">
                  {t.security.blockedKeywordsTitle}
                </h3>
              </div>
              <span className="text-xs font-mono text-[#a7a49c]">
                [{currentPolicy.blockedKeywords.length}]
              </span>
            </div>
            <p className="text-xs font-mono text-[#a7a49c] mb-4">
              {t.security.blockedKeywordsDesc}
            </p>

            {/* Keyword Chips */}
            <div className="flex flex-wrap gap-2 p-3 bg-[#050505] border border-[#2e2e2e] min-h-[90px] mb-3">
              {currentPolicy.blockedKeywords.map((kw) => (
                <span
                  key={kw}
                  className="inline-flex items-center gap-1.5 px-2 py-1 text-xs font-mono font-bold bg-[#121212] text-[#fb7185] border border-[#e11d48]/50"
                >
                  <Lock className="w-2.5 h-2.5 text-[#e11d48]" />
                  <span>{kw}</span>
                  <button
                    onClick={() => handleRemoveKeyword(kw)}
                    className="hover:text-white transition-colors cursor-pointer ml-1"
                    title={t.common.delete}
                  >
                    <X className="w-3 h-3" />
                  </button>
                </span>
              ))}
            </div>

            {/* Add new keyword input */}
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={newKeyword}
                onChange={(e) => setNewKeyword(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddKeyword()}
                placeholder={t.security.addKeywordPlaceholder}
                className="flex-1 bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-xs font-mono text-[#f2efe6] focus:border-[#d9ff3f] focus:outline-none uppercase"
              />
              <button
                onClick={handleAddKeyword}
                className="brutal-button brutal-button--acid text-xs min-h-[36px]"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>{t.security.addBtn}</span>
              </button>
            </div>
          </div>

          {/* PL/SQL Permission & Max Rows */}
          <div className="brutal-card p-5">
            <span className="eyebrow">{t.security.plsqlSwitchTitle}</span>
            <h3 className="font-mono text-sm font-bold text-[#f2efe6] mt-1 mb-4">
              {t.security.plsqlSwitchTitle}
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* PL/SQL Toggle */}
              <div className="p-3.5 bg-[#050505] border border-[#2e2e2e] flex items-start gap-3">
                <input
                  type="checkbox"
                  id="allowPlsqlToggle"
                  checked={currentPolicy.allowPlsql}
                  onChange={(e) =>
                    setCurrentPolicy((p) =>
                      models.SecurityPolicy.createFrom({ ...p, allowPlsql: e.target.checked })
                    )
                  }
                  className="mt-1 accent-[#d9ff3f] cursor-pointer"
                />
                <div>
                  <label
                    htmlFor="allowPlsqlToggle"
                    className="font-mono text-xs font-bold text-[#f2efe6] cursor-pointer block"
                  >
                    {t.security.plsqlSwitchTitle}
                  </label>
                  <p className="text-[11px] font-mono text-[#a7a49c] mt-1 leading-relaxed">
                    {t.security.plsqlSwitchDesc}
                  </p>
                </div>
              </div>

              {/* Max Rows Limit */}
              <div className="p-3.5 bg-[#050505] border border-[#2e2e2e]">
                <label className="font-mono text-xs font-bold text-[#f2efe6] block mb-1">
                  {t.security.maxRowsTitle}
                </label>
                <input
                  type="number"
                  value={currentPolicy.maxRows}
                  onChange={(e) =>
                    setCurrentPolicy((p) =>
                      models.SecurityPolicy.createFrom({
                        ...p,
                        maxRows: parseInt(e.target.value) || 100,
                      })
                    )
                  }
                  min={10}
                  max={5000}
                  className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-1.5 text-xs font-mono text-[#f2efe6] focus:border-[#d9ff3f] focus:outline-none"
                />
                <p className="text-[10px] font-mono text-[#a7a49c] mt-1.5">
                  {t.security.maxRowsDesc}
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column: Live Query Validator Playground */}
        <div className="space-y-4">
          <div className="brutal-card p-5">
            <span className="eyebrow">{t.security.validatorTitle}</span>
            <h3 className="font-mono text-sm font-bold text-[#f2efe6] mt-1 mb-2">
              {t.security.validatorTitle}
            </h3>
            <p className="text-xs font-mono text-[#a7a49c] mb-3">
              {t.security.validatorDesc}
            </p>

            <textarea
              value={testQuery}
              onChange={(e) => setTestQuery(e.target.value)}
              placeholder={t.security.testQueryPlaceholder}
              rows={5}
              className="w-full bg-[#000000] border border-[#2e2e2e] p-3 text-xs font-mono text-[#f2efe6] focus:border-[#d9ff3f] focus:outline-none resize-none leading-relaxed"
            />

            <div className="flex items-center justify-between mt-3">
              {/* Preset quick test buttons */}
              <div className="flex items-center gap-2 text-[10px] font-mono">
                <button
                  type="button"
                  onClick={() => setTestQuery('DELETE FROM EMPLOYEES WHERE ID = 10')}
                  className="text-[#fb7185] hover:underline font-bold"
                >
                  [ TEST DELETE ]
                </button>
                <button
                  type="button"
                  onClick={() => setTestQuery('SELECT * FROM DUAL')}
                  className="text-[#d9ff3f] hover:underline font-bold"
                >
                  [ TEST SELECT ]
                </button>
              </div>

              <button
                type="button"
                onClick={handleRunValidationTest}
                className="brutal-button brutal-button--acid text-xs min-h-[32px] py-1 px-3"
              >
                {t.security.testQueryBtn}
              </button>
            </div>

            {/* Test result feedback */}
            {testResult && (
              <div
                className={`mt-4 p-3 border font-mono text-xs ${
                  testResult.allowed
                    ? 'border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f]'
                    : 'border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185]'
                }`}
              >
                <div className="flex items-center gap-1.5 font-bold mb-1">
                  {testResult.allowed ? (
                    <CheckCircle2 className="w-4 h-4 text-[#d9ff3f]" />
                  ) : (
                    <XCircle className="w-4 h-4 text-[#e11d48]" />
                  )}
                  <span>{testResult.allowed ? t.security.queryPermitted : t.security.queryBlocked}</span>
                </div>
                {testResult.reason && (
                  <div className="text-[11px] opacity-90 mt-1">{testResult.reason}</div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
