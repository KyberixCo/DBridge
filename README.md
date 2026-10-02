# dbridge — Tactical Oracle Database MCP Bridge & SQL Studio

<p align="center">
  <img src="DBridgeIcon%20Exports/icon_512x512.png" alt="dbridge Logo" width="128" height="128" />
</p>

<p align="center">
  <strong>High-performance, secure Model Context Protocol (MCP) bridge and tactical SQL studio for Oracle Database environments.</strong>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#architecture">Architecture</a> •
  <a href="#mcp-tools">MCP Tools</a> •
  <a href="#ai-client-setup">AI Setup</a> •
  <a href="#security--credential-storage">Security</a> •
  <a href="#installation--building">Building</a> •
  <a href="#license">License</a>
</p>

<p align="center">
  <strong>Language:</strong>
  <a href="README.md">English</a> |
  <a href="README.es.md">Español</a>
</p>

---

## Overview

**dbridge** (by **Kyberix**) is a desktop application and headless CLI server built with **Go**, **Wails v2**, **React 18**, **TypeScript**, and **Tailwind CSS**. Designed under the **Kyberix AEGIS Design System** (*Tactical Cyber-Minimalism*), it provides a military-grade bridge between local/remote Oracle Databases and AI coding assistants like Claude Desktop, Cursor, VS Code (GitHub Copilot), IntelliJ IDEA, Windsurf, and Continue.

### Why dbridge?

- **Zero Oracle Client Dependencies**: Pure Go database connectivity (`go-ora/v2`) — no CGO, no bulky Oracle Instant Client installations, no DLL/dylib headaches.
- **Enterprise-Grade Secret Security**: Passwords and credentials never sit in plaintext on disk; they are safeguarded directly inside the operating system's native vaults (**macOS Keychain** & **Windows Credential Manager**).
- **Hardened MCP Firewall**: Prevents rogue AI agent queries through a strict Read-Only mode, configurable blocked keywords, PL/SQL sandboxing, and context row limits.
- **Full Observability**: Live telemetry and audit log stream inspecting every tool invocation, latency measurement, query payload, and agent client ID.
- **Tactical SQL & PL/SQL Studio**: In-app schema browser and SQL query runner with real-time `DBMS_OUTPUT` streaming for debugging and schema verification.
- **Bilingual Experience**: Built-in English and Spanish localization with automatic system locale detection and persistent settings.

---

## Features

### 1. Pure Go Oracle Engine
- Direct TCP/TCPS connection support for **Oracle 11g, 12c, 19c, 21c, and 23ai**.
- Supports both modern `SERVICE_NAME` and legacy `SID` routing.
- **Assisted Configuration Import (`tnsnames.ora` & DBeaver)**:
  - Discovers and parses `tnsnames.ora` files (from `$TNS_ADMIN`, `$ORACLE_HOME`, or file picker), auto-populating Host, Port, Service Name/SID, and TCPS protocol.
  - Automatically imports Oracle connection definitions directly from your existing **DBeaver** workspace (`data-sources.json`).
- Native SSL/TLS (TCPS) encryption and **Oracle Wallet** (`cwallet.sso`) directory support for Oracle Autonomous Cloud Databases (ATP / ADW).
- Real-time connection testing with latency benchmarking and Oracle banner identification.

