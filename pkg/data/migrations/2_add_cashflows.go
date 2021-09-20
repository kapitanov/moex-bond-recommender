package migrations

func init() {
	migrateSQL := `
DROP MATERIALIZED VIEW IF EXISTS cashflows;

CREATE MATERIALIZED VIEW cashflows AS
SELECT bond_id,
       date,
       type,
       value_rub
FROM payments
WHERE bond_id IN (
    SELECT id
    FROM bonds
    WHERE face_unit = 'RUB' AND is_traded = true
)
  AND date > NOW()::date
  AND value > 0
ORDER BY date, bond_id;

CREATE INDEX ix_cashflows_bond_id ON cashflows (bond_id);
CREATE UNIQUE INDEX ix_cashflows_unique ON cashflows (bond_id, date, type);
`

	rollback := `
DROP MATERIALIZED VIEW IF EXISTS cashflows;
`

	registerSQL("2_add_cashflows", migrateSQL, rollback)
}
