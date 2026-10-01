import React, { useState } from 'react';
import { X, CheckCircle2, AlertTriangle, Loader2, Database, ShieldCheck, Eye, EyeOff } from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import { TestConnection } from '../../wailsjs/go/main/App';
import { useI18n } from '../i18n/LanguageContext';

interface ConnectionModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSave: (profile: models.ConnectionProfile) => Promise<void>;
  initialProfile?: models.ConnectionProfile | null;
}

export const ConnectionModal: React.FC<ConnectionModalProps> = ({
  isOpen,
  onClose,
  onSave,
  initialProfile,
}) => {
  const { t } = useI18n();
  const [formData, setFormData] = useState<models.ConnectionProfile>(
    initialProfile || {
      id: '',
      name: '',
      host: 'localhost',
      port: 1521,
      serviceName: 'XEPDB1',
      sid: '',
      isSid: false,
      username: '',
      password: '',
      ssl: false,
      walletPath: '',
      createdAt: null,
      updatedAt: null,
    } as any
  );

  const [showPassword, setShowPassword] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<models.ConnectionTestResult | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');

  if (!isOpen) return null;

  const handleChange = (field: keyof models.ConnectionProfile, val: any) => {
    setFormData((prev) => models.ConnectionProfile.createFrom({ ...prev, [field]: val }));
    setTestResult(null);
    setErrorMsg('');
  };

  const handleTest = async () => {
    if (!formData.host || !formData.username) {
      setErrorMsg(t.modal.hostUserRequiredError);
      return;
    }
    setIsTesting(true);
    setTestResult(null);
    setErrorMsg('');
    try {
      const res = await TestConnection(formData);
      setTestResult(res);
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    } finally {
      setIsTesting(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name || !formData.host || !formData.username) {
      setErrorMsg(t.modal.requiredFieldsError);
      return;
    }
    setIsSaving(true);
    setErrorMsg('');
    try {
      await onSave(formData);
      onClose();
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-xs p-4">
      <div className="brutal-card w-full max-w-lg bg-[#0c0c0c] border-2 border-[#2e2e2e] text-[#f2efe6] shadow-[8px_8px_0_#050505]">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-[#2e2e2e] bg-[#050505]">
          <div className="flex items-center gap-2">
            <span className="eyebrow text-xs font-black">
              {formData.id ? t.modal.editTitle : t.modal.newTitle}
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-[#a7a49c] hover:text-white transition-colors cursor-pointer"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-5 space-y-4 text-xs font-mono">
          {/* Profile Name */}
          <div>
            <label className="block text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider mb-1">
              {t.modal.nameLabel}
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => handleChange('name', e.target.value)}
              placeholder={t.modal.namePlaceholder}
              className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
              required
            />
          </div>

          {/* Host & Port */}
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <label className="block text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider mb-1">
                {t.modal.hostLabel}
              </label>
              <input
                type="text"
                value={formData.host}
                onChange={(e) => handleChange('host', e.target.value)}
                placeholder="localhost o 127.0.0.1"
                className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider mb-1">
                {t.modal.portLabel}
              </label>
              <input
                type="number"
                value={formData.port}
                onChange={(e) => handleChange('port', parseInt(e.target.value) || 1521)}
                className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                required
              />
            </div>
          </div>

          {/* Service Name vs SID toggle */}
          <div>
            <div className="flex items-center justify-between mb-1">
              <label className="text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider">
                {t.modal.idTypeLabel}
              </label>
              <div className="flex items-center gap-2 text-[10px]">
                <button
                  type="button"
                  onClick={() => handleChange('isSid', false)}
                  className={`px-2 py-0.5 border ${
                    !formData.isSid
                      ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505] font-black'
                      : 'border-[#2e2e2e] text-[#a7a49c]'
                  }`}
                >
                  {t.modal.serviceNameOption}
                </button>
                <button
                  type="button"
                  onClick={() => handleChange('isSid', true)}
                  className={`px-2 py-0.5 border ${
                    formData.isSid
                      ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505] font-black'
                      : 'border-[#2e2e2e] text-[#a7a49c]'
                  }`}
                >
                  {t.modal.sidOption}
                </button>
              </div>
            </div>

            {!formData.isSid ? (
              <input
                type="text"
                value={formData.serviceName}
                onChange={(e) => handleChange('serviceName', e.target.value)}
                placeholder="Ej: FREEPDB1, XEPDB1, ORCL"
                className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                required
              />
            ) : (
              <input
                type="text"
                value={formData.sid}
                onChange={(e) => handleChange('sid', e.target.value)}
                placeholder="Ej: XE, ORCL, FREE"
                className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                required
              />
            )}
          </div>

          {/* Username & Password */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider mb-1">
                {t.modal.usernameLabel}
              </label>
              <input
                type="text"
                value={formData.username}
                onChange={(e) => handleChange('username', e.target.value)}
                placeholder="SYSTEM / oramcp / HR"
                className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                required
              />
            </div>
            <div>
              <label className="block text-[11px] text-[#a7a49c] uppercase font-bold tracking-wider mb-1">
                {t.modal.passwordLabel}
              </label>
              <div className="relative">
                <input
                  type={showPassword ? 'text' : 'password'}
                  value={formData.password}
                  onChange={(e) => handleChange('password', e.target.value)}
                  placeholder="••••••••"
                  className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-2 pr-8 text-[#f2efe6] text-xs focus:border-[#d9ff3f] focus:outline-none"
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-2 text-[#a7a49c] hover:text-white cursor-pointer"
                >
                  {showPassword ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                </button>
              </div>
            </div>
          </div>

          <div className="text-[10px] font-mono text-[#a7a49c] bg-[#141414] p-2 border border-[#262626]">
            {t.modal.passwordVaultNotice}
          </div>

          {/* SSL Toggle */}
          <div className="flex items-center gap-2 pt-1">
            <input
              type="checkbox"
              id="sslToggle"
              checked={formData.ssl}
              onChange={(e) => handleChange('ssl', e.target.checked)}
              className="accent-[#d9ff3f] cursor-pointer"
            />
            <label htmlFor="sslToggle" className="text-xs text-[#f2efe6] cursor-pointer">
              {t.modal.sslLabel}
            </label>
          </div>

          {/* Test Feedback Area */}
          {testResult && (
            <div
              className={`p-3 border text-xs ${
                testResult.success
                  ? 'border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f]'
                  : 'border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185]'
              }`}
            >
              <div className="flex items-center gap-2 font-bold mb-1">
                {testResult.success ? (
                  <CheckCircle2 className="w-4 h-4 text-[#d9ff3f]" />
                ) : (
                  <AlertTriangle className="w-4 h-4 text-[#e11d48]" />
                )}
                <span>{testResult.success ? t.connections.testOk : t.connections.testError}</span>
                <span className="ml-auto text-[10px] opacity-80">{testResult.latencyMs}ms</span>
              </div>
              <div className="text-[11px] font-mono whitespace-pre-wrap">{testResult.message}</div>
              {testResult.serverVersion && (
                <div className="text-[10px] text-[#f2efe6]/80 mt-1 truncate">
                  Banner: {testResult.serverVersion}
                </div>
              )}
            </div>
          )}

          {errorMsg && (
            <div className="p-2 border border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185] text-xs">
              {errorMsg}
            </div>
          )}

          {/* Modal Actions */}
          <div className="flex items-center justify-between pt-3 border-t border-[#2e2e2e]">
            <button
              type="button"
              onClick={handleTest}
              disabled={isTesting}
              className="brutal-button brutal-button--purple text-xs min-h-[34px] py-1 px-3"
            >
              {isTesting ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  <span>{t.modal.testingBtn}</span>
                </>
              ) : (
                <>
                  <ShieldCheck className="w-3.5 h-3.5" />
                  <span>{t.modal.testBtn}</span>
                </>
              )}
            </button>

            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={onClose}
                className="brutal-button brutal-button--ghost text-xs min-h-[34px] py-1 px-3"
              >
                {t.common.cancel}
              </button>
              <button
                type="submit"
                disabled={isSaving}
                className="brutal-button brutal-button--acid text-xs min-h-[34px] py-1 px-3"
              >
                {isSaving ? t.modal.savingBtn : t.modal.saveBtn}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};
