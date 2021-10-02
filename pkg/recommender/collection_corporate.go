package recommender

func init() {
	register("corporate", "Корпоративные облигации", func(duration Duration) string {
		// Выбираются все облигации по набору условий:
		// - торгующиеся
		// - погашение не позднее N лет с текущего момента
		// - тип - "corporate_bond"
		// - нет признака "только для квалифицированных инвесторов"
		// - нет признакак "высокий риск"
		// - валюта номинала - рубль
		// - приведенная доходность больше нуля и согласуется с критерием "три сигмы"
		text := `
SELECT id
FROM (
         SELECT bonds.id,
                r.interest_rate,
                AVG(r.interest_rate) OVER ()    AS mean,
                STDDEV(r.interest_rate) OVER () AS stddev
         FROM bonds
                  INNER JOIN issuers
                             ON issuers.id = bonds.issuer_id
                  INNER JOIN reports r ON bonds.id = r.bond_id

         WHERE is_traded
           AND maturity_date IS NOT NULL
           AND type = 'corporate_bond'
           AND qualified_only = FALSE
           AND high_risk = FALSE
           AND face_unit = 'RUB'
           AND r.interest_rate > 0
     ) xs
WHERE interest_rate <= (mean + 3 * stddev)
`
		return text
	})
}
