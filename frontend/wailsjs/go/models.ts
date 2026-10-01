export namespace models {
	
	export class AuditLogEntry {
	    id: string;
	    // Go type: time
	    timestamp: any;
	    clientInfo: string;
	    toolName: string;
	    query: string;
	    allowed: boolean;
	    reason: string;
	    executionMs: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuditLogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.clientInfo = source["clientInfo"];
	        this.toolName = source["toolName"];
	        this.query = source["query"];
	        this.allowed = source["allowed"];
	        this.reason = source["reason"];
	        this.executionMs = source["executionMs"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ColumnInfo {
	    name: string;
	    dataType: string;
	    dataLength: number;
	    nullable: boolean;
	    isPrimaryKey: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.dataType = source["dataType"];
	        this.dataLength = source["dataLength"];
	        this.nullable = source["nullable"];
	        this.isPrimaryKey = source["isPrimaryKey"];
	    }
	}
	export class ConnectionProfile {
	    id: string;
	    name: string;
	    host: string;
	    port: number;
	    serviceName: string;
	    sid: string;
	    isSid: boolean;
	    username: string;
	    password: string;
	    ssl: boolean;
	    walletPath: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.serviceName = source["serviceName"];
	        this.sid = source["sid"];
	        this.isSid = source["isSid"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.ssl = source["ssl"];
	        this.walletPath = source["walletPath"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConnectionTestResult {
	    success: boolean;
	    message: string;
	    serverVersion: string;
	    latencyMs: number;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.serverVersion = source["serverVersion"];
	        this.latencyMs = source["latencyMs"];
	    }
	}
	export class MCPServerStatus {
	    running: boolean;
	    port: number;
	    sseUrl: string;
	    httpUrl: string;
	    activeConnectionId: string;
	    activeConnectionName: string;
	    totalRequests: number;
	    blockedRequests: number;
	
	    static createFrom(source: any = {}) {
	        return new MCPServerStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.port = source["port"];
	        this.sseUrl = source["sseUrl"];
	        this.httpUrl = source["httpUrl"];
	        this.activeConnectionId = source["activeConnectionId"];
	        this.activeConnectionName = source["activeConnectionName"];
	        this.totalRequests = source["totalRequests"];
	        this.blockedRequests = source["blockedRequests"];
	    }
	}
	export class PLSQLResult {
	    success: boolean;
	    output: string[];
	    rowsAffected: number;
	    executionMs: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new PLSQLResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.output = source["output"];
	        this.rowsAffected = source["rowsAffected"];
	        this.executionMs = source["executionMs"];
	        this.error = source["error"];
	    }
	}
	export class QueryResult {
	    columns: string[];
	    rows: any[];
	    rowCount: number;
	    executionMs: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.rowCount = source["rowCount"];
	        this.executionMs = source["executionMs"];
	        this.error = source["error"];
	    }
	}
	export class SchemaInfo {
	    tables: string[];
	    views: string[];
	    procedures: string[];
	
	    static createFrom(source: any = {}) {
	        return new SchemaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tables = source["tables"];
	        this.views = source["views"];
	        this.procedures = source["procedures"];
	    }
	}
	export class SecurityPolicy {
	    mode: string;
	    blockedKeywords: string[];
	    allowPlsql: boolean;
	    maxRows: number;
	
	    static createFrom(source: any = {}) {
	        return new SecurityPolicy(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	        this.blockedKeywords = source["blockedKeywords"];
	        this.allowPlsql = source["allowPlsql"];
	        this.maxRows = source["maxRows"];
	    }
	}
	export class TNSEntry {
	    alias: string;
	    host: string;
	    port: number;
	    protocol: string;
	    serviceName: string;
	    sid: string;
	    isSid: boolean;
	    ssl: boolean;
	    raw: string;
	
	    static createFrom(source: any = {}) {
	        return new TNSEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.protocol = source["protocol"];
	        this.serviceName = source["serviceName"];
	        this.sid = source["sid"];
	        this.isSid = source["isSid"];
	        this.ssl = source["ssl"];
	        this.raw = source["raw"];
	    }
	}

}

