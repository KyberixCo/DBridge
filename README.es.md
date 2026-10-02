# dbridge — Puente MCP Táctico para Oracle Database & SQL Studio

<p align="center">
  <img src="DBridgeIcon%20Exports/icon_512x512.png" alt="Logo de dbridge" width="128" height="128" />
</p>

<p align="center">
  <strong>Puente Model Context Protocol (MCP) seguro y de alto rendimiento junto a un estudio SQL táctico para bases de datos Oracle.</strong>
</p>

<p align="center">
  <a href="#características">Características</a> •
  <a href="#arquitectura">Arquitectura</a> •
  <a href="#herramientas-mcp">Herramientas MCP</a> •
  <a href="#configuración-con-asistentes-ia">Configuración IA</a> •
  <a href="#seguridad-y-bóvedas-nativas">Seguridad</a> •
  <a href="#instalación-y-compilación">Compilación</a> •
  <a href="#licencia">Licencia</a>
</p>

<p align="center">
  <strong>Idioma / Language:</strong>
  <a href="README.md">English</a> |
  <a href="README.es.md">Español</a>
</p>

---

## Visión General

**dbridge** (desarrollado por **Kyberix**) es una aplicación de escritorio y servidor CLI headless desarrollada con **Go**, **Wails v2**, **React 18**, **TypeScript** y **Tailwind CSS**. Concebida bajo el sistema de diseño **Kyberix AEGIS** (*Cyber-Minimalismo Táctico*), proporciona un puente de grado militar entre bases de datos Oracle (locales o remotas) y asistentes de programación con IA como Claude Desktop, Cursor, VS Code (GitHub Copilot), IntelliJ IDEA, Windsurf y Continue.

### ¿Por qué dbridge?

- **Cero dependencias de clientes Oracle externos**: Conectividad nativa pura en Go (`go-ora/v2`) — sin requerir CGO, sin descargar Oracle Instant Client y sin problemas de DLLs o dylibs.
- **Seguridad nativa para contraseñas**: Las credenciales jamás se almacenan en texto plano en disco; se resguardan de forma encriptada en los llaveros nativos del sistema operativo (**macOS Keychain** y **Windows Credential Manager**).
- **Firewall MCP endurecido**: Protege la integridad de la base de datos contra acciones destructivas de los agentes IA mediante modo Solo Lectura estricto, palabras clave bloqueadas configurables, control granular de PL/SQL y límites de filas por consulta.
- **Telemetría y Auditoría en Tiempo Real**: Inspección total de cada invocación de herramientas, medición de latencia, sentencias ejecutadas y clientes conectados.
- **Estudio SQL & PL/SQL Integrado**: Explorador de esquemas y ejecutor de consultas con soporte en vivo para la captura de salidas de `DBMS_OUTPUT.PUT_LINE`.
- **Experiencia Bilingüe Completa**: Localización nativa en Inglés y Español con detección automática del idioma del sistema y persistencia local.

---

## Características

### 1. Motor Oracle en Go Puro
- Conexión TCP y TCPS directa con versiones **Oracle 11g, 12c, 19c, 21c y 23ai**.
- Soporte para identificadores modernos `SERVICE_NAME` y tradicionales `SID`.
- **Importación Asistida (`tnsnames.ora` & DBeaver)**:
  - Detecta y analiza archivos `tnsnames.ora` (desde `$TNS_ADMIN`, `$ORACLE_HOME` o selector de archivos), autocompletando Host, Puerto, Service Name/SID y protocolo TCPS.
  - Importa perfiles de conexión Oracle existentes directamente desde tu workspace de **DBeaver** (`data-sources.json`).
- Encriptación SSL/TLS nativa y soporte para carpetas con **Oracle Wallet** (`cwallet.sso`) para conexiones seguras con Oracle Autonomous Cloud (ATP / ADW).
- Prueba interactiva de conexión con cálculo de latencia e identificación del banner de versión de la base de datos.

