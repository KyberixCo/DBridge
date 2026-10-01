package oracle

import (
	"testing"
)

func TestParseTNSContent(t *testing.T) {
	sampleTNS := `
# Sample tnsnames.ora
# Author: Oracle DBA

ORCLPDB1 =
  (DESCRIPTION =
    (ADDRESS = (PROTOCOL = TCP)(HOST = oracledb.internal)(PORT = 1521))
    (CONNECT_DATA =
      (SERVER = DEDICATED)
      (SERVICE_NAME = orclpdb1.localdomain)
    )
  )

# Legacy SID entry with multiple aliases
XE, XE.WORLD =
  (DESCRIPTION =
    (ADDRESS_LIST =
      (ADDRESS = (PROTOCOL = TCP)(HOST = 192.168.1.100)(PORT = 1522))
    )
    (CONNECT_DATA =
      (SID = XE)
    )
  )

# Secure cloud database with TCPS
ATP_CLOUD =
  (DESCRIPTION =
    (ADDRESS = (PROTOCOL = TCPS)(HOST = atp.adb.us-ashburn-1.oraclecloud.com)(PORT = 2484))
    (CONNECT_DATA =
      (SERVICE_NAME = db2024_high.adb.oraclecloud.com)
    )
  )
`

	entries, err := ParseTNSContent(sampleTNS)
	if err != nil {
		t.Fatalf("unexpected error parsing TNS: %v", err)
	}

	if len(entries) != 4 { // ORCLPDB1, XE, XE.WORLD, ATP_CLOUD
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	// 1. Check ORCLPDB1
	e0 := entries[0]
	if e0.Alias != "ORCLPDB1" || e0.Host != "oracledb.internal" || e0.Port != 1521 || e0.ServiceName != "orclpdb1.localdomain" || e0.IsSID {
		t.Errorf("e0 mismatch: %+v", e0)
	}

	// 2. Check XE (SID)
	e1 := entries[1]
	if e1.Alias != "XE" || e1.Host != "192.168.1.100" || e1.Port != 1522 || e1.SID != "XE" || !e1.IsSID {
		t.Errorf("e1 mismatch: %+v", e1)
	}

	// 3. Check XE.WORLD (secondary alias)
	e2 := entries[2]
	if e2.Alias != "XE.WORLD" || e2.Host != "192.168.1.100" || e2.Port != 1522 || e2.SID != "XE" || !e2.IsSID {
		t.Errorf("e2 mismatch: %+v", e2)
	}

	// 4. Check ATP_CLOUD (TCPS / SSL)
	e3 := entries[3]
	if e3.Alias != "ATP_CLOUD" || !e3.SSL || e3.Port != 2484 || e3.ServiceName != "db2024_high.adb.oraclecloud.com" {
		t.Errorf("e3 mismatch: %+v", e3)
	}
}
