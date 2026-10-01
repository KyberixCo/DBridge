import React, { useState, useEffect, useRef, useMemo } from 'react';
import {
  Play,
  Terminal,
  Database,
  Table,
  Copy,
  Check,
  Download,
  AlertCircle,
  RefreshCw,
  Loader2,
  ChevronRight,
  ChevronDown,
  FolderOpen,
  Save,
  FileCode,
  StepForward,
  Zap,
  Maximize2,
  Minimize2,
  Trash2,
} from 'lucide-react';
import { models } from '../../wailsjs/go/models';
import {
  ExecuteQuery,
  ExecutePLSQL,
  GetSchemaObjects,
  GetTableSchema,
  OpenSQLFile,
  SaveSQLFile,
} from '../../wailsjs/go/main/App';
import { useI18n } from '../i18n/LanguageContext';

export interface SQLStatement {
  index: number;
  text: string;
  rawText: string;
  start: number;
  end: number;
}

export function parseSQLStatements(sql: string): SQLStatement[] {
  const statements: SQLStatement[] = [];
  const len = sql.length;
  let i = 0;
  let stmtStart = 0;
  let inSingleQuote = false;
  let inDoubleQuote = false;
  let inLineComment = false;
  let inBlockComment = false;

  while (i < len) {
    const ch = sql[i];
    const next = i + 1 < len ? sql[i + 1] : '';

    if (inLineComment) {
      if (ch === '\n') {
        inLineComment = false;
      }
      i++;
      continue;
    }

    if (inBlockComment) {
      if (ch === '*' && next === '/') {
        inBlockComment = false;
        i += 2;
        continue;
      }
      i++;
      continue;
    }

    if (inSingleQuote) {
      if (ch === "'") {
        if (next === "'") {
          i += 2;
          continue;
        }
        inSingleQuote = false;
      }
      i++;
      continue;
    }

    if (inDoubleQuote) {
      if (ch === '"') {
        inDoubleQuote = false;
      }
      i++;
      continue;
    }

    if (ch === '-' && next === '-') {
      inLineComment = true;
      i += 2;
      continue;
    }
    if (ch === '/' && next === '*') {
      inBlockComment = true;
      i += 2;
      continue;
    }

    if (ch === "'") {
      inSingleQuote = true;
      i++;
      continue;
    }
    if (ch === '"') {
      inDoubleQuote = true;
      i++;
      continue;
    }

    let isDelimiter = false;
    let delimiterLen = 1;

    if (ch === ';') {
      isDelimiter = true;
      delimiterLen = 1;
    } else if (ch === '/') {
      let prevIdx = i - 1;
      let startOfLine = true;
      while (prevIdx >= stmtStart && sql[prevIdx] !== '\n') {
        if (sql[prevIdx] !== ' ' && sql[prevIdx] !== '\t' && sql[prevIdx] !== '\r') {
          startOfLine = false;
          break;
        }
        prevIdx--;
      }
      if (startOfLine) {
        let nextIdx = i + 1;
        let endOfLine = true;
        while (nextIdx < len && sql[nextIdx] !== '\n') {
          if (sql[nextIdx] !== ' ' && sql[nextIdx] !== '\t' && sql[nextIdx] !== '\r') {
            endOfLine = false;
            break;
          }
          nextIdx++;
        }
        if (endOfLine) {
          isDelimiter = true;
          delimiterLen = 1;
        }
      }
    }

    if (isDelimiter) {
      const rawText = sql.slice(stmtStart, i);
      const cleanText = rawText.trim().replace(/[;/]+$/, '').trim();
      if (cleanText.length > 0) {
        statements.push({
          index: statements.length,
          text: cleanText,
          rawText,
          start: stmtStart,
          end: i + delimiterLen,
        });
      }
      stmtStart = i + delimiterLen;
      i += delimiterLen;
      continue;
    }

    i++;
  }

  if (stmtStart < len) {
    const rawText = sql.slice(stmtStart);
    const cleanText = rawText.trim().replace(/[;/]+$/, '').trim();
    if (cleanText.length > 0) {
      statements.push({
        index: statements.length,
        text: cleanText,
        rawText,
        start: stmtStart,
        end: len,
      });
    }
  }

  return statements;
}

interface ScriptStepResult {
  id: number;
  sql: string;
  result?: models.QueryResult;
  error?: string;
}

interface QueryStudioTabProps {
  activeConnection: models.ConnectionProfile | null;
}

