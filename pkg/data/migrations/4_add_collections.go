package migrations

func init() {
	migrateSQL := `
DROP TABLE IF EXISTS collection_bonds;

CREATE TABLE collection_bonds
(
    id            int  NOT NULL GENERATED BY DEFAULT AS IDENTITY CONSTRAINT pk_collection_bonds PRIMARY KEY,
    collection_id text NOT NULL,
    duration      int  NOT NULL,
    bond_id       int  NOT NULL CONSTRAINT "FK_collection_bonds_bond" REFERENCES bonds ON DELETE CASCADE,
    index         int  NOT NULL
);

CREATE UNIQUE INDEX ix_collection_bonds_uniq ON collection_bonds (collection_id, duration, bond_id);
CREATE INDEX ix_collection_bonds_collection_id ON collection_bonds (collection_id);
CREATE INDEX ix_collection_bonds_duration ON collection_bonds (duration);
CREATE INDEX ix_collection_bonds_lookup ON collection_bonds (collection_id, duration, index ASC);
`

	rollback := `
DROP TABLE IF EXISTS collection_bonds;
`

	registerSQL("4_add_collections", migrateSQL, rollback)
}
