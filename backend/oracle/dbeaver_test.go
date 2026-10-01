package oracle

import (
	"testing"
)

func TestParseDBeaverContent(t *testing.T) {
	dbeaverJSON := `{
		"connections": {
			"oracle-1": {
				"provider": "oracle",
				"driver": "oracle_thin",
				"name": "Production 19c PDB",
				"configuration": {
					"host": "db.corp.internal",
					"port": "1521",
					"database": "SALESPDB",
					"user": "sales_admin",
					"provider-properties": {
						"@dbeaver-sid-service@": "SERVICE"
					}
				}
			},
			"oracle-2": {
				"provider": "oracle",
				"driver": "oracle_thin",
				"name": "Legacy 11g SID",
				"configuration": {
					"host": "oldhost.local",
					"port": "1521",
					"database": "ORCL",
					"user": "system",
					"provider-properties": {
						"@dbeaver-sid-service@": "SID"
					}
				}
			},
			"postgres-ignored": {
				"provider": "postgresql",
				"driver": "postgres-jdbc",
				"name": "App DB Postgres",
				"configuration": {
					"host": "localhost",
					"port": "5432"
				}
			}
		}
	}`

	profiles, err := ParseDBeaverContent([]byte(dbeaverJSON))
	if err != nil {
		t.Fatalf("unexpected error parsing DBeaver content: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 Oracle profiles, got %d", len(profiles))
	}

	// Verify one is SERVICE and one is SID
	var foundService, foundSID bool
	for _, p := range profiles {
		if p.Name == "Production 19c PDB" {
			foundService = true
			if p.IsSID || p.ServiceName != "SALESPDB" || p.Username != "sales_admin" || p.Host != "db.corp.internal" {
				t.Errorf("Production profile mismatch: %+v", p)
			}
		}
		if p.Name == "Legacy 11g SID" {
			foundSID = true
			if !p.IsSID || p.SID != "ORCL" || p.Username != "system" || p.Host != "oldhost.local" {
				t.Errorf("Legacy profile mismatch: %+v", p)
			}
		}
	}

	if !foundService || !foundSID {
		t.Errorf("Did not find both expected profiles")
	}
}
