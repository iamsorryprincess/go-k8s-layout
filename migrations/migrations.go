package migrations

import "embed"

const PostgresDir = "postgres"

//go:embed postgres/*.sql
var FS embed.FS