### 2. Multi-Layer MCP Security Firewall
- **Read-Only Enforcement (SELECT Only)**: Automatically denies destructive DDL and DML statements (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`, `MERGE`, etc.).
- **Comment-Stripping Lexer**: Cleans comments (`--` and `/* ... */`) and string literals before validation to defeat bypass vectors.
- **Dynamic Keyword Blacklist**: Custom blacklists configurable directly from the interface.
- **PL/SQL Guardrail**: Toggle permission for agents to invoke anonymous PL/SQL execution (`oracle_execute_plsql`).
- **Context Protection Window**: Configurable maximum row count threshold to avoid flooding LLM token context windows.
- **Interactive Policy Simulator**: Test and validate how security policies evaluate arbitrary SQL queries in real-time.

### 3. Dual-Transport MCP Server
- **HTTP / Server-Sent Events (SSE)**: Configurable port (default `:8085`):
  - `GET /sse`: Persistent Server-Sent Events stream.
  - `POST /message?sessionId=...`: Standard JSON-RPC 2.0 message receiver.
  - `POST /mcp`: Direct synchronous HTTP JSON-RPC endpoint.
- **Headless CLI Stdio Transport**:
  - Run directly from command line or automated scripts: `./dbridge --mcp` or `./dbridge stdio`.

### 4. Assisted 1-Click AI Integrations
- **Claude Desktop**: One-click configuration writing to `claude_desktop_config.json`.
- **VS Code / GitHub Copilot**: One-click configuration generator for `.vscode/mcp.json`.
- **Cursor IDE**: One-click setup for `~/.cursor/mcp.json`.
- **IntelliJ IDEA / JetBrains**: Step-by-step assisted guide for native Model Context Protocol SSE setup.

### 5. Real-Time Telemetry & Audit Logs
- Captures every AI tool invocation with precise microsecond timing, tool name, client identifier, status (ALLOWED / BLOCKED), and full parameter payloads.
- Streamed in real-time to the frontend via the Wails event bus (`mcp:audit`).
- Detailed tactical inspection drawer with reason codes for blocked attempts.

### 6. SQL & PL/SQL Studio
- Interactive schema tree (Tables, Views, Stored Procedures, and Packages).
- Column definitions, Oracle data types, nullability, and primary key indicators.
- Tabular data grid with execution telemetry and CSV export.
- Retro CRT-style PL/SQL console capturing `DBMS_OUTPUT.PUT_LINE` output in real time.

---

## MCP Tools

dbridge exposes the following Model Context Protocol tools to connected AI models:

| Tool Name | Description | Parameters |
| :--- | :--- | :--- |
| `oracle_query` | Executes a sanitized SQL query against the active database. | `query` (string, required), `max_rows` (number, optional) |
| `oracle_execute_plsql` | Executes an anonymous PL/SQL block and captures `DBMS_OUTPUT`. | `block` (string, required) |
| `oracle_list_tables` | Lists all tables, views, and packages in the current database schema. | *none* |
| `oracle_describe_table`| Fetches table metadata: column names, types, nullable status, and PKs. | `table_name` (string, required) |
| `oracle_list_connections` | Lists configured database connection profiles and indicates active target. | *none* |
| `oracle_switch_connection` | Hot-switches the active database target used by MCP calls. | `connection_id` (string, required) |

### Direct STDIO and Copilot inspection mode

The same binary opens the GUI without arguments and runs MCP with `--mcp` or `stdio`. No Python or shell wrapper is required. VS Code auto-configuration writes `type: "stdio"` and the absolute executable path. The **Inspection** checkbox adds `--inspect`.

Example `.vscode/mcp.json` (replace the path and profile ID):

```json
{
  "servers": {
    "dbridge-oracle": {
      "type": "stdio",
      "command": "/absolute/path/dbridge",
      "args": ["--mcp", "--inspect", "--connection", "PROFILE_ID", "--timeout", "30s"]
    }
  }
}
```

On Windows, use the absolute path to `dbridge.exe`. `--connection` accepts an existing profile ID and requires `--inspect`. Without it, inspection pins the active profile at process startup; restart to choose another. It does not change the saved active connection. Only `oracle_list_tables`, `oracle_describe_table` and `oracle_query` are exposed; direct calls to hidden tools are also rejected.

Inspection overrides write permissions, disables PL/SQL, and rejects database links and direct routine calls outside a conservative list of built-in SQL functions. `COUNT`, `NVL` and `UPPER` are supported; custom functions, schema-qualified routines and `UTL_HTTP.REQUEST` are blocked. Some valid SQL expressions, such as CTEs with explicit column lists, must be simplified in this mode.

MCP queries in inspection or `read_only` policy run in a transaction with `SET TRANSACTION READ ONLY` and always roll back. Queries fail closed if Oracle cannot establish this mode. Also use an Oracle user with `CREATE SESSION` and only the required read grants, without DML/DDL permissions or execution grants on write/network routines. Review views and synonyms: indirectly invoked autonomous routines can bypass transaction restrictions. Consequently, `oracle_query` retains `readOnlyHint: false`, even in inspection mode.

Schema and connection inspection tools advertise read-only, idempotent behavior. SQL and PL/SQL conservatively advertise potential destructive effects and external access; switching advertises an idempotent local state change. Annotations describe behavior; they neither grant permissions nor bypass client approvals. Table resources return schema metadata only, never data rows. Server instructions direct the agent to MCP tools and identify database results as data rather than instructions.

Tool arguments are validated on the server. `max_rows` must be a positive integer within the policy limit; its default is the smaller of 100 and that limit. A policy without a positive limit falls back to 500. Multiple SQL statements and incomplete literals/comments are rejected. Read-only validation also blocks `NEXTVAL`, `FOR UPDATE`, and PL/SQL functions in CTEs. `custom` and `full` policies retain their configured permissions for single SQL statements.

MCP operations default to a 30-second deadline; `--timeout` accepts positive durations up to `10m`. STDIO supports `notifications/cancelled`, continues reading during operations, and suppresses responses to explicitly cancelled requests. Cancellation propagates to the driver and cannot undo already committed operations in other modes. The pending queue is capped at 32 requests; input frames are capped at 10 MiB. Diagnostics go to stderr, stdout contains only JSON-RPC, and transport read/write failures are surfaced.

After replacing the executable, restart it through **MCP: List Servers** and run **MCP: Reset Cached Tools**. Check that Copilot calls MCP tools instead of running Python in the terminal. VS Code controls trust and tool approvals. Optional `sandboxEnabled` support on macOS/Linux needs access rules compatible with `.oramcp`, wallets, the system credential store and Oracle hosts; validate those dependencies before enabling it. See the [official VS Code MCP configuration reference](https://code.visualstudio.com/docs/agents/reference/mcp-configuration).

---

## AI Client Setup

### Claude Desktop

#### Option A: 1-Click Setup (Inside dbridge)
Open dbridge, click **SSE :8085** in the top bar, go to the **1-Click Install** tab, and click **Configure Claude Desktop**.

#### Option B: Manual Configuration
Add this to your `claude_desktop_config.json`:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**SSE Mode (Recommended while dbridge UI is open):**
```json
{
  "mcpServers": {
    "dbridge-oracle": {
      "url": "http://localhost:8085/sse"
    }
  }
}
```

**Stdio Mode (Headless binary execution):**
```json
{
  "mcpServers": {
    "dbridge-oracle": {
      "command": "/Applications/dbridge.app/Contents/MacOS/dbridge",
      "args": ["--mcp"]
    }
  }
}
```

---

### VS Code & GitHub Copilot

Configuration can be placed globally in `%APPDATA%\Code\User\mcp.json` (Windows) or `~/Library/Application Support/Code/User/mcp.json` (macOS), or per-project in `.vscode/mcp.json`.

**Option 1: CLI Command Mode (Recommended — runs headless via PATH without keeping the app open):**
```json
{
  "servers": {
    "dbridge-oracle": {
      "command": "dbridge",
      "args": ["--mcp"]
    }
  }
}
```

**Option 2: Network SSE Mode (Connects to running dbridge GUI app):**
```json
{
  "servers": {
    "dbridge-oracle": {
      "type": "sse",
      "url": "http://localhost:8085/sse"
    }
  }
}
```

---

### Cursor IDE

Add this to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dbridge-oracle": {
      "url": "http://localhost:8085/sse"
    }
  }
}
```

---

### IntelliJ IDEA / JetBrains

1. Open **Settings / Preferences** (`Cmd+,` on macOS, `Ctrl+Alt+S` on Windows).
2. Navigate to **Tools** → **Model Context Protocol (MCP)**.
3. Click **+ (Add)** and select **Server-Sent Events (SSE)**.
4. Set **Name** to `dbridge-oracle`.
5. Set **URL** to `http://localhost:8085/sse`.
6. Save and apply.

---

## Security & Credential Storage

Database passwords and secret tokens are never written in plaintext to configuration files. dbridge integrates natively with operating system secret vaults:

- **macOS**: Native **Apple Keychain Services** (via `security` subsystem).
- **Windows**: Native **Windows Credential Manager** (`wincred` subsystem).
- **Linux / Fallback**: AES-256 encrypted file vault stored in `~/.dbridge/secrets.enc` with strict file permissions (`0600`).

Metadata configuration is persisted at `~/.dbridge/config.json`.

---

## Architecture

```
dbridge/
├── app.go                 # Wails Application Bridge & IPC Controllers
├── main.go                # Application Entrypoint & CLI Flag Dispatcher
├── backend/
│   ├── audit/             # Real-time Telemetry & In-Memory Audit Buffer
│   ├── config/            # Connection Profiles & Persistence Management
│   ├── db/                # Pure Go Oracle Engine (go-ora/v2) & PL/SQL Runner
│   ├── mcp/               # Model Context Protocol Server (SSE & Stdio)
│   ├── security/          # SQL Firewall, Tokenizer, & OS Vault Service
│   └── tests/             # Security and Unit Test Suites
├── frontend/
│   ├── src/
│   │   ├── components/    # Brutalist UI Views & Modals (Connections, Studio, etc.)
│   │   ├── i18n/          # Internationalization Engine (English / Español)
│   │   ├── App.tsx        # Main Application Container
│   │   └── main.tsx       # React 18 Mounting
└── build/                 # Wails Packaging Assets, Icons, and Binaries
```

---

## Installation & Building

### Prerequisites

- **Go**: 1.22 or higher
- **Node.js**: 18+ and npm
- **Wails CLI v2**:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### Development Mode

Run live reload for both backend Go code and React frontend:

```bash
# Clone the repository
git clone https://github.com/Kyberix/dbridge.git
cd dbridge

# Start development server with live reload
wails dev
```

### Production Build

Package self-contained binaries:

```bash
# Build for current OS (macOS or Windows)
wails build -clean

# Output binary location:
# macOS:   build/bin/dbridge.app
# Windows: build/bin/dbridge.exe
```

### Running Test Suite

```bash
go test -v ./backend/...
```

---

## License

This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

See the [LICENSE](LICENSE) file for the complete license terms.

---

<p align="center">
  <strong>KYBERIX // ADVANCED TACTICAL SOFTWARE</strong><br />
  <em>Empowering developers and AI agents with robust, zero-compromise database connectivity.</em>
</p>