### 2. Firewall de Seguridad MCP Multicapa
- **Modo Solo Lectura Estricto (SELECT Only)**: Bloquea automáticamente sentencias DDL y DML destructivas (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`, `MERGE`, etc.).
- **Lexer con Eliminación de Comentarios**: Limpia comentarios (`--` y `/* ... */`) y cadenas literales antes de evaluar las reglas, evitando técnicas de evasión de seguridad.
- **Lista Dinámica de Palabras Clave Bloqueadas**: Personalizable directamente desde la interfaz.
- **Control de Ejecución PL/SQL**: Habilita o deshabilita la capacidad de los agentes IA para correr bloques anónimos PL/SQL (`oracle_execute_plsql`).
- **Límite de Ventana de Filas**: Techo máximo de registros por consulta para no saturar la ventana de contexto de los modelos LLM.
- **Simulador de Políticas en Vivo**: Permite probar cómo reacciona el firewall ante cualquier sentencia SQL arbitraria en tiempo real.

### 3. Doble Transporte para el Servidor MCP
- **HTTP / Server-Sent Events (SSE)** en puerto configurable (por defecto `:8085`):
  - `GET /sse`: Stream persistente de eventos SSE.
  - `POST /message?sessionId=...`: Receptor de mensajes estándar JSON-RPC 2.0.
  - `POST /mcp`: Endpoint HTTP directo para llamadas JSON-RPC.
- **Transporte CLI Stdio (Modo Headless)**:
  - Ejecutable desde terminal o entornos automatizados: `./dbridge --mcp` o `./dbridge stdio`.

### 4. Integraciones Asistidas en 1-Clic
- **Claude Desktop**: Configuración automática con un solo clic en `claude_desktop_config.json`.
- **VS Code / GitHub Copilot**: Generación y registro automático en `.vscode/mcp.json`.
- **Cursor IDE**: Configuración directa en `~/.cursor/mcp.json`.
- **IntelliJ IDEA / JetBrains**: Guía paso a paso para configurar el soporte nativo de MCP vía SSE.

### 5. Telemetría y Registros de Auditoría en Tiempo Real
- Captura de cada invocación de herramientas con marca de tiempo precisa en microsegundos, nombre de la herramienta, identificador del cliente/agente, estado (PERMITIDA / BLOQUEADA) y parámetros completos.
- Transmisión en tiempo real al frontend mediante el bus de eventos de Wails (`mcp:audit`).
- Panel lateral deslizable para inspeccionar detalladamente el motivo de bloqueo o los resultados.

### 6. Estudio de Consultas SQL & PL/SQL
- Árbol interactivo de objetos del esquema (Tablas, Vistas, Procedimientos y Paquetes).
- Detalle de columnas, tipos de datos de Oracle, nulabilidad e indicadores de clave primaria (PK).
- Grilla de resultados con telemetría de ejecución y exportación a CSV.
- Consola CRT de ejecución PL/SQL con captura en vivo de `DBMS_OUTPUT.PUT_LINE`.

---

## Herramientas MCP

dbridge expone las siguientes herramientas del Model Context Protocol a los modelos IA conectados:

| Nombre de Herramienta | Descripción | Parámetros |
| :--- | :--- | :--- |
| `oracle_query` | Ejecuta una consulta SQL sanitizada sujeta al filtro de seguridad. | `query` (cadena, obligatoria), `max_rows` (número, opcional) |
| `oracle_execute_plsql` | Ejecuta un bloque anónimo PL/SQL y captura la salida `DBMS_OUTPUT`. | `block` (cadena, obligatoria) |
| `oracle_list_tables` | Lista todas las tablas, vistas y paquetes del esquema activo. | *ninguno* |
| `oracle_describe_table`| Obtiene metadatos de una tabla: columnas, tipos, nulabilidad y PKs. | `table_name` (cadena, obligatoria) |
| `oracle_list_connections` | Lista los perfiles de conexión configurados e indica el activo. | *ninguno* |
| `oracle_switch_connection` | Cambia la base de datos activa utilizada por el servidor MCP. | `connection_id` (cadena, obligatoria) |

El servidor marca `oracle_list_tables`, `oracle_describe_table` y `oracle_list_connections` como herramientas de solo lectura para que los clientes MCP puedan tratarlas como consultas. `oracle_query`, `oracle_execute_plsql` y `oracle_switch_connection` no llevan esa marca: su efecto depende de la operación y de la política configurada. Estas anotaciones informan al cliente; la política de dbridge sigue determinando qué se permite.

Los recursos MCP anunciados para las tablas del esquema activo se pueden abrir con `resources/read`. Devuelven nombres de columnas, tipos, nulabilidad y claves primarias en JSON, sin filas de datos. El servidor también indica al agente que utilice las herramientas MCP para Oracle. Las advertencias de VS Code por scripts Python o comandos de terminal pertenecen al agente y a la configuración de aprobaciones de VS Code; dbridge no puede suprimirlas.

---

## Configuración con Asistentes IA

### Claude Desktop

#### Opción A: Configuración en 1-Clic (Dentro de dbridge)
Abre dbridge, haz clic en **SSE :8085** en la barra superior, ve a la pestaña **Instalación 1-Clic** y pulsa **Configurar Claude Desktop**.

#### Opción B: Configuración Manual
Edita tu archivo `claude_desktop_config.json`:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**Modo SSE (Recomendado con dbridge abierto):**
```json
{
  "mcpServers": {
    "dbridge-oracle": {
      "url": "http://localhost:8085/sse"
    }
  }
}
```

**Modo Stdio (Ejecución binaria headless):**
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

La configuración puede colocarse a nivel global en `%APPDATA%\Code\User\mcp.json` (Windows) o `~/Library/Application Support/Code/User/mcp.json` (macOS), o por proyecto en `.vscode/mcp.json`.

**Opción 1: Modo Comando CLI (Recomendado — corre en segundo plano vía PATH sin requerir la GUI abierta):**
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

**Opción 2: Modo Red SSE (Se conecta a la app dbridge abierta en puerto :8085):**
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

Agrega lo siguiente en `~/.cursor/mcp.json`:

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

1. Abre **Ajustes / Preferencias** (`Cmd+,` en macOS, `Ctrl+Alt+S` en Windows).
2. Ve a **Tools** → **Model Context Protocol (MCP)**.
3. Haz clic en **+ (Agregar)** y selecciona **Server-Sent Events (SSE)**.
4. En **Name** escribe `dbridge-oracle`.
5. En **URL** coloca `http://localhost:8085/sse`.
6. Guarda y aplica los cambios.

---

## Seguridad y Bóvedas Nativas

Las contraseñas de las bases de datos nunca se guardan en texto plano en archivos de configuración. dbridge se integra de forma directa con los almacenes seguros del sistema operativo:

- **macOS**: **Apple Keychain Services** nativo (mediante el comando de seguridad del sistema).
- **Windows**: **Windows Credential Manager** nativo (mediante la API `wincred`).
- **Linux / Respaldo**: Bóveda de archivos cifrada con AES-256 en `~/.dbridge/secrets.enc` con permisos estrictos (`0600`).

La configuración de perfiles y metadatos se guarda en `~/.dbridge/config.json`.

---

## Arquitectura

```
dbridge/
├── app.go                 # Puente Wails IPC y controladores de eventos
├── main.go                # Punto de entrada y procesador de flags CLI
├── backend/
│   ├── audit/             # Telemetría y buffer en memoria de auditoría
│   ├── config/            # Perfiles de conexión y persistencia
│   ├── db/                # Motor Oracle en Go puro (go-ora/v2) y PL/SQL runner
│   ├── mcp/               # Servidor Model Context Protocol (SSE y Stdio)
│   ├── security/          # Firewall SQL, tokenizador y bóveda del SO
│   └── tests/             # Suites de pruebas de seguridad y lógica
├── frontend/
│   ├── src/
│   │   ├── components/    # Vistas y modales brutalistas (Conexiones, Studio, etc.)
│   │   ├── i18n/          # Motor de internacionalización (Inglés / Español)
│   │   ├── App.tsx        # Contenedor raíz de la aplicación
│   │   └── main.tsx       # Montaje de React 18
└── build/                 # Recursos de empaquetado Wails, iconos y binarios
```

---

## Instalación y Compilación

### Requisitos Previos

- **Go**: 1.22 o superior
- **Node.js**: 18+ y npm
- **Wails CLI v2**:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### Modo de Desarrollo

Ejecuta el entorno con recarga en caliente para Go y React:

```bash
# Clonar el repositorio
git clone https://github.com/Kyberix/dbridge.git
cd dbridge

# Iniciar servidor de desarrollo con live reload
wails dev
```

### Compilación para Producción

Genera los binarios empaquetados e instaladores:

```bash
# Compilar para el sistema operativo actual (macOS o Windows)
wails build -clean

# Ubicación del ejecutable generado:
# macOS:   build/bin/dbridge.app
# Windows: build/bin/dbridge.exe
```

### Ejecutar Pruebas Unitarias

```bash
go test -v ./backend/...
```

---

## Licencia

Este proyecto está licenciado bajo la **GNU General Public License v3.0 (GPL-3.0)**.

Consulta el archivo [LICENSE](LICENSE) para ver los términos y condiciones completos de la licencia.

---

<p align="center">
  <strong>KYBERIX // ADVANCED TACTICAL SOFTWARE</strong><br />
  <em>Potenciando a desarrolladores y agentes de IA con conectividad robusta y sin compromisos a bases de datos.</em>
</p>