export const QueryStudioTab: React.FC<QueryStudioTabProps> = ({ activeConnection }) => {
  const { t } = useI18n();
  const [mode, setMode] = useState<'sql' | 'plsql'>('sql');
  const [sqlCode, setSqlCode] = useState<string>(
    'SELECT TABLE_NAME, NUM_ROWS, STATUS FROM USER_TABLES ORDER BY TABLE_NAME;\n\nSELECT USERNAME, ACCOUNT_STATUS, CREATED FROM USER_USERS;'
  );
  const [plsqlCode, setPlsqlCode] = useState<string>(
`DECLARE
  v_version VARCHAR2(100);
BEGIN
  SELECT BANNER INTO v_version FROM V$VERSION WHERE ROWNUM = 1;
  DBMS_OUTPUT.PUT_LINE('--- KYBERIX ORA_MCP / PLSQL OUTPUT ---');
  DBMS_OUTPUT.PUT_LINE('Oracle Database: ' || v_version);
  DBMS_OUTPUT.PUT_LINE('Timestamp: ' || TO_CHAR(SYSDATE, 'YYYY-MM-DD HH24:MI:SS'));
  DBMS_OUTPUT.PUT_LINE('Status: OK (Connected & Verified)');
END;`
  );

  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const [cursorPos, setCursorPos] = useState<number>(0);

  // Script file tracking
  const [scriptFilePath, setScriptFilePath] = useState<string | null>(null);
  const [scriptFileName, setScriptFileName] = useState<string | null>(null);

  // Execution states
  const [isRunning, setIsRunning] = useState(false);
  const [queryResult, setQueryResult] = useState<models.QueryResult | null>(null);
  const [plsqlResult, setPlsqlResult] = useState<models.PLSQLResult | null>(null);
  const [errorMsg, setErrorMsg] = useState<string>('');

  // Stepped / Script results tracking
  const [scriptResults, setScriptResults] = useState<ScriptStepResult[]>([]);
  const [activeResultTab, setActiveResultTab] = useState<number>(0);

  // Schema exploration sidebar state
  const [schema, setSchema] = useState<models.SchemaInfo | null>(null);
  const [loadingSchema, setLoadingSchema] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedTable, setSelectedTable] = useState<string | null>(null);
  const [tableColumns, setTableColumns] = useState<models.ColumnInfo[]>([]);
  const [loadingCols, setLoadingCols] = useState(false);
  const [isCopied, setIsCopied] = useState(false);
  const [hasSelection, setHasSelection] = useState<boolean>(false);
  const [isExpandedEditor, setIsExpandedEditor] = useState<boolean>(false);

  // Parse SQL statements for step-by-step execution
  const statements = useMemo(() => parseSQLStatements(sqlCode), [sqlCode]);

  // Find statement at current cursor position
  const currentStmt = useMemo(() => {
    if (statements.length === 0) return null;
    const found = statements.find((s) => cursorPos >= s.start && cursorPos <= s.end);
    if (found) return found;
    if (cursorPos < statements[0].start) return statements[0];
    return statements[statements.length - 1];
  }, [statements, cursorPos]);

  const updateCursorPos = () => {
    if (textareaRef.current) {
      const start = textareaRef.current.selectionStart || 0;
      const end = textareaRef.current.selectionEnd || 0;
      setCursorPos(start);
      setHasSelection(end > start);
    }
  };

  const cursorLineCol = useMemo(() => {
    const currentCode = mode === 'sql' ? sqlCode : plsqlCode;
    const textBefore = currentCode.slice(0, cursorPos);
    const lines = textBefore.split('\n');
    const line = lines.length;
    const col = lines[lines.length - 1].length + 1;
    return { line, col };
  }, [cursorPos, sqlCode, plsqlCode, mode]);

  const selectStatement = (stmt: SQLStatement) => {
    if (textareaRef.current) {
      textareaRef.current.focus();
      textareaRef.current.setSelectionRange(stmt.start, stmt.end);
      setCursorPos(stmt.start);
    }
  };

  // Load schema when active connection changes
  useEffect(() => {
    if (activeConnection) {
      loadSchema();
    } else {
      setSchema(null);
    }
  }, [activeConnection?.id]);

  const loadSchema = async () => {
    if (!activeConnection) return;
    setLoadingSchema(true);
    try {
      const res = await GetSchemaObjects(activeConnection.id);
      setSchema(res);
    } catch (err: any) {
      console.error('Error fetching schema:', err);
    } finally {
      setLoadingSchema(false);
    }
  };

  const handleSelectTable = async (tbl: string) => {
    setSelectedTable(tbl);
    if (!activeConnection) return;
    setLoadingCols(true);
    try {
      const cols = await GetTableSchema(activeConnection.id, tbl);
      setTableColumns(cols || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoadingCols(false);
    }
  };

  // Open .sql file dialog
  const handleOpenScript = async () => {
    try {
      const res = await OpenSQLFile();
      if (res && res.content !== undefined) {
        setSqlCode(res.content);
        setScriptFilePath(res.path || '');
        setScriptFileName(res.name || '');
        setScriptResults([]);
        setQueryResult(null);
        setErrorMsg('');
        setCursorPos(0);
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  // Save .sql file dialog
  const handleSaveScript = async (saveAs: boolean = false) => {
    try {
      const targetPath = saveAs ? '' : (scriptFilePath || '');
      const res = await SaveSQLFile(targetPath, sqlCode);
      if (res && res.path) {
        setScriptFilePath(res.path);
        setScriptFileName(res.name);
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    }
  };

  // Execute Current Statement or Selected Text
  const handleExecuteCurrent = async () => {
    if (!activeConnection) {
      setErrorMsg(t.queryStudio.noActiveDb);
      return;
    }

    if (mode === 'plsql') {
      setIsRunning(true);
      setErrorMsg('');
      setPlsqlResult(null);
      try {
        const res = await ExecutePLSQL(activeConnection.id, plsqlCode);
        if (res.error) {
          setErrorMsg(res.error);
        }
        setPlsqlResult(res);
      } catch (err: any) {
        setErrorMsg(err?.message || String(err));
      } finally {
        setIsRunning(false);
      }
      return;
    }

    // SQL Mode
    let targetSql = '';
    const selectedText = textareaRef.current
      ? textareaRef.current.value.substring(textareaRef.current.selectionStart, textareaRef.current.selectionEnd).trim()
      : '';

    if (selectedText.length > 0) {
      targetSql = selectedText;
    } else if (currentStmt) {
      targetSql = currentStmt.text;
    } else {
      targetSql = sqlCode;
    }

    const cleanSql = targetSql.trim().replace(/[;/]+$/, '').trim();
    if (!cleanSql) return;

    setIsRunning(true);
    setErrorMsg('');
    setQueryResult(null);

    try {
      const res = await ExecuteQuery(activeConnection.id, cleanSql, 500);
      if (res.error) {
        setErrorMsg(res.error);
        setScriptResults([{ id: 1, sql: cleanSql, error: res.error }]);
      } else {
        setQueryResult(res);
        setScriptResults([{ id: 1, sql: cleanSql, result: res }]);
      }
      setActiveResultTab(0);
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    } finally {
      setIsRunning(false);
    }
  };

  // Execute Step and Advance to next statement
  const handleExecuteStepAndAdvance = async () => {
    if (!activeConnection) {
      setErrorMsg(t.queryStudio.noActiveDb);
      return;
    }

    const stmtToRun = currentStmt || (statements.length > 0 ? statements[0] : null);
    if (!stmtToRun) {
      await handleExecuteCurrent();
      return;
    }

    const cleanSql = stmtToRun.text.trim().replace(/[;/]+$/, '').trim();
    if (!cleanSql) return;

    setIsRunning(true);
    setErrorMsg('');
    setQueryResult(null);

    try {
      const res = await ExecuteQuery(activeConnection.id, cleanSql, 500);
      const stepIndex = stmtToRun.index;

      if (res.error) {
        setErrorMsg(res.error);
        setScriptResults((prev) => {
          const next = [...prev];
          next[stepIndex] = { id: stepIndex + 1, sql: cleanSql, error: res.error };
          return next;
        });
      } else {
        setQueryResult(res);
        setScriptResults((prev) => {
          const next = [...prev];
          next[stepIndex] = { id: stepIndex + 1, sql: cleanSql, result: res };
          return next;
        });
      }
      setActiveResultTab(stepIndex);

      // Advance cursor to next statement if available
      const nextIndex = stmtToRun.index + 1;
      if (nextIndex < statements.length) {
        selectStatement(statements[nextIndex]);
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    } finally {
      setIsRunning(false);
    }
  };

  // Execute Entire Script Sequentially
  const handleExecuteAllScript = async () => {
    if (!activeConnection) {
      setErrorMsg(t.queryStudio.noActiveDb);
      return;
    }

    if (statements.length === 0) {
      await handleExecuteCurrent();
      return;
    }

    setIsRunning(true);
    setErrorMsg('');
    setQueryResult(null);

    const results: ScriptStepResult[] = [];

    try {
      for (let i = 0; i < statements.length; i++) {
        const stmt = statements[i];
        const cleanSql = stmt.text.trim().replace(/[;/]+$/, '').trim();
        if (!cleanSql) continue;

        try {
          const res = await ExecuteQuery(activeConnection.id, cleanSql, 500);
          if (res.error) {
            results.push({ id: i + 1, sql: cleanSql, error: res.error });
          } else {
            results.push({ id: i + 1, sql: cleanSql, result: res });
          }
        } catch (stepErr: any) {
          results.push({ id: i + 1, sql: cleanSql, error: stepErr?.message || String(stepErr) });
        }
      }

      setScriptResults(results);

      // Display the first result or first error
      if (results.length > 0) {
        const firstActive = results[0];
        setActiveResultTab(0);
        if (firstActive.result) {
          setQueryResult(firstActive.result);
        }
        if (firstActive.error) {
          setErrorMsg(firstActive.error);
        }
      }
    } catch (err: any) {
      setErrorMsg(err?.message || String(err));
    } finally {
      setIsRunning(false);
    }
  };

  // Keyboard shortcuts: Cmd/Ctrl + Enter (Current) or Cmd/Ctrl + Shift + Enter (All)
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      if (e.shiftKey && mode === 'sql') {
        handleExecuteAllScript();
      } else {
        handleExecuteCurrent();
      }
    }
  };

  const handleSelectResultTab = (index: number) => {
    setActiveResultTab(index);
    const item = scriptResults[index];
    if (item) {
      setQueryResult(item.result || null);
      setErrorMsg(item.error || '');
    }
  };

  const handleExportCSV = () => {
    if (!queryResult || !queryResult.rows || queryResult.rows.length === 0) return;
    const cols = queryResult.columns;
    const header = cols.join(',');
    const rows = queryResult.rows.map((row) =>
      cols
        .map((c) => {
          const val = row[c] === null || row[c] === undefined ? '' : String(row[c]);
          return `"${val.replace(/"/g, '""')}"`;
        })
        .join(',')
    );
    const csvContent = 'data:text/csv;charset=utf-8,' + [header, ...rows].join('\n');
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', `query_result_${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const copyResults = () => {
    if (queryResult) {
      navigator.clipboard.writeText(JSON.stringify(queryResult.rows, null, 2));
      setIsCopied(true);
      setTimeout(() => setIsCopied(false), 2000);
    }
  };

  const filteredTables =
    schema?.tables.filter((t) => t.toLowerCase().includes(searchTerm.toLowerCase())) || [];

  return (
    <div className="h-full flex flex-row overflow-hidden bg-[#050505]">
      {/* Left Sidebar: Schema Explorer */}
      <div className="w-64 border-r border-[#2e2e2e] bg-[#0c0c0c] flex flex-col shrink-0">
        <div className="p-3 border-b border-[#2e2e2e] flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            <span className="eyebrow text-[10px]">{t.queryStudio.schemaTitle}</span>
          </div>
          <button
            onClick={loadSchema}
            disabled={loadingSchema || !activeConnection}
            className="text-[#a7a49c] hover:text-[#f2efe6] transition-colors cursor-pointer"
            title={t.queryStudio.loadingSchema}
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loadingSchema ? 'animate-spin' : ''}`} />
          </button>
        </div>

        <div className="p-2 border-b border-[#2e2e2e]">
          <input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder={t.queryStudio.searchSchemaPlaceholder}
            className="w-full bg-[#121212] border border-[#2e2e2e] px-2 py-1 text-xs text-[#f2efe6] font-mono focus:border-[#d9ff3f] focus:outline-none"
          />
        </div>

        {/* Tables list */}
        <div className="flex-1 overflow-y-auto p-1 font-mono text-xs divide-y divide-[#2e2e2e]/40">
          {!activeConnection ? (
            <div className="p-4 text-center text-[#a7a49c] text-[11px]">
              {t.queryStudio.emptySchema}
            </div>
          ) : loadingSchema ? (
            <div className="p-4 flex items-center justify-center gap-2 text-[#a7a49c] text-xs">
              <Loader2 className="w-3.5 h-3.5 animate-spin text-[#d9ff3f]" />
              <span>{t.queryStudio.loadingSchema}</span>
            </div>
          ) : filteredTables.length === 0 ? (
            <div className="p-4 text-center text-[#a7a49c] text-[11px]">
              {t.queryStudio.emptySchema}
            </div>
          ) : (
            filteredTables.map((tbl) => {
              const isSelected = selectedTable === tbl;
              return (
                <div key={tbl}>
                  <button
                    onClick={() => handleSelectTable(tbl)}
                    className={`w-full text-left px-2 py-1.5 flex items-center justify-between transition-colors cursor-pointer ${
                      isSelected
                        ? 'bg-[#171717] text-[#d9ff3f] font-bold border-l-2 border-[#d9ff3f]'
                        : 'text-[#f2efe6] hover:bg-[#121212]'
                    }`}
                  >
                    <div className="flex items-center gap-1.5 truncate">
                      <Table className="w-3 h-3 text-[#a7a49c] shrink-0" />
                      <span className="truncate">{tbl}</span>
                    </div>
                    {isSelected ? (
                      <ChevronDown className="w-3 h-3 shrink-0" />
                    ) : (
                      <ChevronRight className="w-3 h-3 shrink-0 opacity-40" />
                    )}
                  </button>

                  {/* Expanded Columns Details */}
                  {isSelected && (
                    <div className="bg-[#050505] p-2 border-y border-[#2e2e2e] space-y-1.5">
                      <div className="flex items-center gap-1 mb-1">
                        <button
                          onClick={() => {
                            setMode('sql');
                            setSqlCode(`SELECT * FROM ${tbl} FETCH FIRST 50 ROWS ONLY;`);
                          }}
                          className="text-[10px] text-[#d9ff3f] hover:underline font-bold"
                        >
                          [ SELECT * (50) ]
                        </button>
                      </div>

                      {loadingCols ? (
                        <div className="text-[10px] text-[#a7a49c] flex items-center gap-1">
                          <Loader2 className="w-2.5 h-2.5 animate-spin" />
                          <span>Leyendo columnas...</span>
                        </div>
                      ) : (
                        <div className="space-y-1 max-h-40 overflow-y-auto">
                          {tableColumns.map((c) => (
                            <div
                              key={c.name}
                              className="flex items-center justify-between text-[10px]"
                            >
                              <span
                                className={
                                  c.isPrimaryKey ? 'text-[#fbbf24] font-bold' : 'text-[#f2efe6]'
                                }
                              >
                                {c.isPrimaryKey && '🔑 '}
                                {c.name}
                              </span>
                              <span className="text-[#a7a49c]">{c.dataType}</span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              );
            })
          )}
        </div>
      </div>

      {/* Main Studio Area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Editor Tactical Header Bar */}
        <div className="flex flex-wrap items-center justify-between gap-3 px-3 py-1.5 bg-[#0a0a0a] border-b border-[#2e2e2e] select-none">
          {/* Left Group: Mode Segmented Switcher & File Hub */}
          <div className="flex items-center gap-3">
            {/* Segmented Control for SQL / PL/SQL */}
            <div className="inline-flex p-0.5 bg-[#141414] border border-[#2e2e2e]">
              <button
                type="button"
                onClick={() => setMode('sql')}
                className={`px-3 py-1 text-xs font-mono font-bold flex items-center gap-1.5 transition-all cursor-pointer ${
                  mode === 'sql'
                    ? 'bg-[#d9ff3f] text-[#050505] shadow-[2px_2px_0_#383838]'
                    : 'text-[#a7a49c] hover:text-[#f2efe6] hover:bg-[#1f1f1f]'
                }`}
                title={t.queryStudio.tabSql}
              >
                <Database className="w-3.5 h-3.5 shrink-0" />
                <span>SQL</span>
              </button>
              <button
                type="button"
                onClick={() => setMode('plsql')}
                className={`px-3 py-1 text-xs font-mono font-bold flex items-center gap-1.5 transition-all cursor-pointer ${
                  mode === 'plsql'
                    ? 'bg-[#a855f7] text-[#050505] shadow-[2px_2px_0_#383838]'
                    : 'text-[#a7a49c] hover:text-[#f2efe6] hover:bg-[#1f1f1f]'
                }`}
                title={t.queryStudio.tabPlsql}
              >
                <Terminal className="w-3.5 h-3.5 shrink-0" />
                <span>PL/SQL</span>
              </button>
            </div>

            {/* Script File Dock (SQL mode only) */}
            {mode === 'sql' && (
              <div className="flex items-center border border-[#2e2e2e] bg-[#121212] font-mono text-xs">
                {/* File Badge / Name */}
                <div
                  className="flex items-center gap-1.5 px-2.5 py-1 text-[#f2efe6] max-w-[200px] truncate border-r border-[#2e2e2e]"
                  title={scriptFilePath || t.queryStudio.noFileLoaded}
                >
                  <FileCode className="w-3.5 h-3.5 text-[#d9ff3f] shrink-0" />
                  <span className="truncate text-[11px] font-semibold">
                    {scriptFileName || t.queryStudio.noFileLoaded}
                  </span>
                </div>

                {/* Open file */}
                <button
                  type="button"
                  onClick={handleOpenScript}
                  className="p-1.5 text-[#a7a49c] hover:text-[#d9ff3f] hover:bg-[#1c1c1c] transition-colors cursor-pointer border-r border-[#2e2e2e]"
                  title={t.queryStudio.openScriptBtn}
                >
                  <FolderOpen className="w-3.5 h-3.5" />
                </button>

                {/* Save file */}
                <button
                  type="button"
                  onClick={() => handleSaveScript(false)}
                  className="p-1.5 text-[#a7a49c] hover:text-[#d9ff3f] hover:bg-[#1c1c1c] transition-colors cursor-pointer"
                  title={t.queryStudio.saveScriptBtn}
                >
                  <Save className="w-3.5 h-3.5" />
                </button>

                {/* Save As (if file has path) */}
                {scriptFilePath && (
                  <button
                    type="button"
                    onClick={() => handleSaveScript(true)}
                    className="px-2 py-1 text-[10px] text-[#a7a49c] hover:text-[#f2efe6] hover:bg-[#1c1c1c] border-l border-[#2e2e2e] transition-colors cursor-pointer uppercase font-bold"
                    title={t.queryStudio.saveAsScriptBtn}
                  >
                    {t.queryStudio.saveAsScriptBtn}
                  </button>
                )}
              </div>
            )}
          </div>

          {/* Right Group: Unified Execution Engine */}
          <div className="flex items-center gap-2">
            {/* Secondary Execution Group (SQL mode only) */}
            {mode === 'sql' && (
              <div className="flex items-center border border-[#2e2e2e] bg-[#121212]">
                {/* Step & Next */}
                <button
                  type="button"
                  onClick={handleExecuteStepAndAdvance}
                  disabled={isRunning || !activeConnection || statements.length === 0}
                  className="flex items-center gap-1.5 px-2.5 py-1 text-xs font-mono font-bold text-[#f2efe6] hover:text-[#d9ff3f] hover:bg-[#1a1a1a] transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer border-r border-[#2e2e2e]"
                  title="Ejecutar sentencia actual y avanzar cursor (Paso a paso)"
                >
                  <StepForward className="w-3.5 h-3.5 text-[#d9ff3f]" />
                  <span>{t.queryStudio.runStepBtn}</span>
                </button>

                {/* Run All Script */}
                <button
                  type="button"
                  onClick={handleExecuteAllScript}
                  disabled={isRunning || !activeConnection || statements.length === 0}
                  className="flex items-center gap-1.5 px-2.5 py-1 text-xs font-mono font-bold text-[#f2efe6] hover:text-[#fbbf24] hover:bg-[#1a1a1a] transition-colors disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer"
                  title="Ejecutar todas las sentencias en secuencia (⌘⇧↵)"
                >
                  <Zap className="w-3.5 h-3.5 text-[#fbbf24] fill-current" />
                  <span>{t.queryStudio.runScriptBtn}</span>
                </button>
              </div>
            )}

            {/* Primary HERO Action Button */}
            <button
              type="button"
              onClick={handleExecuteCurrent}
              disabled={isRunning || !activeConnection}
              className={`brutal-button ${
                mode === 'sql' ? 'brutal-button--acid' : 'brutal-button--purple'
              } min-h-[32px] py-1 px-3 text-xs font-black shadow-[2px_2px_0_#383838] flex items-center gap-2`}
              title="Cmd/Ctrl + Enter"
            >
              {isRunning ? (
                <>
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                  <span>{t.queryStudio.runningBtn}</span>
                </>
              ) : (
                <>
                  <Play className="w-3.5 h-3.5 fill-current" />
                  <span>
                    {mode === 'sql'
                      ? hasSelection
                        ? t.queryStudio.runSelectionBtn
                        : t.queryStudio.runCurrentBtn
                      : t.queryStudio.runPlsqlBtn}
                  </span>
                  <span className="ml-0.5 px-1 py-0.2 bg-black/20 text-[9px] font-mono font-bold tracking-tight rounded">
                    ⌘↵
                  </span>
                </>
              )}
            </button>
          </div>
        </div>

        {/* Code Editor Panel */}
        <div
          className={`${
            isExpandedEditor ? 'h-80' : 'h-48'
          } bg-[#000000] relative flex transition-all duration-150`}
        >
          <textarea
            ref={textareaRef}
            value={mode === 'sql' ? sqlCode : plsqlCode}
            onChange={(e) => {
              if (mode === 'sql') setSqlCode(e.target.value);
              else setPlsqlCode(e.target.value);
              updateCursorPos();
            }}
            onClick={updateCursorPos}
            onKeyUp={updateCursorPos}
            onSelect={updateCursorPos}
            onKeyDown={handleKeyDown}
            placeholder={
              mode === 'sql'
                ? 'SELECT * FROM ...;\nSELECT * FROM ...;'
                : 'BEGIN\n  DBMS_OUTPUT.PUT_LINE(...);\nEND;'
            }
            className="w-full h-full p-3.5 font-mono text-xs text-[#f2efe6] bg-transparent resize-none focus:outline-none leading-relaxed selection:bg-[#d9ff3f] selection:text-black"
            spellCheck={false}
          />
        </div>

        {/* Tactical Editor Status Bar */}
        <div className="bg-[#0c0c0c] border-y border-[#2e2e2e] px-3 py-1 flex items-center justify-between text-[10px] font-mono text-[#a7a49c] select-none">
          {/* Left: Position & Statement context */}
          <div className="flex items-center gap-3">
            <span>{t.queryStudio.cursorPos(cursorLineCol.line, cursorLineCol.col)}</span>
            <span className="text-[#383838]">|</span>
            {mode === 'sql' && (
              <>
                <span>{t.queryStudio.statementsDetected(statements.length)}</span>
                {statements.length > 0 && currentStmt && (
                  <span className="px-1.5 py-0.2 bg-[#171717] border border-[#2e2e2e] text-[#d9ff3f] font-bold">
                    {t.queryStudio.statementCounter(currentStmt.index + 1, statements.length)}
                  </span>
                )}
              </>
            )}
          </div>

          {/* Right: Quick actions & resize */}
          <div className="flex items-center gap-3">
            <span className="text-[#737373] hidden sm:inline">
              {t.queryStudio.shortcutHint}
            </span>
            <span className="text-[#383838] hidden sm:inline">|</span>
            <button
              type="button"
              onClick={() => {
                if (mode === 'sql') setSqlCode('');
                else setPlsqlCode('');
              }}
              className="text-[#a7a49c] hover:text-[#fb7185] transition-colors cursor-pointer"
              title="Limpiar editor"
            >
              {t.queryStudio.clearBuffer}
            </button>
            <span className="text-[#383838]">|</span>
            <button
              type="button"
              onClick={() => setIsExpandedEditor(!isExpandedEditor)}
              className="text-[#a7a49c] hover:text-[#d9ff3f] transition-colors cursor-pointer flex items-center gap-1"
              title={isExpandedEditor ? 'Reducir editor' : 'Expandir editor'}
            >
              {isExpandedEditor ? (
                <Minimize2 className="w-3 h-3" />
              ) : (
                <Maximize2 className="w-3 h-3" />
              )}
            </button>
          </div>
        </div>

        {/* Multi-Query Stepped Results Tab Bar */}
        {mode === 'sql' && scriptResults.length > 1 && (
          <div className="flex items-center gap-1 px-3 py-1.5 bg-[#0a0a0a] border-b border-[#2e2e2e] overflow-x-auto">
            <span className="text-[10px] font-mono text-[#a7a49c] uppercase font-bold shrink-0 mr-2">
              {t.queryStudio.scriptResultsTitle}:
            </span>
            {scriptResults.map((item, idx) => {
              const isSelected = activeResultTab === idx;
              const hasError = !!item.error;
              return (
                <button
                  key={idx}
                  onClick={() => handleSelectResultTab(idx)}
                  className={`px-2.5 py-1 text-[11px] font-mono flex items-center gap-1.5 border transition-colors cursor-pointer shrink-0 ${
                    isSelected
                      ? hasError
                        ? 'border-[#e11d48] bg-[#e11d48]/20 text-[#fb7185] font-bold'
                        : 'border-[#d9ff3f] bg-[#d9ff3f]/10 text-[#d9ff3f] font-bold'
                      : hasError
                      ? 'border-[#2e2e2e] bg-[#121212] text-[#fb7185]/70 hover:border-[#e11d48]'
                      : 'border-[#2e2e2e] bg-[#121212] text-[#a7a49c] hover:text-[#f2efe6]'
                  }`}
                >
                  <span>STMT #{item.id}</span>
                  {hasError ? (
                    <span className="text-[9px] px-1 bg-[#e11d48] text-white font-black">ERR</span>
                  ) : item.result ? (
                    <span className="text-[9px] px-1 bg-[#1c2e12] text-[#d9ff3f] border border-[#d9ff3f]/40">
                      {item.result.rowCount}r ({item.result.executionMs}ms)
                    </span>
                  ) : null}
                </button>
              );
            })}
          </div>
        )}

        {/* Execution Output Header */}
        <div className="flex items-center justify-between px-4 py-2 bg-[#0c0c0c] border-b border-[#2e2e2e] text-xs font-mono">
          <div className="flex items-center gap-3">
            <span className="eyebrow text-[10px]">
              {mode === 'sql' ? '// SQL_OUTPUT' : `// ${t.queryStudio.plsqlConsoleTitle}`}
            </span>

            {queryResult && (
              <span className="px-2 py-0.5 bg-[#121212] border border-[#d9ff3f] text-[#d9ff3f] text-[10px] font-bold">
                {t.queryStudio.executionTelemetry(queryResult.rowCount, queryResult.executionMs)}
              </span>
            )}

            {plsqlResult && (
              <span
                className={`px-2 py-0.5 border text-[10px] font-bold ${
                  plsqlResult.success
                    ? 'border-[#d9ff3f] bg-[#121212] text-[#d9ff3f]'
                    : 'border-[#e11d48] bg-[#121212] text-[#fb7185]'
                }`}
              >
                {plsqlResult.success
                  ? t.queryStudio.plsqlSuccess(plsqlResult.executionMs)
                  : t.common.error}
              </span>
            )}
          </div>

          {/* Action buttons (Copy / CSV / Clear Console) */}
          <div className="flex items-center gap-2">
            {mode === 'plsql' && plsqlResult && (
              <button
                type="button"
                onClick={() => setPlsqlResult(null)}
                className="brutal-button brutal-button--ghost py-0.5 px-2 min-h-[26px] text-[10px] flex items-center gap-1.5"
                title={t.queryStudio.clearConsole}
              >
                <Trash2 className="w-3 h-3 text-[#a7a49c]" />
                <span>{t.queryStudio.clearConsole}</span>
              </button>
            )}

            {queryResult && queryResult.rows && queryResult.rows.length > 0 && (
              <>
                <button
                  type="button"
                  onClick={copyResults}
                  className="brutal-button brutal-button--ghost py-0.5 px-2 min-h-[26px] text-[10px] flex items-center gap-1.5"
                  title="Copy JSON"
                >
                  {isCopied ? (
                    <Check className="w-3 h-3 text-[#d9ff3f]" />
                  ) : (
                    <Copy className="w-3 h-3" />
                  )}
                  <span>{isCopied ? t.common.copied : 'JSON'}</span>
                </button>
                <button
                  type="button"
                  onClick={handleExportCSV}
                  className="brutal-button brutal-button--acid py-0.5 px-2 min-h-[26px] text-[10px] flex items-center gap-1.5 font-black"
                  title={t.queryStudio.exportCsvBtn}
                >
                  <Download className="w-3 h-3" />
                  <span>{t.queryStudio.exportCsvBtn}</span>
                </button>
              </>
            )}
          </div>
        </div>

        {/* Error Alert Display */}
        {errorMsg && (
          <div className="p-3 bg-[#e11d48]/10 border-b border-[#e11d48] text-[#fb7185] font-mono text-xs flex items-start gap-2">
            <AlertCircle className="w-4 h-4 shrink-0 mt-0.5" />
            <div>
              <div className="font-bold text-[11px] tracking-wider uppercase">
                {t.common.error}
              </div>
              <div className="text-xs whitespace-pre-wrap mt-0.5">{errorMsg}</div>
            </div>
          </div>
        )}

        {/* Results Viewer Area */}
        <div className="flex-1 overflow-auto bg-[#050505] relative">
          {mode === 'sql' ? (
            /* Tabular Grid View */
            queryResult && queryResult.rows && queryResult.rows.length > 0 ? (
              <table className="w-full text-left font-mono text-xs border-collapse">
                <thead className="bg-[#101010] sticky top-0 z-10 border-b border-[#2e2e2e]">
                  <tr>
                    <th className="py-2 px-3 text-[#a7a49c] text-[10px] uppercase font-bold border-r border-[#2e2e2e] w-12 text-center">
                      #
                    </th>
                    {queryResult.columns.map((col) => (
                      <th
                        key={col}
                        className="py-2 px-3 text-[#f2efe6] text-[11px] font-bold border-r border-[#2e2e2e] whitespace-nowrap"
                      >
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-[#2e2e2e]/50">
                  {queryResult.rows.map((row, idx) => (
                    <tr key={idx} className="hover:bg-[#121212] transition-colors">
                      <td className="py-1.5 px-3 text-[#a7a49c] text-[10px] border-r border-[#2e2e2e] text-center select-none">
                        {idx + 1}
                      </td>
                      {queryResult.columns.map((col) => (
                        <td
                          key={col}
                          className="py-1.5 px-3 text-[#f2efe6] border-r border-[#2e2e2e]/40 whitespace-nowrap max-w-xs truncate"
                        >
                          {row[col] === null || row[col] === undefined ? (
                            <span className="text-[#a7a49c] italic">null</span>
                          ) : (
                            String(row[col])
                          )}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : !isRunning && !errorMsg ? (
              <div className="h-full flex flex-col items-center justify-center text-[#a7a49c] font-mono text-xs p-6 text-center">
                <Database className="w-8 h-8 text-[#2e2e2e] mb-2" />
                <span>{t.queryStudio.emptyResults}</span>
              </div>
            ) : null
          ) : (
            /* PL/SQL Terminal Screen (DBMS_OUTPUT) */
            <div className="h-full bg-black p-4 font-mono text-xs overflow-auto scanlines text-[#d9ff3f]">
              {plsqlResult ? (
                <div className="space-y-1">
                  <div className="text-[#a7a49c] pb-2 border-b border-[#2e2e2e]">
                    [DBMS_OUTPUT // {plsqlResult.output.length} LINES]
                  </div>
                  {plsqlResult.output.length === 0 ? (
                    <div className="text-[#a7a49c] italic py-2">
                      ({t.queryStudio.noDbmsOutput})
                    </div>
                  ) : (
                    plsqlResult.output.map((line, idx) => (
                      <div key={idx} className="leading-relaxed whitespace-pre-wrap">
                        <span className="text-[#a7a49c] select-none mr-2">
                          {String(idx + 1).padStart(3, '0')}:
                        </span>
                        <span>{line}</span>
                      </div>
                    ))
                  )}
                  {plsqlResult.rowsAffected > 0 && (
                    <div className="text-[#a855f7] pt-2 border-t border-[#2e2e2e]">
                      Rows affected: {plsqlResult.rowsAffected}
                    </div>
                  )}
                </div>
              ) : !isRunning && !errorMsg ? (
                <div className="h-full flex flex-col items-center justify-center text-[#a7a49c] text-xs text-center">
                  <Terminal className="w-8 h-8 text-[#2e2e2e] mb-2" />
                  <span>{t.queryStudio.tabPlsql}</span>
                </div>
              ) : null}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
