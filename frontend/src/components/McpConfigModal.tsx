import React, { useState } from 'react';
import {
  X,
  Copy,
  Check,
  Radio,
  Terminal,
  Cpu,
  RefreshCw,
  Sparkles,
  CheckCircle2,
  AlertCircle,
  HelpCircle,
  Settings2,
  FolderOpen
} from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import {
  RestartMCPServer,
  AutoConfigureClaude,
  AutoConfigureVSCode,
  AutoConfigureCursor,
} from '../../wailsjs/go/main/App';
import { TacticalButton } from './atoms/TacticalButton';
import { BracketBadge } from './atoms/BracketBadge';
import { useI18n } from '../i18n/LanguageContext';

interface McpConfigModalProps {
  isOpen: boolean;
  onClose: () => void;
  status: models.MCPServerStatus | null;
  onRefreshStatus: () => void;
}

export const McpConfigModal: React.FC<McpConfigModalProps> = ({
  isOpen,
  onClose,
  status,
  onRefreshStatus,
}) => {
  const { t } = useI18n();
  const [activeSubTab, setActiveSubTab] = useState<'quick' | 'vscode' | 'intellij' | 'claude'>('quick');
  const [copiedSection, setCopiedSection] = useState<string | null>(null);
  const [port, setPort] = useState<number>(status?.port || 8085);
  const [isRestarting, setIsRestarting] = useState(false);
  const [autoMessage, setAutoMessage] = useState<{ text: string; isError: boolean } | null>(null);
  const [inspectionMode, setInspectionMode] = useState(false);

  if (!isOpen) return null;

  const currentPort = status?.port || 8085;
  const sseUrl = `http://localhost:${currentPort}/sse`;
  const httpUrl = `http://localhost:${currentPort}/mcp`;

  const copyToClipboard = (text: string, section: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSection(section);
    setTimeout(() => setCopiedSection(null), 2000);
  };

  const handleRestart = async () => {
    setIsRestarting(true);
    setAutoMessage(null);
    try {
      await RestartMCPServer(port);
      onRefreshStatus();
      setAutoMessage({ text: `Servidor reiniciado en puerto ${port}`, isError: false });
    } catch (err: any) {
      setAutoMessage({ text: `Error reiniciando: ${err?.message || err}`, isError: true });
    } finally {
      setIsRestarting(false);
    }
  };

  // Automated 1-Click Configurations
  const handleAutoClaude = async () => {
    setAutoMessage(null);
    try {
      const res = await AutoConfigureClaude();
      setAutoMessage({ text: res, isError: false });
    } catch (err: any) {
      setAutoMessage({ text: err?.message || String(err), isError: true });
    }
  };

  const handleAutoVSCode = async (transport: 'command' | 'sse' = 'command') => {
    setAutoMessage(null);
    try {
      const res = await AutoConfigureVSCode('', transport === 'command' && inspectionMode ? 'inspect' : transport);
      setAutoMessage({ text: res, isError: false });
    } catch (err: any) {
      setAutoMessage({ text: err?.message || String(err), isError: true });
    }
  };

  const handleAutoCursor = async () => {
    setAutoMessage(null);
    try {
      const res = await AutoConfigureCursor();
      setAutoMessage({ text: res, isError: false });
    } catch (err: any) {
      setAutoMessage({ text: err?.message || String(err), isError: true });
    }
  };

  const vsCodeCommandConfig = {
    servers: {
      "dbridge-oracle": {
        type: "stdio",
        command: "dbridge",
        args: inspectionMode ? ["--mcp", "--inspect"] : ["--mcp"]
      }
    }
  };

  const vsCodeSseConfig = {
    servers: {
      "dbridge-oracle": {
        type: "sse",
        url: sseUrl
      }
    }
  };

  const claudeDesktopConfig = {
    mcpServers: {
      oracle: {
        url: sseUrl
      }
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-xs p-4">
      <div className="brutal-card w-full max-w-3xl bg-[#0c0c0c] border-2 border-[#2e2e2e] text-[#f2efe6] shadow-[10px_10px_0_#050505] flex flex-col max-h-[92vh]">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-[#2e2e2e] bg-[#050505] shrink-0">
          <div className="flex items-center gap-2">
            <Radio className="w-4 h-4 text-[#d9ff3f]" />
            <span className="eyebrow text-xs font-black">
              {t.mcpModal.title}
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
        <div className="p-6 space-y-5 overflow-y-auto font-mono text-xs flex-1">
          {/* 1. PORT CONFIGURATION & SERVER TELEMETRY */}
          <div className="p-4 bg-[#050505] border border-[#2e2e2e] space-y-3">
            <div className="flex items-center justify-between flex-wrap gap-2">
              <div className="flex items-center gap-3">
                <span className="w-2.5 h-2.5 bg-[#d9ff3f] animate-ping" />
                <div>
                  <div className="text-[#f2efe6] font-bold text-sm">
                    {t.mcpModal.telemetryTitle}: <span className="text-[#d9ff3f]">{t.mcpModal.statusActive}</span>
                  </div>
                  <div className="text-[11px] text-[#a7a49c] mt-0.5">
                    Transport: <span className="text-white">SSE</span> + <span className="text-white">HTTP JSON-RPC</span>
                  </div>
                </div>
              </div>

              {/* Port Editor */}
              <div className="flex items-center gap-2 bg-[#121212] p-1.5 border border-[#2e2e2e]">
                <Settings2 className="w-3.5 h-3.5 text-[#a7a49c]" />
                <span className="text-[11px] text-[#a7a49c] uppercase font-bold">{t.mcpModal.httpSsePortLabel}:</span>
                <input
                  type="number"
                  value={port}
                  onChange={(e) => setPort(parseInt(e.target.value) || 8085)}
                  className="w-16 bg-[#050505] border border-[#2e2e2e] px-1.5 py-0.5 text-center text-xs text-[#f2efe6] focus:border-[#d9ff3f] focus:outline-none"
                />
                <TacticalButton
                  size="sm"
                  variant="acid"
                  onClick={handleRestart}
                  disabled={isRestarting}
                  icon={<RefreshCw className={`w-3 h-3 ${isRestarting ? 'animate-spin' : ''}`} />}
                >
                  {isRestarting ? t.mcpModal.restartingBtn : t.mcpModal.restartBtn}
                </TacticalButton>
              </div>
            </div>

            {/* Quick URL Strip */}
            <div className="flex items-center justify-between bg-[#121212] p-2 border border-[#2e2e2e] text-[11px]">
              <div className="flex items-center gap-2 truncate">
                <span className="text-[#a7a49c]">{t.mcpModal.sseEndpointLabel}</span>
                <code className="text-[#d9ff3f] font-bold select-all">{sseUrl}</code>
              </div>
              <button
                onClick={() => copyToClipboard(sseUrl, 'sse-url')}
                className="text-[10px] text-[#a7a49c] hover:text-[#d9ff3f] cursor-pointer"
              >
                {copiedSection === 'sse-url' ? `[ ${t.common.copied} ]` : `[ ${t.common.copy} ]`}
              </button>
            </div>

            {/* Feedback Message */}
            {autoMessage && (
              <div
                className={`p-2 border text-xs flex items-center gap-2 ${
                  autoMessage.isError
                    ? 'border-[#e11d48] bg-[#e11d48]/10 text-[#fb7185]'
                    : 'border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f]'
                }`}
              >
                {autoMessage.isError ? <AlertCircle className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                <span className="font-bold">{autoMessage.text}</span>
              </div>
            )}
          </div>

          {/* 2. AUTOMATIC 1-CLICK ACTIONS */}
          <div className="p-4 bg-[#050505] border border-[#2e2e2e]">
            <div className="flex items-center gap-2 mb-2">
              <Sparkles className="w-4 h-4 text-[#d9ff3f]" />
              <span className="eyebrow text-xs">// {t.mcpModal.tab1ClickTitle}</span>
            </div>
            <p className="text-[11px] text-[#a7a49c] mb-3 leading-relaxed">
              {t.mcpModal.tab1ClickDesc}
            </p>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="p-3 bg-[#101010] border border-[#2e2e2e] flex flex-col justify-between">
                <div>
                  <div className="font-bold text-[#f2efe6] text-xs">VS CODE / COPILOT</div>
                  <div className="text-[10px] text-[#a7a49c] mt-1">%APPDATA%\Code\User\mcp.json</div>
                </div>
                <div className="mt-3 space-y-1.5">
                  <label className="flex items-start gap-2 text-[10px] text-[#a7a49c] cursor-pointer">
                    <input type="checkbox" checked={inspectionMode} onChange={event => setInspectionMode(event.target.checked)} />
                    <span>{t.mcpModal.inspectionMode}</span>
                  </label>
                  <TacticalButton
                    size="sm"
                    variant="acid"
                    onClick={() => handleAutoVSCode('command')}
                    className="w-full text-[10px]"
                    title="Configura VS Code usando comando CLI directo (dbridge --mcp)"
                  >
                    {t.mcpModal.installVSCodeCommandBtn}
                  </TacticalButton>
                  <button
                    type="button"
                    onClick={() => handleAutoVSCode('sse')}
                    className="w-full text-center py-1 text-[10px] font-mono font-bold text-[#a7a49c] hover:text-[#f2efe6] bg-[#1a1a1a] hover:bg-[#252525] border border-[#2e2e2e] transition-colors cursor-pointer"
                    title="Configura VS Code usando URL de red SSE"
                  >
                    {t.mcpModal.installVSCodeSseBtn}
                  </button>
                </div>
              </div>

              <div className="p-3 bg-[#101010] border border-[#2e2e2e] flex flex-col justify-between">
                <div>
                  <div className="font-bold text-[#f2efe6] text-xs">CLAUDE DESKTOP</div>
                  <div className="text-[10px] text-[#a7a49c] mt-1">claude_desktop_config.json</div>
                </div>
                <TacticalButton
                  size="sm"
                  variant="purple"
                  onClick={handleAutoClaude}
                  className="mt-3 w-full"
                >
                  {t.mcpModal.installClaudeBtn}
                </TacticalButton>
              </div>

              <div className="p-3 bg-[#101010] border border-[#2e2e2e] flex flex-col justify-between">
                <div>
                  <div className="font-bold text-[#f2efe6] text-xs">CURSOR / WINDSURF</div>
                  <div className="text-[10px] text-[#a7a49c] mt-1">~/.cursor/mcp.json</div>
                </div>
                <TacticalButton
                  size="sm"
                  variant="ghost"
                  onClick={handleAutoCursor}
                  className="mt-3 w-full"
                >
                  {t.mcpModal.installCursorBtn}
                </TacticalButton>
              </div>
            </div>
          </div>

          {/* 3. STEP-BY-STEP GUIDES TABS */}
          <div className="border border-[#2e2e2e] bg-[#050505]">
            <div className="flex border-b border-[#2e2e2e] bg-[#0c0c0c] text-xs font-bold">
              <button
                onClick={() => setActiveSubTab('vscode')}
                className={`py-2 px-4 border-r border-[#2e2e2e] transition-colors ${
                  activeSubTab === 'vscode'
                    ? 'bg-[#141414] text-[#d9ff3f] border-b-2 border-b-[#d9ff3f]'
                    : 'text-[#a7a49c] hover:text-[#f2efe6]'
                }`}
              >
                {t.mcpModal.vscodeTab}
              </button>
              <button
                onClick={() => setActiveSubTab('intellij')}
                className={`py-2 px-4 border-r border-[#2e2e2e] transition-colors ${
                  activeSubTab === 'intellij'
                    ? 'bg-[#141414] text-[#d9ff3f] border-b-2 border-b-[#d9ff3f]'
                    : 'text-[#a7a49c] hover:text-[#f2efe6]'
                }`}
              >
                {t.mcpModal.intellijTab}
              </button>
              <button
                onClick={() => setActiveSubTab('claude')}
                className={`py-2 px-4 transition-colors ${
                  activeSubTab === 'claude'
                    ? 'bg-[#141414] text-[#d9ff3f] border-b-2 border-b-[#d9ff3f]'
                    : 'text-[#a7a49c] hover:text-[#f2efe6]'
                }`}
              >
                {t.mcpModal.claudeTab}
              </button>
            </div>

            <div className="p-4 space-y-3">
              {activeSubTab === 'vscode' && (
                <div className="space-y-4">
                  <div className="text-xs text-[#f2efe6] leading-relaxed">
                    <strong>{t.mcpModal.vscodeGuideTitle}</strong>
                    <ol className="list-decimal pl-4 mt-2 space-y-1.5 text-[#a7a49c]">
                      <li>
                        {t.mcpModal.vscodeGuideStep1}
                      </li>
                      <li>
                        {t.mcpModal.vscodeGuideStep2}
                      </li>
                    </ol>
                  </div>

                  {/* 1. Command Mode (Recommended) */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] text-[#d9ff3f] font-bold uppercase tracking-wider">
                        {t.mcpModal.vscodeModeCommand}
                      </span>
                    </div>
                    <div className="relative">
                      <pre className="p-3 bg-[#000000] border border-[#2e2e2e] text-[11px] text-[#d9ff3f] overflow-x-auto leading-relaxed">
                        {JSON.stringify(vsCodeCommandConfig, null, 2)}
                      </pre>
                      <button
                        onClick={() => copyToClipboard(JSON.stringify(vsCodeCommandConfig, null, 2), 'vscode-cmd-json')}
                        className="absolute top-2 right-2 brutal-button brutal-button--ghost py-0.5 px-2 min-h-[26px] text-[10px]"
                      >
                        {copiedSection === 'vscode-cmd-json' ? `[ ${t.common.copied} ]` : `[ ${t.common.copy} JSON ]`}
                      </button>
                    </div>
                  </div>

                  {/* 2. SSE Mode */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] text-[#a7a49c] font-bold uppercase tracking-wider">
                        {t.mcpModal.vscodeModeSse}
                      </span>
                    </div>
                    <div className="relative">
                      <pre className="p-3 bg-[#000000] border border-[#2e2e2e] text-[11px] text-[#a7a49c] overflow-x-auto leading-relaxed">
                        {JSON.stringify(vsCodeSseConfig, null, 2)}
                      </pre>
                      <button
                        onClick={() => copyToClipboard(JSON.stringify(vsCodeSseConfig, null, 2), 'vscode-sse-json')}
                        className="absolute top-2 right-2 brutal-button brutal-button--ghost py-0.5 px-2 min-h-[26px] text-[10px]"
                      >
                        {copiedSection === 'vscode-sse-json' ? `[ ${t.common.copied} ]` : `[ ${t.common.copy} JSON ]`}
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {activeSubTab === 'intellij' && (
                <div className="space-y-3">
                  <div className="text-xs text-[#f2efe6] leading-relaxed">
                    <strong>{t.mcpModal.intellijGuideTitle}</strong>
                    <ol className="list-decimal pl-4 mt-2 space-y-1.5 text-[#a7a49c]">
                      <li>
                        {t.mcpModal.intellijStep1}
                      </li>
                      <li>
                        {t.mcpModal.intellijStep2}
                      </li>
                      <li>
                        {t.mcpModal.intellijStep3}
                      </li>
                    </ol>
                  </div>

                  <div className="p-3 bg-[#101010] border border-[#2e2e2e] flex items-center justify-between">
                    <div>
                      <span className="text-[10px] text-[#a7a49c] uppercase font-bold">{t.mcpModal.sseEndpointLabel}</span>
                      <div className="text-[#d9ff3f] font-bold text-xs">{sseUrl}</div>
                    </div>
                    <TacticalButton
                      size="sm"
                      variant="acid"
                      onClick={() => copyToClipboard(sseUrl, 'intellij-url')}
                    >
                      {copiedSection === 'intellij-url' ? t.common.copied : t.common.copy}
                    </TacticalButton>
                  </div>
                </div>
              )}

              {activeSubTab === 'claude' && (
                <div className="space-y-3">
                  <div className="text-xs text-[#f2efe6] leading-relaxed">
                    <strong>{t.mcpModal.claudeGuideTitle}</strong>
                    <p className="mt-1 text-[#a7a49c]">
                      {t.mcpModal.claudeGuideDesc}
                    </p>
                  </div>

                  <div className="relative">
                    <pre className="p-3 bg-[#000000] border border-[#2e2e2e] text-[11px] text-[#a855f7] overflow-x-auto leading-relaxed">
                      {JSON.stringify(claudeDesktopConfig, null, 2)}
                    </pre>
                    <button
                      onClick={() => copyToClipboard(JSON.stringify(claudeDesktopConfig, null, 2), 'claude-json')}
                      className="absolute top-2 right-2 brutal-button brutal-button--ghost py-0.5 px-2 min-h-[26px] text-[10px]"
                    >
                      {copiedSection === 'claude-json' ? `[ ${t.common.copied} ]` : `[ ${t.common.copy} JSON ]`}
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="flex items-center justify-end px-5 py-3 border-t border-[#2e2e2e] bg-[#050505] shrink-0">
          <TacticalButton variant="ghost" onClick={onClose}>
            {t.common.close}
          </TacticalButton>
        </div>
      </div>
    </div>
  );
};
