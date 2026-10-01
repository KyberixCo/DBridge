import React, { useState, useEffect } from 'react';
import { Database, Radio, ExternalLink, Languages } from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import { Environment } from '../../wailsjs/runtime/runtime';
import { WindowControls, PlatformType } from './atoms/WindowControls';
import { BracketBadge } from './atoms/BracketBadge';
import { TacticalButton } from './atoms/TacticalButton';
import { useI18n } from '../i18n/LanguageContext';
import { LanguageModal } from './LanguageModal';

interface HeaderProps {
  activeTab: string;
  setActiveTab: (tab: string) => void;
  activeConnection: models.ConnectionProfile | null;
  serverStatus: models.MCPServerStatus | null;
  securityPolicy: models.SecurityPolicy | null;
  onOpenMcpModal: () => void;
  auditCount: number;
}

export const Header: React.FC<HeaderProps> = ({
  activeTab,
  setActiveTab,
  activeConnection,
  serverStatus,
  securityPolicy,
  onOpenMcpModal,
  auditCount,
}) => {
  const [platform, setPlatform] = useState<PlatformType>('darwin');
  const [isLangModalOpen, setIsLangModalOpen] = useState(false);
  const { language, setLanguage, t } = useI18n();

  useEffect(() => {
    Environment()
      .then((env) => {
        if (env.platform === 'darwin') {
          setPlatform('darwin');
        } else if (env.platform === 'windows') {
          setPlatform('windows');
        } else {
          setPlatform('other');
        }
      })
      .catch(() => {});
  }, []);

  const navItems = [
    { id: 'connections', label: t.header.navConnections, num: '01' },
    { id: 'queries', label: t.header.navQueries, num: '02' },
    { id: 'security', label: t.header.navSecurity, num: '03' },
    { id: 'audit', label: `${t.header.navAudit}${auditCount > 0 ? ` [${auditCount}]` : ''}`, num: '04' },
  ];

  return (
    <div className="select-none shrink-0">
      {/* 1. TOP BAR // INTEGRATED WINDOW CONTROLS & TACTICAL BRANDING */}
      <header
        style={{ '--wails-draggable': 'drag' } as React.CSSProperties}
        className="h-10 border-b border-[#2e2e2e] bg-[#0c0c0c] flex items-center justify-between shrink-0 select-none cursor-default"
      >
        {/* Left: Window Controls (macOS) + Brand */}
        <div className="flex items-center h-full min-w-0">
          {platform === 'darwin' && <WindowControls platform="darwin" />}

          <div className="flex items-center gap-2.5 min-w-0 pl-3">
            <div className="w-5 h-5 bg-[#d9ff3f] text-[#050505] font-mono font-black text-xs grid place-items-center tracking-tighter shrink-0">
              K
            </div>
            <div className="flex items-center gap-2 truncate">
              <span className="font-mono font-black text-xs text-[#f2efe6] tracking-tight uppercase">
                KYBERIX
              </span>
              <span className="text-[#383838] font-mono text-xs">/</span>
              <span className="eyebrow text-[10px] truncate">
                DBRIDGE // ORACLE AI BRIDGE
              </span>
            </div>
          </div>
        </div>

        {/* Right: Telemetry Badges & Windows Controls */}
        <div className="flex items-center h-full shrink-0">
          <div
            style={{ '--wails-draggable': 'no-drag' } as React.CSSProperties}
            className="flex items-center gap-2 shrink-0 pr-3"
          >
            {/* DB Target Bracket */}
            {activeConnection ? (
              <BracketBadge
                label={`${t.header.dbTarget}: ${activeConnection.name}`}
                variant="solid-acid"
                pulse={true}
              />
            ) : (
              <BracketBadge
                label={`${t.header.dbTarget}: ${t.header.noDbTarget}`}
                variant="solid-crimson"
              />
            )}

            {/* Policy Badge */}
            <BracketBadge
              label={`POLICY: ${
                securityPolicy?.mode === 'read_only'
                  ? t.header.secReadOnly
                  : t.header.secPermissive
              }`}
              variant={securityPolicy?.mode === 'read_only' ? 'mint' : 'danger'}
            />

            {securityPolicy?.allowPlsql && (
              <BracketBadge
                label="PL/SQL: ON"
                variant="purple"
              />
            )}

            {/* Language Modal Trigger Button */}
            <button
              onClick={() => setIsLangModalOpen(true)}
              className="flex items-center gap-1.5 px-2 py-1 border border-[#2e2e2e] bg-[#050505] hover:border-[#d9ff3f] text-[#a7a49c] hover:text-[#d9ff3f] text-[10px] font-mono font-bold transition-all cursor-pointer"
              title={language === 'en' ? 'Interface language: English (Click to change)' : 'Idioma de la interfaz: Español (Click para cambiar)'}
            >
              <Languages className="w-3 h-3 text-[#d9ff3f]" />
              <span className="text-[#f2efe6] tracking-wider">{language.toUpperCase()}</span>
            </button>

            {/* MCP SSE Trigger Button */}
            <TacticalButton
              size="sm"
              variant="purple"
              onClick={onOpenMcpModal}
              icon={<Radio className="w-3 h-3 text-[#050505]" />}
              title={t.header.mcpConfigBtn}
            >
              SSE :{serverStatus?.port || 8085}
            </TacticalButton>
          </div>

          {platform !== 'darwin' && <WindowControls platform="windows" />}
        </div>
      </header>

      {/* 2. TACTICAL NAVIGATION TABS */}
      <nav className="flex items-center px-4 border-b border-[#2e2e2e] bg-[#050505]">
        {navItems.map((item) => {
          const isActive = activeTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveTab(item.id)}
              className={`flex items-center py-2 px-4 text-xs font-mono font-bold tracking-wider uppercase transition-all cursor-pointer border-b-2 ${
                isActive
                  ? 'border-[#d9ff3f] text-[#f2efe6] bg-[#141414]'
                  : 'border-transparent text-[#a7a49c] hover:text-[#f2efe6] hover:bg-[#0c0c0c]'
              }`}
            >
              <span className="number-tag mr-2">{item.num}</span>
              <span>{item.label}</span>
            </button>
          );
        })}
      </nav>

      {/* Language Selection Modal */}
      <LanguageModal
        isOpen={isLangModalOpen}
        onClose={() => setIsLangModalOpen(false)}
      />
    </div>
  );
};
