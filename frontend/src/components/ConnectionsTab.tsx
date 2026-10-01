import React, { useState } from 'react';
import { Database, Plus, Check, Trash2, Edit3, ShieldCheck, Loader2 } from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import { TestConnection } from '../../wailsjs/go/main/App';
import { useI18n } from '../i18n/LanguageContext';

interface ConnectionsTabProps {
  connections: models.ConnectionProfile[];
  activeConnectionId: string;
  onSetActive: (id: string) => Promise<void>;
  onDelete: (id: string) => Promise<void>;
  onEdit: (profile: models.ConnectionProfile) => void;
  onNew: () => void;
}

export const ConnectionsTab: React.FC<ConnectionsTabProps> = ({
  connections,
  activeConnectionId,
  onSetActive,
  onDelete,
  onEdit,
  onNew,
}) => {
  const [testingId, setTestingId] = useState<string | null>(null);
  const [testResults, setTestResults] = useState<Record<string, models.ConnectionTestResult>>({});
  const { t } = useI18n();

  const handleTest = async (conn: models.ConnectionProfile) => {
    setTestingId(conn.id);
    try {
      const res = await TestConnection(conn);
      setTestResults((prev) => ({ ...prev, [conn.id]: res }));
    } catch (err: any) {
      setTestResults((prev) => ({
        ...prev,
        [conn.id]: {
          success: false,
          message: err?.message || String(err),
          serverVersion: '',
          latencyMs: 0,
        } as any,
      }));
    } finally {
      setTestingId(null);
    }
  };

  return (
    <div className="h-full flex flex-col p-6 overflow-y-auto">
      {/* Section Header */}
      <div className="flex items-center justify-between pb-4 border-b border-[#2e2e2e]">
        <div>
          <div className="flex items-center gap-2">
            <span className="eyebrow">{t.connections.subsystem}</span>
            <span className="text-[#a7a49c] text-xs font-mono">{t.connections.configuredCount(connections.length)}</span>
          </div>
          <h2 className="text-xl font-black font-mono text-[#f2efe6] tracking-wide mt-1">
            {t.connections.title}
          </h2>
          <p className="text-xs text-[#a7a49c] font-mono mt-0.5">
            {t.connections.description}
          </p>
        </div>

        <button
          onClick={onNew}
          className="brutal-button brutal-button--acid"
        >
          <Plus className="w-4 h-4" />
          <span>{t.connections.newConnectionBtn}</span>
        </button>
      </div>

      {/* Connections Grid */}
      {connections.length === 0 ? (
        <div className="flex-1 flex flex-col items-center justify-center border border-dashed border-[#2e2e2e] bg-[#0c0c0c] my-6 p-12 text-center">
          <Database className="w-12 h-12 text-[#2e2e2e] mb-3" />
          <h3 className="font-mono text-sm font-bold text-[#f2efe6] tracking-wider">
            {t.connections.noConnectionsTitle}
          </h3>
          <p className="text-xs font-mono text-[#a7a49c] max-w-md mt-1 mb-4">
            {t.connections.noConnectionsDesc}
          </p>
          <button
            onClick={onNew}
            className="brutal-button brutal-button--acid"
          >
            <Plus className="w-4 h-4" />
            <span>{t.connections.setupFirstBtn}</span>
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-6">
          {connections.map((conn) => {
            const isActive = conn.id === activeConnectionId;
            const isTesting = testingId === conn.id;
            const test = testResults[conn.id];

            return (
              <div
                key={conn.id}
                className={`brutal-card flex flex-col justify-between p-4 transition-all ${
                  isActive ? 'brutal-card-selected' : ''
                }`}
              >
                {/* Card Header */}
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span className="eyebrow text-[10px]">
                      {conn.isSid ? 'SID' : 'SERVICE'}: {conn.isSid ? conn.sid : conn.serviceName}
                    </span>

                    {isActive ? (
                      <span className="flex items-center gap-1.5 px-2 py-0.5 bg-[#d9ff3f] text-[#050505] text-[10px] font-mono font-black">
                        <Check className="w-3 h-3 text-[#050505]" />
                        <span>{t.connections.activeInMcp}</span>
                      </span>
                    ) : (
                      <button
                        onClick={() => onSetActive(conn.id)}
                        className="text-[11px] font-mono font-bold text-[#a7a49c] hover:text-[#d9ff3f] transition-colors cursor-pointer"
                        title={t.connections.activateTooltip}
                      >
                        {t.connections.activateBtn}
                      </button>
                    )}
                  </div>

                  <h3 className="font-mono text-sm font-bold text-[#f2efe6] tracking-wide truncate">
                    {conn.name}
                  </h3>

                  {/* Telemetry info */}
                  <div className="mt-3 space-y-1.5 text-xs font-mono text-[#a7a49c] bg-[#050505] p-2.5 border border-[#2e2e2e]">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] text-[#a7a49c]">{t.connections.hostPort}</span>
                      <span className="text-[#f2efe6] font-mono">{conn.host}:{conn.port || 1521}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] text-[#a7a49c]">{t.connections.user}</span>
                      <span className="text-[#38bdf8] font-mono">{conn.username}</span>
                    </div>
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] text-[#a7a49c]">{t.connections.encryption}</span>
                      <span className={conn.ssl ? 'text-[#d9ff3f]' : 'text-[#a7a49c]'}>
                        {conn.ssl ? 'TCPS / SSL' : 'PLAIN TCP'}
                      </span>
                    </div>
                  </div>

                  {/* Test Feedback if executed */}
                  {test && (
                    <div
                      className={`mt-2 p-2 border text-[11px] font-mono ${
                        test.success
                          ? 'border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f]'
                          : 'border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185]'
                      }`}
                    >
                      <div className="flex items-center justify-between font-bold">
                        <span>{test.success ? t.connections.testOk : t.connections.testError}</span>
                        <span>{test.latencyMs}ms</span>
                      </div>
                      <div className="text-[10px] truncate mt-0.5">{test.message}</div>
                      {test.serverVersion && (
                        <div className="text-[9px] opacity-80 truncate">{test.serverVersion}</div>
                      )}
                    </div>
                  )}
                </div>

                {/* Card Actions Footer */}
                <div className="flex items-center justify-between pt-3 mt-4 border-t border-[#2e2e2e]">
                  <button
                    onClick={() => handleTest(conn)}
                    disabled={isTesting}
                    className="brutal-button brutal-button--ghost py-1 px-3 min-h-[30px] text-xs"
                  >
                    {isTesting ? (
                      <Loader2 className="w-3 h-3 animate-spin" />
                    ) : (
                      <ShieldCheck className="w-3 h-3 text-[#d9ff3f]" />
                    )}
                    <span>{isTesting ? t.connections.testing : t.connections.testBtn}</span>
                  </button>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => onEdit(conn)}
                      className="p-1.5 text-[#a7a49c] hover:text-[#f2efe6] hover:bg-[#171717] transition-colors"
                      title={t.connections.editTooltip}
                    >
                      <Edit3 className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => onDelete(conn.id)}
                      className="p-1.5 text-[#a7a49c] hover:text-[#e11d48] hover:bg-[#171717] transition-colors"
                      title={t.connections.deleteTooltip}
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
