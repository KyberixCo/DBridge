import React, { useState, useEffect } from 'react';
import {
  X,
  CheckCircle2,
  AlertTriangle,
  Loader2,
  Database,
  ShieldCheck,
  Eye,
  EyeOff,
  FolderOpen,
  Sparkles,
  Check,
} from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import {
  TestConnection,
  SelectTNSFile,
  DetectTNSFiles,
  ParseTNSFile,
  SelectDBeaverFile,
  DetectDBeaverFiles,
  ParseDBeaverFile,
} from '../../wailsjs/go/main/App';
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

  // Assisted Import States (TNS & DBeaver)
  const [importMode, setImportMode] = useState<'none' | 'tns' | 'dbeaver'>('none');
  const [detectedTNS, setDetectedTNS] = useState<string[]>([]);
  const [tnsEntries, setTnsEntries] = useState<models.TNSEntry[]>([]);
  const [detectedDBeaver, setDetectedDBeaver] = useState<string[]>([]);
  const [dbeaverProfiles, setDbeaverProfiles] = useState<models.ConnectionProfile[]>([]);
  const [importNotice, setImportNotice] = useState<string>('');

  useEffect(() => {
    if (isOpen) {
      DetectTNSFiles().then((files) => setDetectedTNS(files || [])).catch(() => {});
      DetectDBeaverFiles().then((files) => setDetectedDBeaver(files || [])).catch(() => {});
    }
  }, [isOpen]);

  const handleLoadTNSPath = async (filePath: string) => {
    try {
      const entries = await ParseTNSFile(filePath);
      setTnsEntries(entries || []);
      setImportNotice(t.modal.tnsAliasesLoaded((entries || []).length));
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  const handleBrowseTNS = async () => {
    try {
      const file = await SelectTNSFile();
      if (file) {
        await handleLoadTNSPath(file);
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  const handleApplyTNSEntry = (entry: models.TNSEntry) => {
    setFormData((prev) =>
      models.ConnectionProfile.createFrom({
        ...prev,
        name: prev.name && prev.name !== 'localhost' ? prev.name : entry.alias,
        host: entry.host || prev.host,
        port: entry.port || 1521,
        isSid: entry.isSid,
        serviceName: entry.serviceName || (entry.isSid ? '' : prev.serviceName),
        sid: entry.sid || (entry.isSid ? prev.sid : ''),
        ssl: entry.ssl,
      })
    );
    setImportNotice(`✓ ${entry.alias} (${entry.isSid ? 'SID: ' + entry.sid : 'SERVICE: ' + entry.serviceName})`);
  };

  const handleLoadDBeaverPath = async (filePath: string) => {
    try {
      const profiles = await ParseDBeaverFile(filePath);
      setDbeaverProfiles(profiles || []);
      setImportNotice(t.modal.dbeaverConnsLoaded((profiles || []).length));
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  const handleBrowseDBeaver = async () => {
    try {
      const file = await SelectDBeaverFile();
      if (file) {
        await handleLoadDBeaverPath(file);
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  const handleApplyDBeaverProfile = (p: models.ConnectionProfile) => {
    setFormData((prev) =>
      models.ConnectionProfile.createFrom({
        ...prev,
        name: p.name || prev.name,
        host: p.host || prev.host,
        port: p.port || 1521,
        isSid: p.isSid,
        serviceName: p.serviceName,
        sid: p.sid,
        username: p.username || prev.username,
        ssl: p.ssl,
      })
    );
    setImportNotice(`✓ DBeaver: ${p.name} (${p.isSid ? 'SID: ' + p.sid : 'SERVICE: ' + p.serviceName})`);
  };

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
        <form onSubmit={handleSubmit} className="p-5 space-y-4 text-xs font-mono max-h-[85vh] overflow-y-auto">
          {/* ASSISTED IMPORT SECTION (TNSNAMES.ORA & DBEAVER) */}
          <div className="p-3 bg-[#080808] border border-[#2e2e2e] space-y-2.5">
            <div className="flex items-center justify-between">
              <span className="text-[10px] text-[#a7a49c] font-black uppercase tracking-wider flex items-center gap-1.5">
                <Sparkles className="w-3.5 h-3.5 text-[#d9ff3f]" />
                {t.modal.importToggleLabel}
              </span>
              <div className="flex items-center gap-1.5">
                <button
                  type="button"
                  onClick={() => setImportMode(importMode === 'tns' ? 'none' : 'tns')}
                  className={`px-2 py-0.5 text-[10px] font-bold border transition-colors cursor-pointer ${
                    importMode === 'tns'
                      ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505]'
                      : 'border-[#2e2e2e] bg-[#121212] text-[#a7a49c] hover:text-[#f2efe6]'
                  }`}
                >
                  {t.modal.importTabTns}
                </button>
                <button
                  type="button"
                  onClick={() => setImportMode(importMode === 'dbeaver' ? 'none' : 'dbeaver')}
                  className={`px-2 py-0.5 text-[10px] font-bold border transition-colors cursor-pointer ${
                    importMode === 'dbeaver'
                      ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505]'
                      : 'border-[#2e2e2e] bg-[#121212] text-[#a7a49c] hover:text-[#f2efe6]'
                  }`}
                >
                  {t.modal.importTabDbeaver}
                </button>
              </div>
            </div>

            {/* TNS Tab */}
            {importMode === 'tns' && (
              <div className="pt-2 border-t border-[#1e1e1e] space-y-2 text-[11px]">
                <div className="flex items-center gap-2 flex-wrap">
                  <button
                    type="button"
                    onClick={handleBrowseTNS}
                    className="flex items-center gap-1.5 px-2.5 py-1 bg-[#141414] hover:bg-[#1f1f1f] border border-[#2e2e2e] text-[#f2efe6] font-bold text-[10px] cursor-pointer"
                  >
                    <FolderOpen className="w-3 h-3 text-[#d9ff3f]" />
                    {t.modal.tnsImportBtn}
                  </button>
                  {detectedTNS.length > 0 && (
                    <button
                      type="button"
                      onClick={() => handleLoadTNSPath(detectedTNS[0])}
                      className="text-[10px] text-[#d9ff3f] hover:underline cursor-pointer truncate max-w-[240px]"
                      title={detectedTNS[0]}
                    >
                      {t.modal.tnsDetectedLabel} {detectedTNS[0].split('/').pop()}
                    </button>
                  )}
                </div>

                {tnsEntries.length > 0 && (
                  <div>
                    <select
                      onChange={(e) => {
                        const entry = tnsEntries.find((item) => item.alias === e.target.value);
                        if (entry) handleApplyTNSEntry(entry);
                      }}
                      defaultValue=""
                      className="w-full bg-[#121212] border border-[#d9ff3f] px-2.5 py-1.5 text-[#f2efe6] text-[11px] focus:outline-none cursor-pointer"
                    >
                      <option value="" disabled>
                        {t.modal.tnsSelectAliasPlaceholder}
                      </option>
                      {tnsEntries.map((e) => (
                        <option key={e.alias} value={e.alias}>
                          {e.alias} → {e.host}:{e.port} ({e.isSid ? `SID: ${e.sid}` : `SERVICE: ${e.serviceName}`})
                        </option>
                      ))}
                    </select>
                  </div>
                )}
                <p className="text-[10px] text-[#706e68]">{t.modal.tnsNotice}</p>
              </div>
            )}

            {/* DBeaver Tab */}
            {importMode === 'dbeaver' && (
              <div className="pt-2 border-t border-[#1e1e1e] space-y-2 text-[11px]">
                <div className="flex items-center gap-2 flex-wrap">
                  <button
                    type="button"
                    onClick={handleBrowseDBeaver}
                    className="flex items-center gap-1.5 px-2.5 py-1 bg-[#141414] hover:bg-[#1f1f1f] border border-[#2e2e2e] text-[#f2efe6] font-bold text-[10px] cursor-pointer"
                  >
                    <FolderOpen className="w-3 h-3 text-[#d9ff3f]" />
                    {t.modal.dbeaverImportBtn}
                  </button>
                  {detectedDBeaver.length > 0 && (
                    <button
                      type="button"
                      onClick={() => handleLoadDBeaverPath(detectedDBeaver[0])}
                      className="text-[10px] text-[#d9ff3f] hover:underline cursor-pointer truncate max-w-[240px]"
                      title={detectedDBeaver[0]}
                    >
                      {t.modal.dbeaverDetectedLabel} auto-detect
                    </button>
                  )}
                </div>

                {dbeaverProfiles.length > 0 && (
                  <div>
                    <select
                      onChange={(e) => {
                        const prof = dbeaverProfiles.find((item) => item.id === e.target.value);
                        if (prof) handleApplyDBeaverProfile(prof);
                      }}
                      defaultValue=""
                      className="w-full bg-[#121212] border border-[#d9ff3f] px-2.5 py-1.5 text-[#f2efe6] text-[11px] focus:outline-none cursor-pointer"
                    >
                      <option value="" disabled>
                        {t.modal.dbeaverSelectConnPlaceholder}
                      </option>
                      {dbeaverProfiles.map((p) => (
                        <option key={p.id} value={p.id}>
                          {p.name} → {p.host}:{p.port} ({p.isSid ? `SID: ${p.sid}` : `SERVICE: ${p.serviceName}`})
                        </option>
                      ))}
                    </select>
                  </div>
                )}
                <p className="text-[10px] text-[#706e68]">{t.modal.dbeaverNotice}</p>
              </div>
            )}

            {/* Import Notice */}
            {importNotice && (
              <div className="flex items-center gap-1.5 text-[10px] text-[#d9ff3f] font-mono">
                <Check className="w-3 h-3 shrink-0" />
                <span>{importNotice}</span>
              </div>
            )}
          </div>

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
