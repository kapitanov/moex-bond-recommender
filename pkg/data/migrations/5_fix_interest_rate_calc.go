package migrations

func init() {
	migrateSQL := `
DROP MATERIALIZED VIEW IF EXISTS reports;

CREATE MATERIALIZED VIEW reports AS
WITH cte_1 AS (
    SELECT bonds.maturity_date - NOW()::date                                               AS days_till_maturity,
           marketdata.currency                                                             AS currency,
           COALESCE(marketdata.last, marketdata.close_price, marketdata.legal_close_price) AS open_price,
           marketdata.accrued_interest                                                     AS open_accrued_interest,
           marketdata.face_value                                                           AS open_face_value,
           bonds.id                                                                        AS bond_id,
           bonds.issuer_id                                                                 AS bond_issuer_id,
           bonds.moex_id                                                                   AS bond_moex_id,
           bonds.security_id                                                               AS bond_security_id,
           bonds.short_name                                                                AS bond_short_name,
           bonds.full_name                                                                 AS bond_full_name,
           bonds.isin                                                                      AS bond_isin,
           bonds.is_traded                                                                 AS bond_is_traded,
           bonds.qualified_only                                                            AS bond_qualified_only,
           bonds.high_risk                                                                 AS bond_high_risk,
           bonds.type                                                                      AS bond_type,
           bonds.primary_board_id                                                          AS bond_primary_board_id,
           bonds.market_price_board_id                                                     AS bond_market_price_board_id,
           bonds.initial_face_value                                                        AS bond_initial_face_value,
           bonds.face_unit                                                                 AS bond_face_unit,
           bonds.issue_date                                                                AS bond_issue_date,
           bonds.maturity_date                                                             AS bond_maturity_date,
           bonds.listing_level                                                             AS bond_listing_level,
           bonds.coupon_freq                                                               AS bond_coupon_freq,
           bonds.created                                                                   AS bond_created,
           bonds.updated                                                                   AS bond_updated,
           issuers.id                                                                      AS issuer_id,
           issuers.moex_id                                                                 AS issuer_moex_id,
           issuers.name                                                                    AS issuer_name,
           issuers.inn                                                                     AS issuer_inn,
           issuers.okpo                                                                    AS issuer_okpo,
           issuers.created                                                                 AS issuer_created,
           issuers.updated                                                                 AS issuer_updated,
           marketdata.id                                                                   AS marketdata_id,
           marketdata.bond_id                                                              AS marketdata_bond_id,
           marketdata.time                                                                 AS marketdata_time,
           marketdata.face_value                                                           AS marketdata_face_value,
           marketdata.currency                                                             AS marketdata_currency,
           marketdata.last                                                                 AS marketdata_last,
           marketdata.last_change                                                          AS marketdata_last_change,
           marketdata.close_price                                                          AS marketdata_close_price,
           marketdata.legal_close_price                                                    AS marketdata_legal_close_price,
           marketdata.accrued_interest                                                     AS marketdata_accrued_interest
    FROM bonds
             INNER JOIN issuers ON issuers.id = bonds.issuer_id
             INNER JOIN marketdata ON bonds.id = marketdata.bond_id
    WHERE bonds.face_unit = 'RUB'
      AND bonds.is_traded = TRUE
      AND bonds.maturity_date > NOW()::date
      AND (marketdata.last IS NOT NULL OR marketdata.close_price IS NOT NULL OR
           marketdata.legal_close_price IS NOT NULL)
      AND marketdata.accrued_interest IS NOT NULL
      AND marketdata.face_value IS NOT NULL
      AND marketdata.currency = 'RUB'
),
     cte_2 AS (
         SELECT open_price * open_face_value / 100::numeric + open_accrued_interest AS open_value,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'C'),
                         0)                                                         AS coupon_payments,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'A'),
                         0)                                                         AS amortization_payments,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'M'),
                         0)                                                         AS maturity_payments,
                cte_1.*
         FROM cte_1
     ),
     cte_3 AS (
         SELECT ROUND(open_value * 0.0005, 2)                               AS open_fee,
                coupon_payments + amortization_payments + maturity_payments AS revenue,
                CASE
                    WHEN open_price < 100
                        THEN ROUND(
                                (coupon_payments + amortization_payments + open_face_value * (1 - open_price / 100)) *
                                0.13, 2)
                    ELSE
                        ROUND((coupon_payments + amortization_payments) * 0.13, 2)
                    END                                                     AS taxes,
                cte_2.*
         FROM cte_2
     )
SELECT ROUND(revenue - open_value - open_fee - taxes, 2)                                                        AS profit_loss,
       ROUND(100.0 * (revenue - open_value - open_fee - taxes) / open_value, 2)                                 AS relative_profit_loss,
       ROUND(100.0 * (revenue - open_value - open_fee - taxes) / open_value / (days_till_maturity / 356.25),
             2)                                                                                                 AS interest_rate,
       cte_3.*
FROM cte_3;

CREATE UNIQUE INDEX ix_reports_bond_id ON reports (bond_id);
CREATE INDEX ix_reports_interest_rate ON reports (interest_rate DESC);
CREATE INDEX ix_reports_bond_type ON reports (bond_type);
`

	rollback := `
DROP MATERIALIZED VIEW IF EXISTS reports;

CREATE MATERIALIZED VIEW reports AS
WITH cte_1 AS (
    SELECT bonds.maturity_date - NOW()::date                                               AS days_till_maturity,
           marketdata.currency                                                             AS currency,
           COALESCE(marketdata.last, marketdata.close_price, marketdata.legal_close_price) AS open_price,
           marketdata.accrued_interest                                                     AS open_accrued_interest,
           marketdata.face_value                                                           AS open_face_value,
           bonds.id                                                                        AS bond_id,
           bonds.issuer_id                                                                 AS bond_issuer_id,
           bonds.moex_id                                                                   AS bond_moex_id,
           bonds.security_id                                                               AS bond_security_id,
           bonds.short_name                                                                AS bond_short_name,
           bonds.full_name                                                                 AS bond_full_name,
           bonds.isin                                                                      AS bond_isin,
           bonds.is_traded                                                                 AS bond_is_traded,
           bonds.qualified_only                                                            AS bond_qualified_only,
           bonds.high_risk                                                                 AS bond_high_risk,
           bonds.type                                                                      AS bond_type,
           bonds.primary_board_id                                                          AS bond_primary_board_id,
           bonds.market_price_board_id                                                     AS bond_market_price_board_id,
           bonds.initial_face_value                                                        AS bond_initial_face_value,
           bonds.face_unit                                                                 AS bond_face_unit,
           bonds.issue_date                                                                AS bond_issue_date,
           bonds.maturity_date                                                             AS bond_maturity_date,
           bonds.listing_level                                                             AS bond_listing_level,
           bonds.coupon_freq                                                               AS bond_coupon_freq,
           bonds.created                                                                   AS bond_created,
           bonds.updated                                                                   AS bond_updated,
           issuers.id                                                                      AS issuer_id,
           issuers.moex_id                                                                 AS issuer_moex_id,
           issuers.name                                                                    AS issuer_name,
           issuers.inn                                                                     AS issuer_inn,
           issuers.okpo                                                                    AS issuer_okpo,
           issuers.created                                                                 AS issuer_created,
           issuers.updated                                                                 AS issuer_updated,
           marketdata.id                                                                   AS marketdata_id,
           marketdata.bond_id                                                              AS marketdata_bond_id,
           marketdata.time                                                                 AS marketdata_time,
           marketdata.face_value                                                           AS marketdata_face_value,
           marketdata.currency                                                             AS marketdata_currency,
           marketdata.last                                                                 AS marketdata_last,
           marketdata.last_change                                                          AS marketdata_last_change,
           marketdata.close_price                                                          AS marketdata_close_price,
           marketdata.legal_close_price                                                    AS marketdata_legal_close_price,
           marketdata.accrued_interest                                                     AS marketdata_accrued_interest
    FROM bonds
             INNER JOIN issuers ON issuers.id = bonds.issuer_id
             INNER JOIN marketdata ON bonds.id = marketdata.bond_id
    WHERE bonds.face_unit = 'RUB'
      AND bonds.is_traded = TRUE
      AND bonds.maturity_date > NOW()::date
      AND (marketdata.last IS NOT NULL OR marketdata.close_price IS NOT NULL OR
           marketdata.legal_close_price IS NOT NULL)
      AND marketdata.accrued_interest IS NOT NULL
      AND marketdata.face_value IS NOT NULL
      AND marketdata.currency = 'RUB'
),
     cte_2 AS (
         SELECT open_price * open_face_value / 100::numeric + open_accrued_interest AS open_value,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'C'),
                         0)                                                         AS coupon_payments,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'A'),
                         0)                                                         AS amortization_payments,
                COALESCE((SELECT SUM(value_rub) FROM cashflows WHERE bond_id = cte_1.bond_id AND type = 'M'),
                         0)                                                         AS maturity_payments,
                cte_1.*
         FROM cte_1
     ),
     cte_3 AS (
         SELECT ROUND(open_value * 0.0005, 2)                                                  AS open_fee,
                coupon_payments + amortization_payments + maturity_payments                    AS revenue,
                ROUND((coupon_payments + amortization_payments + maturity_payments) * 0.13, 2) AS taxes,
                cte_2.*
         FROM cte_2
     )
SELECT ROUND(revenue - open_value - open_fee, 2)                        AS profit_loss,
       ROUND(100.0 * (revenue - open_value - open_fee) / open_value, 2) AS relative_profit_loss,
       ROUND(100.0 * (revenue - open_value - open_fee) / open_value / (days_till_maturity / 356.25), 2) AS interest_rate,
       cte_3.*
FROM cte_3;

CREATE UNIQUE INDEX ix_reports_bond_id ON reports (bond_id);
CREATE INDEX ix_reports_interest_rate ON reports (interest_rate DESC);
CREATE INDEX ix_reports_bond_type ON reports (bond_type);
`

	registerSQL("5_fix_interest_rate_calc", migrateSQL, rollback)
}
