export type Language = 'en' | 'es';

export interface Translations {
  common: {
    oracleAiBridge: string;
    cancel: string;
    save: string;
    delete: string;
    edit: string;
    test: string;
    active: string;
    error: string;
    success: string;
    loading: string;
    clear: string;
    copy: string;
    copied: string;
    close: string;
    all: string;
  };
  header: {
    navConnections: string;
    navQueries: string;
    navSecurity: string;
    navAudit: string;
    dbTarget: string;
    noDbTarget: string;
    mcpOnline: string;
    mcpOffline: string;
    mcpConfigBtn: string;
    secReadOnly: string;
    secPermissive: string;
  };
  connections: {
    subsystem: string;
    configuredCount: (count: number) => string;
    title: string;
    description: string;
    newConnectionBtn: string;
    noConnectionsTitle: string;
    noConnectionsDesc: string;
    setupFirstBtn: string;
    activeInMcp: string;
    activateBtn: string;
    activateTooltip: string;
    hostPort: string;
    user: string;
    encryption: string;
    testBtn: string;
    testing: string;
    testOk: string;
    testError: string;
    editTooltip: string;
    deleteTooltip: string;
    confirmDelete: string;
  };
  modal: {
    newTitle: string;
    editTitle: string;
    nameLabel: string;
    namePlaceholder: string;
    hostLabel: string;
    portLabel: string;
    idTypeLabel: string;
    serviceNameOption: string;
    sidOption: string;
    serviceNameLabel: string;
    sidLabel: string;
    usernameLabel: string;
    passwordLabel: string;
    passwordVaultNotice: string;
    sslLabel: string;
    walletLabel: string;
    walletPlaceholder: string;
    walletNotice: string;
    requiredFieldsError: string;
    hostUserRequiredError: string;
    testBtn: string;
    testingBtn: string;
    saveBtn: string;
    savingBtn: string;
    tnsSectionTitle: string;
    tnsImportBtn: string;
    tnsDetectedLabel: string;
    tnsSelectAliasPlaceholder: string;
    tnsAliasesLoaded: (count: number) => string;
    tnsNotice: string;
    dbeaverSectionTitle: string;
    dbeaverImportBtn: string;
    dbeaverDetectedLabel: string;
    dbeaverSelectConnPlaceholder: string;
    dbeaverConnsLoaded: (count: number) => string;
    dbeaverNotice: string;
    importToggleLabel: string;
    importTabTns: string;
    importTabDbeaver: string;
  };
  queryStudio: {
    subsystem: string;
    title: string;
    description: string;
    schemaTitle: string;
    searchSchemaPlaceholder: string;
    loadingSchema: string;
    emptySchema: string;
    tablesHeader: string;
    viewsHeader: string;
    proceduresHeader: string;
    tabSql: string;
    tabPlsql: string;
    runSqlBtn: string;
    runPlsqlBtn: string;
    runningBtn: string;
    exportCsvBtn: string;
    noActiveDb: string;
    executionTelemetry: (rows: number, ms: number) => string;
    plsqlSuccess: (ms: number) => string;
    plsqlConsoleTitle: string;
    clearConsole: string;
    noDbmsOutput: string;
    tableColumnsTitle: (table: string) => string;
    colName: string;
    colType: string;
    colNullable: string;
    colPk: string;
    emptyResults: string;
  };
  security: {
    subsystem: string;
    title: string;
    description: string;
    policyModeTitle: string;
    policyModeDesc: string;
    modeReadOnly: string;
    modeReadOnlyDesc: string;
    modePermissive: string;
    modePermissiveDesc: string;
    plsqlSwitchTitle: string;
    plsqlSwitchDesc: string;
    plsqlEnabled: string;
    plsqlDisabled: string;
    maxRowsTitle: string;
    maxRowsDesc: string;
    blockedKeywordsTitle: string;
    blockedKeywordsDesc: string;
    addKeywordPlaceholder: string;
    addBtn: string;
    resetDefaultsBtn: string;
    savePolicyBtn: string;
    savingPolicyBtn: string;
    savedSuccess: string;
    validatorTitle: string;
    validatorDesc: string;
    testQueryPlaceholder: string;
    testQueryBtn: string;
    queryPermitted: string;
    queryBlocked: string;
  };
  audit: {
    subsystem: string;
    title: string;
    description: string;
    allowedCount: (count: number) => string;
    blockedCount: (count: number) => string;
    clearLogsBtn: string;
    confirmClear: string;
    filterLabel: string;
    filterAll: (count: number) => string;
    filterAllowed: (count: number) => string;
    filterBlocked: (count: number) => string;
    searchPlaceholder: string;
    colTime: string;
    colStatus: string;
    colTool: string;
    colClient: string;
    colLatency: string;
    colQuery: string;
    emptyLogs: string;
    statusAllowed: string;
    statusBlocked: string;
    drawerTitle: string;
    drawerTimestamp: string;
    drawerTool: string;
    drawerClient: string;
    drawerExecutionTime: string;
    drawerReason: string;
    drawerSqlPayload: string;
    drawerMcpParams: string;
  };
  mcpModal: {
    title: string;
    telemetryTitle: string;
    statusActive: string;
    statusOffline: string;
    httpSsePortLabel: string;
    restartBtn: string;
    restartingBtn: string;
    quickTab: string;
    vscodeTab: string;
    intellijTab: string;
    claudeTab: string;
    tab1ClickTitle: string;
    tab1ClickDesc: string;
    installClaudeBtn: string;
    installVSCodeBtn: string;
    installCursorBtn: string;
    manualEndpointsTitle: string;
    sseEndpointLabel: string;
    httpEndpointLabel: string;
    cliStdioTitle: string;
    vscodeGuideTitle: string;
    vscodeGuideStep1: string;
    vscodeGuideStep2: string;
    intellijGuideTitle: string;
    intellijGuideDesc: string;
    intellijStep1: string;
    intellijStep2: string;
    intellijStep3: string;
    claudeGuideTitle: string;
    claudeGuideDesc: string;
  };
  languageModal: {
    title: string;
    description: string;
    englishTitle: string;
    englishDesc: string;
    spanishTitle: string;
    spanishDesc: string;
    activeBadge: string;
    autoDetectNotice: string;
    closeBtn: string;
  };
}

