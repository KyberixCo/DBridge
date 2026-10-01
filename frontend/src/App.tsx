import React, { useState, useEffect } from 'react';
import { models } from '../wailsjs/go/models';
import {
  GetConnections,
  GetActiveConnection,
  SaveConnection,
  DeleteConnection,
  SetActiveConnection,
  GetSecurityPolicy,
  SaveSecurityPolicy,
  GetMCPServerStatus,
  GetAuditLogs,
  ClearAuditLogs,
} from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

import { Header } from './components/Header';
import { ConnectionsTab } from './components/ConnectionsTab';
import { QueryStudioTab } from './components/QueryStudioTab';
import { SecurityTab } from './components/SecurityTab';
import { AuditTab } from './components/AuditTab';
import { ConnectionModal } from './components/ConnectionModal';
import { McpConfigModal } from './components/McpConfigModal';
import { useI18n } from './i18n/LanguageContext';

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('connections');

  // Application Data States
  const [connections, setConnections] = useState<models.ConnectionProfile[]>([]);
  const [activeConnection, setActiveConnection] = useState<models.ConnectionProfile | null>(null);
  const [securityPolicy, setSecurityPolicy] = useState<models.SecurityPolicy>(
    models.SecurityPolicy.createFrom()
  );
  const [serverStatus, setServerStatus] = useState<models.MCPServerStatus | null>(null);
  const [auditLogs, setAuditLogs] = useState<models.AuditLogEntry[]>([]);

  // Modals
  const [isConnModalOpen, setIsConnModalOpen] = useState(false);
  const [editingProfile, setEditingProfile] = useState<models.ConnectionProfile | null>(null);
  const [isMcpModalOpen, setIsMcpModalOpen] = useState(false);

  // Initialize and load data
  useEffect(() => {
    loadAllData();

    // Listen to real-time incoming MCP calls via Wails event bus
    const unbind = EventsOn('mcp:audit', (entry: any) => {
      const parsedEntry = models.AuditLogEntry.createFrom(entry);
      setAuditLogs((prev) => [parsedEntry, ...prev.slice(0, 499)]);
      // Also refresh server stats
      refreshServerStatus();
    });

    return () => {
      unbind();
    };
  }, []);

  const loadAllData = async () => {
    try {
      const [conns, active, policy, status, logs] = await Promise.all([
        GetConnections(),
        GetActiveConnection(),
        GetSecurityPolicy(),
        GetMCPServerStatus(),
        GetAuditLogs(),
      ]);

      setConnections(conns || []);
      setActiveConnection(active || null);
      if (policy) setSecurityPolicy(policy);
      setServerStatus(status || null);
      setAuditLogs(logs || []);
    } catch (err) {
      console.error('Error loading initial data:', err);
    }
  };

  const refreshServerStatus = async () => {
    try {
      const status = await GetMCPServerStatus();
      setServerStatus(status);
    } catch (err) {
      console.error(err);
    }
  };

  const { t } = useI18n();

  // Connection Handlers
  const handleSaveConnection = async (profile: models.ConnectionProfile) => {
    await SaveConnection(profile);
    await loadAllData();
  };

  const handleDeleteConnection = async (id: string) => {
    if (confirm(t.connections.confirmDelete)) {
      await DeleteConnection(id);
      await loadAllData();
    }
  };

  const handleSetActiveConnection = async (id: string) => {
    await SetActiveConnection(id);
    const active = await GetActiveConnection();
    setActiveConnection(active);
    await refreshServerStatus();
  };

  const handleOpenEdit = (profile: models.ConnectionProfile) => {
    setEditingProfile(profile);
    setIsConnModalOpen(true);
  };

  const handleOpenNew = () => {
    setEditingProfile(null);
    setIsConnModalOpen(true);
  };

  // Security Policy Handlers
  const handleSavePolicy = async (newPolicy: models.SecurityPolicy) => {
    await SaveSecurityPolicy(newPolicy);
    setSecurityPolicy(newPolicy);
  };

  // Audit Handlers
  const handleClearLogs = async () => {
    await ClearAuditLogs();
    setAuditLogs([]);
  };

  return (
    <div className="h-screen w-screen flex flex-col bg-[#05090C] text-[#E0E6ED] overflow-hidden select-none font-sans">
      {/* Top Header / HUD */}
      <Header
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        activeConnection={activeConnection}
        serverStatus={serverStatus}
        securityPolicy={securityPolicy}
        onOpenMcpModal={() => setIsMcpModalOpen(true)}
        auditCount={auditLogs.length}
      />

      {/* Main Content Area */}
      <main className="flex-1 overflow-hidden relative">
        {activeTab === 'connections' && (
          <ConnectionsTab
            connections={connections}
            activeConnectionId={activeConnection?.id || ''}
            onSetActive={handleSetActiveConnection}
            onDelete={handleDeleteConnection}
            onEdit={handleOpenEdit}
            onNew={handleOpenNew}
          />
        )}

        {activeTab === 'queries' && (
          <QueryStudioTab activeConnection={activeConnection} />
        )}

        {activeTab === 'security' && (
          <SecurityTab
            policy={securityPolicy}
            onSavePolicy={handleSavePolicy}
          />
        )}

        {activeTab === 'audit' && (
          <AuditTab
            logs={auditLogs}
            onClearLogs={handleClearLogs}
          />
        )}
      </main>

      {/* Connection Create/Edit Modal */}
      <ConnectionModal
        isOpen={isConnModalOpen}
        onClose={() => setIsConnModalOpen(false)}
        onSave={handleSaveConnection}
        initialProfile={editingProfile}
      />

      {/* MCP Configuration & Client Setup Modal */}
      <McpConfigModal
        isOpen={isMcpModalOpen}
        onClose={() => setIsMcpModalOpen(false)}
        status={serverStatus}
        onRefreshStatus={refreshServerStatus}
      />
    </div>
  );
};

export default App;
