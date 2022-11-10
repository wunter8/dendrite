package deltas

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/matrix-org/gomatrixserverlib"
)

var serverNamesTables = []string{
	"userapi_accounts",
	"userapi_account_datas",
	"userapi_devices",
	"userapi_notifications",
	"userapi_openid_tokens",
	"userapi_profiles",
	"userapi_pushers",
	"userapi_threepids",
}

// These tables have a PRIMARY KEY constraint which we need to drop so
// that we can recreate a new unique index that contains the server name.
// If the new key doesn't exist (i.e. the database was created before the
// table rename migration) we'll try to drop the old one instead.
var serverNamesDropPK = map[string]string{
	"userapi_accounts":      "account_accounts",
	"userapi_account_datas": "account_data",
	"userapi_profiles":      "account_profiles",
}

// I know what you're thinking: you're wondering "why doesn't this use $1
// and pass variadic parameters to ExecContext?" — the answer is because
// PostgreSQL doesn't expect the table name to be specified as a substituted
// argument in that way so it results in a syntax error in the query.

func UpServerNames(ctx context.Context, tx *sql.Tx, serverName gomatrixserverlib.ServerName) error {
	for _, table := range serverNamesTables {
		q := fmt.Sprintf(
			"ALTER TABLE IF EXISTS %s ADD COLUMN IF NOT EXISTS server_name TEXT NOT NULL DEFAULT '';",
			pq.QuoteIdentifier(table),
		)
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("add server name to %q error: %w", table, err)
		}
	}
	for newTable, oldTable := range serverNamesDropPK {
		q := fmt.Sprintf(
			"ALTER TABLE IF EXISTS %s DROP CONSTRAINT IF EXISTS %s;",
			pq.QuoteIdentifier(newTable), pq.QuoteIdentifier(newTable+"_pkey"),
		)
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("drop PK from %q error: %w", newTable, err)
		}
		q = fmt.Sprintf(
			"ALTER TABLE IF EXISTS %s DROP CONSTRAINT IF EXISTS %s;",
			pq.QuoteIdentifier(oldTable), pq.QuoteIdentifier(oldTable+"_pkey"),
		)
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("drop PK from %q error: %w", oldTable, err)
		}
	}
	return nil
}
