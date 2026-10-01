import React, { useState } from 'react';
import { Activity, Trash2, X, AlertCircle } from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import { useI18n } from '../i18n/LanguageContext';

interface AuditTabProps {
  logs: models.AuditLogEntry[];
  onClearLogs: () => Promise<void>;
}

export const AuditTab: React.FC<AuditTabProps> = ({ logs, onClearLogs }) => {
  const { t } = useI18n();
  const [filterStatus, setFilterStatus] = useState<'all' | 'allowed' | 'blocked'>('all');
  const [search, setSearch] = useState('');
  const [selectedEntry, setSelectedEntry] = useState<models.AuditLogEntry | null>(null);

  const filteredLogs = logs.filter((log) => {
    if (filterStatus === 'allowed' && !log.allowed) return false;
    if (filterStatus === 'blocked' && log.allowed) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        log.query?.toLowerCase().includes(q) ||
        log.toolName?.toLowerCase().includes(q) ||
        log.clientInfo?.toLowerCase().includes(q)
      );
    }
    return true;
  });

  const totalBlocked = logs.filter((l) => !l.allowed).length;
  const totalAllowed = logs.filter((l) => l.allowed).length;

  return (
    <div className="h-full flex flex-col p-6 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between pb-4 border-b border-[#2e2e2e] shrink-0">
        <div>
          <div className="flex items-center gap-2">
            <span className="eyebrow">{t.audit.subsystem}</span>
            <span className="text-[10px] font-mono font-bold px-1.5 py-0.5 bg-[#a855f7] text-[#050505]">
              [MCP LIVE STREAM]
            </span>
          </div>
          <h2 className="text-xl font-black font-mono text-[#f2efe6] tracking-wide mt-1">
            {t.audit.title}
          </h2>
          <p className="text-xs text-[#a7a49c] font-mono mt-0.5">
            {t.audit.description}
          </p>
        </div>

        {/* Counters & Actions */}
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 font-mono text-xs">
            <span className="px-2 py-0.5 bg-[#101010] border border-[#d9ff3f] text-[#d9ff3f] text-[11px] font-bold">
              {t.audit.allowedCount(totalAllowed)}
            </span>
            <span className="px-2 py-0.5 bg-[#101010] border border-[#e11d48] text-[#fb7185] text-[11px] font-bold">
              {t.audit.blockedCount(totalBlocked)}
            </span>
          </div>

          <button
            onClick={() => {
              if (confirm(t.audit.confirmClear)) {
                onClearLogs();
              }
            }}
            className="brutal-button brutal-button--crimson py-1 px-3 min-h-[32px] text-xs"
          >
            <Trash2 className="w-3.5 h-3.5" />
            <span>{t.audit.clearLogsBtn}</span>
          </button>
        </div>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex items-center justify-between py-3 border-b border-[#2e2e2e] shrink-0 font-mono text-xs">
        <div className="flex items-center gap-2">
          <span className="text-[#a7a49c] text-[10px] uppercase font-bold">{t.audit.filterLabel}</span>
          <button
            onClick={() => setFilterStatus('all')}
            className={`px-2 py-0.5 border text-xs font-bold cursor-pointer ${
              filterStatus === 'all'
                ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505]'
                : 'border-[#2e2e2e] text-[#a7a49c] hover:text-[#f2efe6]'
            }`}
          >
            {t.audit.filterAll(logs.length)}
          </button>
          <button
            onClick={() => setFilterStatus('allowed')}
            className={`px-2 py-0.5 border text-xs font-bold cursor-pointer ${
              filterStatus === 'allowed'
                ? 'border-[#d9ff3f] bg-[#d9ff3f] text-[#050505]'
                : 'border-[#2e2e2e] text-[#a7a49c] hover:text-[#f2efe6]'
            }`}
          >
            {t.audit.filterAllowed(totalAllowed)}
          </button>
          <button
            onClick={() => setFilterStatus('blocked')}
            className={`px-2 py-0.5 border text-xs font-bold cursor-pointer ${
              filterStatus === 'blocked'
                ? 'border-[#e11d48] bg-[#e11d48] text-[#ffffff]'
                : 'border-[#2e2e2e] text-[#a7a49c] hover:text-[#f2efe6]'
            }`}
          >
            {t.audit.filterBlocked(totalBlocked)}
          </button>
        </div>

        <div className="w-72">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t.audit.searchPlaceholder}
            className="w-full bg-[#121212] border border-[#2e2e2e] px-3 py-1 text-xs text-[#f2efe6] font-mono focus:border-[#d9ff3f] focus:outline-none"
          />
        </div>
      </div>

      {/* Main Table / Detail Split View */}
      <div className="flex-1 flex overflow-hidden mt-3 gap-4">
        {/* Table Area */}
        <div className="flex-1 overflow-auto border border-[#2e2e2e] bg-[#0c0c0c]">
          {filteredLogs.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center p-8 text-[#a7a49c] font-mono text-xs text-center">
              <Activity className="w-8 h-8 text-[#2e2e2e] mb-2" />
              <span>{t.audit.emptyLogs}</span>
            </div>
          ) : (
            <table className="w-full text-left font-mono text-xs border-collapse">
              <thead className="bg-[#101010] sticky top-0 z-10 border-b border-[#2e2e2e]">
                <tr>
                  <th className="py-2 px-3 text-[#a7a49c] text-[10px] uppercase font-bold border-r border-[#2e2e2e] w-28">
                    {t.audit.colStatus}
                  </th>
                  <th className="py-2 px-3 text-[#a7a49c] text-[10px] uppercase font-bold border-r border-[#2e2e2e] w-40">
                    {t.audit.colTool}
                  </th>
                  <th className="py-2 px-3 text-[#a7a49c] text-[10px] uppercase font-bold border-r border-[#2e2e2e] w-28">
                    {t.audit.colClient}
                  </th>
                  <th className="py-2 px-3 text-[#f2efe6] text-[11px] font-bold border-r border-[#2e2e2e]">
                    {t.audit.colQuery}
                  </th>
                  <th className="py-2 px-3 text-[#a7a49c] text-[10px] uppercase font-bold w-20 text-right">
                    {t.audit.colLatency}
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#2e2e2e]/50">
                {filteredLogs.map((entry) => {
                  const isSelected = selectedEntry?.id === entry.id;
                  return (
                    <tr
                      key={entry.id}
                      onClick={() => setSelectedEntry(entry)}
                      className={`cursor-pointer transition-colors ${
                        isSelected
                          ? 'bg-[#171717] text-white border-l-2 border-[#d9ff3f]'
                          : 'hover:bg-[#121212] text-[#f2efe6]'
                      }`}
                    >
                      <td className="py-2 px-3 border-r border-[#2e2e2e] whitespace-nowrap">
                        {entry.allowed ? (
                          <span className="px-1.5 py-0.5 bg-[#d9ff3f]/10 border border-[#d9ff3f] text-[#d9ff3f] text-[10px] font-bold">
                            {t.audit.statusAllowed}
                          </span>
                        ) : (
                          <span className="px-1.5 py-0.5 bg-[#e11d48]/10 border border-[#e11d48] text-[#fb7185] text-[10px] font-bold">
                            {t.audit.statusBlocked}
                          </span>
                        )}
                      </td>
                      <td className="py-2 px-3 border-r border-[#2e2e2e] whitespace-nowrap text-[#a855f7] font-bold">
                        {entry.toolName}
                      </td>
                      <td className="py-2 px-3 border-r border-[#2e2e2e] text-[#a7a49c] text-[10px] whitespace-nowrap truncate max-w-[120px]">
                        {entry.clientInfo || 'client'}
                      </td>
                      <td className="py-2 px-3 border-r border-[#2e2e2e] truncate max-w-md">
                        {entry.query || <span className="text-[#a7a49c] italic">(empty)</span>}
                      </td>
                      <td className="py-2 px-3 text-right text-[10px] text-[#a7a49c] whitespace-nowrap">
                        {entry.executionMs}ms
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>

        {/* Selected Entry Detail Drawer */}
        {selectedEntry && (
          <div className="w-96 border border-[#2e2e2e] bg-[#0c0c0c] flex flex-col font-mono text-xs">
            <div className="flex items-center justify-between p-3 border-b border-[#2e2e2e] bg-[#101010]">
              <span className="font-bold text-[#f2efe6] tracking-wider">
                // {t.audit.drawerTitle}
              </span>
              <button
                onClick={() => setSelectedEntry(null)}
                className="text-[#a7a49c] hover:text-[#f2efe6] cursor-pointer"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="p-4 space-y-4 overflow-y-auto flex-1">
              <div>
                <span className="eyebrow text-[10px]">{t.audit.colStatus}:</span>
                <div className="mt-1">
                  {selectedEntry.allowed ? (
                    <div className="p-2 border border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f] font-bold">
                      ✓ {t.audit.statusAllowed}
                    </div>
                  ) : (
                    <div className="p-2 border border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185] font-bold">
                      ✕ {t.audit.statusBlocked}
                    </div>
                  )}
                </div>
              </div>

              {selectedEntry.reason && (
                <div>
                  <span className="eyebrow text-[10px]">{t.audit.drawerReason}</span>
                  <div className="p-2 bg-[#121212] border border-[#2e2e2e] text-[#fb7185] text-[11px] mt-1">
                    {selectedEntry.reason}
                  </div>
                </div>
              )}

              <div>
                <span className="eyebrow text-[10px]">{t.audit.drawerTool}</span>
                <div className="text-[#a855f7] font-bold mt-1 text-sm">
                  {selectedEntry.toolName}
                </div>
              </div>

              <div>
                <span className="eyebrow text-[10px]">{t.audit.drawerSqlPayload}</span>
                <pre className="p-3 bg-[#000000] border border-[#2e2e2e] text-[#f2efe6] text-[11px] mt-1 whitespace-pre-wrap overflow-x-auto leading-relaxed scanlines max-h-56">
                  {selectedEntry.query}
                </pre>
              </div>

              {selectedEntry.error && (
                <div>
                  <span className="eyebrow text-[10px] text-[#fb7185]">{t.common.error}:</span>
                  <div className="p-2 bg-[#e11d48]/15 border border-[#e11d48] text-[#fb7185] text-[11px] mt-1 whitespace-pre-wrap">
                    {selectedEntry.error}
                  </div>
                </div>
              )}

              <div className="grid grid-cols-2 gap-2 text-[10px] bg-[#050505] p-2.5 border border-[#2e2e2e]">
                <div>
                  <span className="text-[#a7a49c]">{t.audit.colLatency}:</span>
                  <div className="text-[#f2efe6] font-bold">{selectedEntry.executionMs}ms</div>
                </div>
                <div>
                  <span className="text-[#a7a49c]">{t.audit.drawerClient}</span>
                  <div className="text-[#f2efe6] font-bold truncate">{selectedEntry.clientInfo}</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