export const translations: Record<Language, Translations> = {
  en: {
    common: {
      oracleAiBridge: 'ORACLE AI BRIDGE',
      cancel: 'CANCEL',
      save: 'SAVE',
      delete: 'DELETE',
      edit: 'EDIT',
      test: 'TEST',
      active: 'ACTIVE',
      error: 'ERROR',
      success: 'SUCCESS',
      loading: 'LOADING...',
      clear: 'CLEAR',
      copy: 'COPY',
      copied: 'COPIED',
      close: 'CLOSE',
      all: 'ALL',
    },
    header: {
      navConnections: 'CONNECTIONS',
      navQueries: 'QUERIES & PL/SQL',
      navSecurity: 'MCP PERMISSIONS',
      navAudit: 'AUDIT',
      dbTarget: 'DB',
      noDbTarget: 'NO DB ACTIVE',
      mcpOnline: 'MCP: ONLINE',
      mcpOffline: 'MCP: OFFLINE',
      mcpConfigBtn: 'CONFIG MCP',
      secReadOnly: 'READ-ONLY (SELECT)',
      secPermissive: 'PERMISSIVE',
    },
    connections: {
      subsystem: '// SUB-SYSTEM: DATABASE_HUB',
      configuredCount: (count) => `[${count} CONNECTIONS CONFIGURED]`,
      title: 'ORACLE CONNECTION MANAGEMENT',
      description: 'Configure database endpoints used by the MCP bridge and the SQL studio.',
      newConnectionBtn: 'NEW CONNECTION',
      noConnectionsTitle: 'NO REGISTERED CONNECTIONS',
      noConnectionsDesc: 'Add a connection to your Oracle database (11g, 12c, 19c, 21c or 23ai) to enable MCP AI agents to inspect schemas and run queries.',
      setupFirstBtn: 'CONFIGURE FIRST CONNECTION',
      activeInMcp: 'ACTIVE IN MCP',
      activateBtn: '[ ACTIVATE ]',
      activateTooltip: 'Set as active connection for MCP agents',
      hostPort: 'HOST:PORT:',
      user: 'USER:',
      encryption: 'ENCRYPTION:',
      testBtn: 'TEST',
      testing: 'TESTING...',
      testOk: '● TEST OK',
      testError: '✕ ERROR',
      editTooltip: 'Edit connection profile',
      deleteTooltip: 'Delete connection profile',
      confirmDelete: 'Are you sure you want to delete this Oracle connection profile?',
    },
    modal: {
      newTitle: '// NEW_ORACLE_CONNECTION',
      editTitle: '// EDIT_ORACLE_PROFILE',
      nameLabel: 'CONNECTION NAME *',
      namePlaceholder: 'e.g. Production Oracle 19c',
      hostLabel: 'HOST / IP *',
      portLabel: 'PORT *',
      idTypeLabel: 'IDENTIFIER TYPE',
      serviceNameOption: 'Service Name',
      sidOption: 'SID (Legacy)',
      serviceNameLabel: 'SERVICE NAME *',
      sidLabel: 'SID *',
      usernameLabel: 'USERNAME *',
      passwordLabel: 'PASSWORD (VAULTED SECURELY IN OS)',
      passwordVaultNotice: 'Secrets are vaulted natively in macOS Keychain or Windows Credential Manager. Never saved in plaintext.',
      sslLabel: 'ENCRYPTION SSL / TCPS',
      walletLabel: 'ORACLE WALLET DIRECTORY (OPTIONAL FOR CLOUD / TCPS)',
      walletPlaceholder: '/path/to/wallet/dir or C:\\wallet',
      walletNotice: 'Specify directory containing cwallet.sso if connecting to Oracle Autonomous Cloud DB.',
      requiredFieldsError: 'Please complete all required fields (*).',
      hostUserRequiredError: 'Host and Username are required to test connection.',
      testBtn: 'TEST CONNECTION',
      testingBtn: 'TESTING...',
      saveBtn: 'SAVE PROFILE',
      savingBtn: 'SAVING...',
      tnsSectionTitle: 'IMPORT FROM TNSNAMES.ORA',
      tnsImportBtn: 'BROWSE TNSNAMES.ORA...',
      tnsDetectedLabel: 'DETECTED IN SYSTEM:',
      tnsSelectAliasPlaceholder: 'SELECT TNS ALIAS TO POPULATE FORM...',
      tnsAliasesLoaded: (count) => `[${count} ALIASES LOADED]`,
      tnsNotice: 'Parses Host, Port, Service Name / SID, and TCPS protocol directly into the form.',
      dbeaverSectionTitle: 'IMPORT FROM DBEAVER',
      dbeaverImportBtn: 'BROWSE DATA-SOURCES.JSON...',
      dbeaverDetectedLabel: 'DETECTED DBEAVER WORKSPACE:',
      dbeaverSelectConnPlaceholder: 'SELECT ORACLE CONNECTION FROM DBEAVER...',
      dbeaverConnsLoaded: (count) => `[${count} ORACLE CONNECTIONS FOUND]`,
      dbeaverNotice: 'Imports Host, Port, Service/SID, and Username directly from DBeaver data sources.',
      importToggleLabel: 'IMPORT PROFILE FROM:',
      importTabTns: 'TNSNAMES.ORA',
      importTabDbeaver: 'DBEAVER',
    },
    queryStudio: {
      subsystem: '// SUB-SYSTEM: SQL_STUDIO',
      title: 'SQL & PL/SQL EXECUTION ENGINE',
      description: 'Inspect schemas, run queries, and execute PL/SQL blocks capturing real-time DBMS_OUTPUT.',
      schemaTitle: 'OBJECT EXPLORER',
      searchSchemaPlaceholder: 'Filter tables, views...',
      loadingSchema: 'Loading schema...',
      emptySchema: 'No schema objects found or no active connection.',
      tablesHeader: 'TABLES',
      viewsHeader: 'VIEWS',
      proceduresHeader: 'PROCEDURES',
      tabSql: 'SQL QUERY (DATA GRID)',
      tabPlsql: 'PL/SQL RUNNER (DBMS_OUTPUT)',
      runSqlBtn: 'EXECUTE SQL',
      runPlsqlBtn: 'EXECUTE PL/SQL',
      runningBtn: 'RUNNING...',
      exportCsvBtn: 'EXPORT CSV',
      noActiveDb: 'No active Oracle connection selected.',
      executionTelemetry: (rows, ms) => `${rows} rows retrieved in ${ms}ms`,
      plsqlSuccess: (ms) => `Block executed successfully in ${ms}ms`,
      plsqlConsoleTitle: 'DBMS_OUTPUT CONSOLE',
      clearConsole: 'CLEAR CONSOLE',
      noDbmsOutput: 'No DBMS_OUTPUT captured for this execution.',
      tableColumnsTitle: (table) => `SCHEMA // ${table}`,
      colName: 'COLUMN',
      colType: 'TYPE',
      colNullable: 'NULLABLE',
      colPk: 'PK',
      emptyResults: 'Execute a SQL query above to see data grid results.',
    },
    security: {
      subsystem: '// SUB-SYSTEM: MCP_FIREWALL',
      title: 'MCP SECURITY POLICIES & PERMISSIONS',
      description: 'Guard rails enforcing strict database protections against unintended or destructive AI modifications.',
      policyModeTitle: 'FIREWALL OPERATION MODE',
      policyModeDesc: 'Controls whether AI agents can only perform queries or modify data.',
      modeReadOnly: 'STRICT READ-ONLY (SELECT ONLY)',
      modeReadOnlyDesc: 'Permits SELECT only. Automatically rejects INSERT, UPDATE, DELETE, DROP, ALTER, and any DDL.',
      modePermissive: 'PERMISSIVE (MODIFICATIONS ALLOWED)',
      modePermissiveDesc: 'Allows write operations, except specifically blocked keywords configured below.',
      plsqlSwitchTitle: 'ALLOW PL/SQL BLOCK EXECUTION',
      plsqlSwitchDesc: 'Enables AI agents to call oracle_execute_plsql for stored packages and anonymous blocks.',
      plsqlEnabled: 'PL/SQL ENABLED',
      plsqlDisabled: 'PL/SQL BLOCKED',
      maxRowsTitle: 'MAX ROWS RETRIEVABLE PER MCP QUERY',
      maxRowsDesc: 'Protects database performance and model context limit by capping result sets.',
      blockedKeywordsTitle: 'BLOCKED SQL KEYWORDS',
      blockedKeywordsDesc: 'Queries containing any of these keywords will be blocked by the MCP parser.',
      addKeywordPlaceholder: 'e.g. TRUNCATE',
      addBtn: 'ADD',
      resetDefaultsBtn: 'RESTORE DEFAULTS',
      savePolicyBtn: 'SAVE SECURITY POLICY',
      savingPolicyBtn: 'SAVING...',
      savedSuccess: 'POLICY APPLIED SUCCESSFULLY',
      validatorTitle: 'LIVE PARSER & VALIDATOR TEST',
      validatorDesc: 'Test how your security policy evaluates incoming SQL queries.',
      testQueryPlaceholder: 'Enter a SQL query to test...',
      testQueryBtn: 'EVALUATE QUERY',
      queryPermitted: 'ALLOWED BY POLICY',
      queryBlocked: 'BLOCKED BY POLICY',
    },
    audit: {
      subsystem: '// TELEMETRY: AUDIT_TRAIL',
      title: 'LIVE MCP AUDIT LOG (STREAM)',
      description: 'Inspect every tool call and database query initiated by connected AI agents.',
      allowedCount: (count) => `ALLOWED: ${count}`,
      blockedCount: (count) => `BLOCKED: ${count}`,
      clearLogsBtn: 'CLEAR AUDIT LOGS',
      confirmClear: 'Clear all audit logs?',
      filterLabel: 'FILTER:',
      filterAll: (count) => `ALL (${count})`,
      filterAllowed: (count) => `ALLOWED (${count})`,
      filterBlocked: (count) => `BLOCKED (${count})`,
      searchPlaceholder: 'Search queries, tools, or clients...',
      colTime: 'TIMESTAMP',
      colStatus: 'STATUS',
      colTool: 'TOOL',
      colClient: 'CLIENT',
      colLatency: 'LATENCY',
      colQuery: 'QUERY / INTENT',
      emptyLogs: 'No audit logs recorded yet. MCP calls will appear here in real time.',
      statusAllowed: 'ALLOWED',
      statusBlocked: 'BLOCKED',
      drawerTitle: 'AUDIT EVENT DETAIL',
      drawerTimestamp: 'TIMESTAMP:',
      drawerTool: 'MCP TOOL:',
      drawerClient: 'CLIENT / AGENT:',
      drawerExecutionTime: 'EXECUTION TIME:',
      drawerReason: 'SECURITY REASON / BLOCK INFO:',
      drawerSqlPayload: 'SQL / PAYLOAD:',
      drawerMcpParams: 'FULL MCP PARAMETERS:',
    },
    mcpModal: {
      title: '// MCP_SERVER_CONFIGURATION',
      telemetryTitle: 'STATUS // TELEMETRY',
      statusActive: 'LISTENING & HEALTHY',
      statusOffline: 'STOPPED',
      httpSsePortLabel: 'HTTP / SSE LISTENER PORT',
      restartBtn: 'APPLY & RESTART',
      restartingBtn: 'RESTARTING...',
      quickTab: '1-CLICK INSTALL',
      vscodeTab: 'VS CODE / COPILOT',
      intellijTab: 'INTELLIJ IDEA',
      claudeTab: 'CLAUDE DESKTOP',
      tab1ClickTitle: 'ASSISTED 1-CLICK INTEGRATIONS',
      tab1ClickDesc: 'Click a button below to automatically register dbridge in your AI coding assistant configuration.',
      installClaudeBtn: 'CONFIGURE CLAUDE DESKTOP',
      installVSCodeBtn: 'CONFIGURE VS CODE (.vscode/mcp.json)',
      installCursorBtn: 'CONFIGURE CURSOR (~/.cursor/mcp.json)',
      manualEndpointsTitle: 'ACTIVE NETWORK ENDPOINTS',
      sseEndpointLabel: 'SSE PROTOCOL URL:',
      httpEndpointLabel: 'DIRECT HTTP/JSON-RPC URL:',
      cliStdioTitle: 'CLI STDIO RUNNER COMMAND',
      vscodeGuideTitle: 'VS CODE & GITHUB COPILOT MCP SETUP',
      vscodeGuideStep1: '1. In VS Code, MCP servers are configured in your workspace .vscode/mcp.json or user settings.',
      vscodeGuideStep2: '2. Click the 1-Click Configure button or add the JSON snippet below to your settings:',
      intellijGuideTitle: 'INTELLIJ IDEA / JETBRAINS SETUP (WINDOWS & MACOS)',
      intellijGuideDesc: 'Configure dbridge in IntelliJ IDEA via the Model Context Protocol settings:',
      intellijStep1: '1. Open IntelliJ Settings (Ctrl+Alt+S on Windows / Cmd+, on macOS) -> Tools -> Model Context Protocol (MCP).',
      intellijStep2: '2. Click Add (+) -> Select "Server-Sent Events (SSE)".',
      intellijStep3: '3. Set Name to "dbridge-oracle" and URL to your active SSE endpoint:',
      claudeGuideTitle: 'CLAUDE DESKTOP SETUP',
      claudeGuideDesc: 'Claude Desktop loads servers from ~/Library/Application Support/Claude/claude_desktop_config.json on macOS, or %APPDATA%\\Claude\\claude_desktop_config.json on Windows.',
    },
    languageModal: {
      title: '// INTERFACE_LANGUAGE_SETTINGS',
      description: 'Select your preferred language for the application workspace.',
      englishTitle: 'ENGLISH',
      englishDesc: 'Default workspace language / Telemetry & Controls',
      spanishTitle: 'ESPAÑOL',
      spanishDesc: 'Localización completa en español / Telemetría y Controles',
      activeBadge: 'ACTIVE',
      autoDetectNotice: 'System default locale detected on launch. Choice is persisted locally.',
      closeBtn: 'CLOSE',
    },
  },
  es: {
    common: {
      oracleAiBridge: 'PUENTE ORACLE PARA IA',
      cancel: 'CANCELAR',
      save: 'GUARDAR',
      delete: 'ELIMINAR',
      edit: 'EDITAR',
      test: 'PROBAR',
      active: 'ACTIVA',
      error: 'ERROR',
      success: 'ÉXITO',
      loading: 'CARGANDO...',
      clear: 'LIMPIAR',
      copy: 'COPIAR',
      copied: 'COPIADO',
      close: 'CERRAR',
      all: 'TODOS',
    },
    header: {
      navConnections: 'CONEXIONES',
      navQueries: 'CONSULTAS & PL/SQL',
      navSecurity: 'PERMISOS MCP',
      navAudit: 'AUDITORÍA',
      dbTarget: 'BD',
      noDbTarget: 'SIN BD ACTIVA',
      mcpOnline: 'MCP: EN LÍNEA',
      mcpOffline: 'MCP: DESCONECTADO',
      mcpConfigBtn: 'CONFIG MCP',
      secReadOnly: 'SOLO LECTURA (SELECT)',
      secPermissive: 'PERMISIVO',
    },
    connections: {
      subsystem: '// SUB-SISTEMA: DATABASE_HUB',
      configuredCount: (count) => `[${count} CONEXIONES CONFIGURADAS]`,
      title: 'GESTIÓN DE CONEXIONES ORACLE',
      description: 'Configura los endpoints de bases de datos que el puente MCP y el explorador SQL utilizarán.',
      newConnectionBtn: 'NUEVA CONEXIÓN',
      noConnectionsTitle: 'NO HAY CONEXIONES REGISTRADAS',
      noConnectionsDesc: 'Agrega una conexión a tu base de datos Oracle (11g, 12c, 19c, 21c o 23ai) para permitir que los agentes MCP consulten esquemas.',
      setupFirstBtn: 'CONFIGURAR PRIMERA CONEXIÓN',
      activeInMcp: 'ACTIVA EN MCP',
      activateBtn: '[ ACTIVAR ]',
      activateTooltip: 'Establecer como conexión activa para MCP',
      hostPort: 'HOST:PUERTO:',
      user: 'USUARIO:',
      encryption: 'ENCRIPTACIÓN:',
      testBtn: 'PROBAR',
      testing: 'PROBANDO...',
      testOk: '● TEST OK',
      testError: '✕ ERROR',
      editTooltip: 'Editar perfil de conexión',
      deleteTooltip: 'Eliminar perfil de conexión',
      confirmDelete: '¿Confirmas que deseas eliminar este perfil de conexión Oracle?',
    },
    modal: {
      newTitle: '// NUEVA_CONEXIÓN_ORACLE',
      editTitle: '// EDITAR_PERFIL_ORACLE',
      nameLabel: 'NOMBRE DE LA CONEXIÓN *',
      namePlaceholder: 'ej. Oracle 19c Producción',
      hostLabel: 'HOST / IP *',
      portLabel: 'PUERTO *',
      idTypeLabel: 'TIPO DE IDENTIFICADOR',
      serviceNameOption: 'Nombre de Servicio',
      sidOption: 'SID (Legacy)',
      serviceNameLabel: 'NOMBRE DE SERVICIO *',
      sidLabel: 'SID *',
      usernameLabel: 'USUARIO *',
      passwordLabel: 'CONTRASEÑA (BÓVEDA SEGURA EN EL SO)',
      passwordVaultNotice: 'Los secretos se guardan de forma nativa en macOS Keychain o Windows Credential Manager. Cero texto plano.',
      sslLabel: 'ENCRIPTACIÓN SSL / TCPS',
      walletLabel: 'DIRECTORIO ORACLE WALLET (OPCIONAL PARA CLOUD / TCPS)',
      walletPlaceholder: '/ruta/al/directorio/wallet o C:\\wallet',
      walletNotice: 'Especifica la carpeta con cwallet.sso si te conectas a Oracle Autonomous Cloud DB.',
      requiredFieldsError: 'Por favor completa todos los campos requeridos (*).',
      hostUserRequiredError: 'Host y Usuario son requeridos para probar la conexión.',
      testBtn: 'PROBAR CONEXIÓN',
      testingBtn: 'PROBANDO...',
      saveBtn: 'GUARDAR PERFIL',
      savingBtn: 'GUARDANDO...',
      tnsSectionTitle: 'IMPORTAR DESDE TNSNAMES.ORA',
      tnsImportBtn: 'BUSCAR TNSNAMES.ORA...',
      tnsDetectedLabel: 'DETECTADOS EN EL SISTEMA:',
      tnsSelectAliasPlaceholder: 'SELECCIONAR ALIAS TNS PARA LLENAR EL FORMULARIO...',
      tnsAliasesLoaded: (count) => `[${count} ALIAS CARGADOS]`,
      tnsNotice: 'Autocompleta Host, Puerto, Service Name / SID y protocolo TCPS en el formulario.',
      dbeaverSectionTitle: 'IMPORTAR DESDE DBEAVER',
      dbeaverImportBtn: 'BUSCAR DATA-SOURCES.JSON...',
      dbeaverDetectedLabel: 'WORKSPACE DE DBEAVER DETECTADO:',
      dbeaverSelectConnPlaceholder: 'SELECCIONAR CONEXIÓN ORACLE DE DBEAVER...',
      dbeaverConnsLoaded: (count) => `[${count} CONEXIONES ORACLE ENCONTRADAS]`,
      dbeaverNotice: 'Importa Host, Puerto, Service/SID y Usuario directamente desde los orígenes de datos de DBeaver.',
      importToggleLabel: 'IMPORTAR PERFIL DESDE:',
      importTabTns: 'TNSNAMES.ORA',
      importTabDbeaver: 'DBEAVER',
    },
    queryStudio: {
      subsystem: '// SUB-SISTEMA: SQL_STUDIO',
      title: 'MOTOR DE CONSULTAS SQL & PL/SQL',
      description: 'Explora esquemas, ejecuta consultas y corre bloques PL/SQL capturando DBMS_OUTPUT en tiempo real.',
      schemaTitle: 'EXPLORADOR DE OBJETOS',
      searchSchemaPlaceholder: 'Filtrar tablas, vistas...',
      loadingSchema: 'Cargando esquema...',
      emptySchema: 'No se encontraron objetos o no hay conexión activa.',
      tablesHeader: 'TABLAS',
      viewsHeader: 'VISTAS',
      proceduresHeader: 'PROCEDIMIENTOS',
      tabSql: 'CONSULTA SQL (DATA GRID)',
      tabPlsql: 'EJECUTOR PL/SQL (DBMS_OUTPUT)',
      runSqlBtn: 'EJECUTAR SQL',
      runPlsqlBtn: 'EJECUTAR PL/SQL',
      runningBtn: 'EJECUTANDO...',
      exportCsvBtn: 'EXPORTAR CSV',
      noActiveDb: 'No hay una conexión Oracle activa seleccionada.',
      executionTelemetry: (rows, ms) => `${rows} filas obtenidas en ${ms}ms`,
      plsqlSuccess: (ms) => `Bloque ejecutado con éxito en ${ms}ms`,
      plsqlConsoleTitle: 'CONSOLA DBMS_OUTPUT',
      clearConsole: 'LIMPIAR CONSOLA',
      noDbmsOutput: 'No se capturó salida DBMS_OUTPUT en esta ejecución.',
      tableColumnsTitle: (table) => `ESQUEMA // ${table}`,
      colName: 'COLUMNA',
      colType: 'TIPO',
      colNullable: 'NULABLE',
      colPk: 'PK',
      emptyResults: 'Ejecuta una consulta SQL para visualizar resultados en la grilla.',
    },
    security: {
      subsystem: '// SUB-SISTEMA: MCP_FIREWALL',
      title: 'POLÍTICAS Y PERMISOS DE SEGURIDAD MCP',
      description: 'Barreras de contención que garantizan que los agentes IA no ejecuten acciones destructivas en tu base Oracle.',
      policyModeTitle: 'MODO DE OPERACIÓN DEL FIREWALL',
      policyModeDesc: 'Define si los agentes IA pueden modificar datos o únicamente realizar consultas.',
      modeReadOnly: 'SOLO LECTURA ESTRICTO (SELECT ONLY)',
      modeReadOnlyDesc: 'Solo permite sentencias SELECT. Bloquea de forma automática INSERT, UPDATE, DELETE, DROP, ALTER y DDL.',
      modePermissive: 'PERMISIVO (MODIFICACIONES HABILITADAS)',
      modePermissiveDesc: 'Permite escrituras, excepto aquellas que coincidan con las palabras clave bloqueadas.',
      plsqlSwitchTitle: 'PERMITIR EJECUCIÓN DE BLOQUES PL/SQL',
      plsqlSwitchDesc: 'Habilita a los agentes IA a llamar oracle_execute_plsql para invocar procedimientos o paquetes.',
      plsqlEnabled: 'PL/SQL HABILITADO',
      plsqlDisabled: 'PL/SQL BLOQUEADO',
      maxRowsTitle: 'LÍMITE MÁXIMO DE FILAS POR CONSULTA MCP',
      maxRowsDesc: 'Protege el rendimiento de la base y el contexto del modelo limitando el tamaño del result set.',
      blockedKeywordsTitle: 'PALABRAS CLAVE SQL BLOQUEADAS',
      blockedKeywordsDesc: 'Cualquier consulta que contenga estas palabras será rechazada por el parser de seguridad.',
      addKeywordPlaceholder: 'ej. TRUNCATE',
      addBtn: 'AGREGAR',
      resetDefaultsBtn: 'RESTAURAR VALORES POR DEFECTO',
      savePolicyBtn: 'GUARDAR POLÍTICA DE SEGURIDAD',
      savingPolicyBtn: 'GUARDANDO...',
      savedSuccess: 'POLÍTICA APLICADA CON ÉXITO',
      validatorTitle: 'SIMULADOR Y PRUEBA DEL PARSER EN VIVO',
      validatorDesc: 'Comprueba cómo reacciona tu política ante cualquier sentencia SQL.',
      testQueryPlaceholder: 'Escribe una consulta SQL para evaluar...',
      testQueryBtn: 'EVALUAR CONSULTA',
      queryPermitted: 'PERMITIDA POR LA POLÍTICA',
      queryBlocked: 'BLOQUEADA POR LA POLÍTICA',
    },
    audit: {
      subsystem: '// TELEMETRÍA: AUDIT_TRAIL',
      title: 'REGISTRO DE AUDITORÍA EN VIVO (MCP CALLS)',
      description: 'Inspecciona cada solicitud de herramientas e intenciones de consultas enviadas por agentes de Inteligencia Artificial.',
      allowedCount: (count) => `PERMITIDAS: ${count}`,
      blockedCount: (count) => `BLOQUEADAS: ${count}`,
      clearLogsBtn: 'LIMPIAR REGISTROS',
      confirmClear: '¿Confirmas que deseas limpiar todos los registros de auditoría?',
      filterLabel: 'FILTRO:',
      filterAll: (count) => `TODOS (${count})`,
      filterAllowed: (count) => `PERMITIDAS (${count})`,
      filterBlocked: (count) => `BLOQUEADAS (${count})`,
      searchPlaceholder: 'Buscar por consulta, herramienta o cliente...',
      colTime: 'TIMESTAMP',
      colStatus: 'ESTADO',
      colTool: 'HERRAMIENTA',
      colClient: 'CLIENTE',
      colLatency: 'LATENCIA',
      colQuery: 'CONSULTA / INTENCIÓN',
      emptyLogs: 'No hay eventos registrados aún. Las llamadas de MCP aparecerán en tiempo real.',
      statusAllowed: 'PERMITIDA',
      statusBlocked: 'BLOQUEADA',
      drawerTitle: 'DETALLE DEL EVENTO DE AUDITORÍA',
      drawerTimestamp: 'TIMESTAMP:',
      drawerTool: 'HERRAMIENTA MCP:',
      drawerClient: 'CLIENTE / AGENTE:',
      drawerExecutionTime: 'TIEMPO DE EJECUCIÓN:',
      drawerReason: 'MOTIVO DE SEGURIDAD / BLOQUEO:',
      drawerSqlPayload: 'SENTENCIA SQL / PAYLOAD:',
      drawerMcpParams: 'PARÁMETROS COMPLETOS MCP:',
    },
    mcpModal: {
      title: '// CONFIGURACIÓN_SERVIDOR_MCP',
      telemetryTitle: 'ESTADO // TELEMETRÍA',
      statusActive: 'ESCUCHANDO & SALUDABLE',
      statusOffline: 'DETENIDO',
      httpSsePortLabel: 'PUERTO HTTP / SSE LISTENER',
      restartBtn: 'APLICAR & REINICIAR',
      restartingBtn: 'REINICIANDO...',
      quickTab: 'INSTALACIÓN 1-CLIC',
      vscodeTab: 'VS CODE / COPILOT',
      intellijTab: 'INTELLIJ IDEA',
      claudeTab: 'CLAUDE DESKTOP',
      tab1ClickTitle: 'INTEGRACIONES ASISTIDAS EN 1 CLIC',
      tab1ClickDesc: 'Presiona un botón para registrar dbridge automáticamente en la configuración de tu asistente IA.',
      installClaudeBtn: 'CONFIGURAR CLAUDE DESKTOP',
      installVSCodeBtn: 'CONFIGURAR VS CODE (.vscode/mcp.json)',
      installCursorBtn: 'CONFIGURAR CURSOR (~/.cursor/mcp.json)',
      manualEndpointsTitle: 'ENDPOINTS DE RED ACTIVOS',
      sseEndpointLabel: 'URL DE PROTOCOLO SSE:',
      httpEndpointLabel: 'URL DIRECTA HTTP/JSON-RPC:',
      cliStdioTitle: 'COMANDO CLI MODO STDIO',
      vscodeGuideTitle: 'CONFIGURACIÓN PARA VS CODE Y GITHUB COPILOT',
      vscodeGuideStep1: '1. En VS Code, los servidores MCP se configuran en .vscode/mcp.json en tu workspace.',
      vscodeGuideStep2: '2. Pulsa el botón de configuración automática en 1-clic o agrega este bloque a tu archivo:',
      intellijGuideTitle: 'CONFIGURACIÓN PARA INTELLIJ IDEA (WINDOWS & MACOS)',
      intellijGuideDesc: 'Configura dbridge en IntelliJ IDEA mediante el menú Model Context Protocol:',
      intellijStep1: '1. Abre Preferencias (Ctrl+Alt+S en Windows / Cmd+, en macOS) -> Tools -> Model Context Protocol (MCP).',
      intellijStep2: '2. Haz clic en Agregar (+) -> Selecciona tipo "Server-Sent Events (SSE)".',
      intellijStep3: '3. En Name pon "dbridge-oracle" y en URL tu endpoint SSE activo:',
      claudeGuideTitle: 'CONFIGURACIÓN PARA CLAUDE DESKTOP',
      claudeGuideDesc: 'Claude Desktop lee servidores desde ~/Library/Application Support/Claude/claude_desktop_config.json en macOS, o %APPDATA%\\Claude\\claude_desktop_config.json en Windows.',
    },
    languageModal: {
      title: '// CONFIGURACIÓN_DE_IDIOMA',
      description: 'Selecciona el idioma preferido para la interfaz de la aplicación.',
      englishTitle: 'ENGLISH',
      englishDesc: 'Idioma predeterminado / Telemetría y Controles',
      spanishTitle: 'ESPAÑOL',
      spanishDesc: 'Localización completa en español / Telemetría y Controles',
      activeBadge: 'ACTIVO',
      autoDetectNotice: 'Detección automática al iniciar según el SO. Tu elección se guarda localmente.',
      closeBtn: 'CERRAR',
    },
  },
};
