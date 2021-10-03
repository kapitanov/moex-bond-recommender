package recommender

func init() {
	register("highrisk", "Высокорисковые облигации", func(duration Duration) string {
		// Выбираются все облигации по набору условий:
		// - торгующиеся
		// - погашение не позднее N лет с текущего момента
		// - тип - "ofz_bond"
		// - нет признака "только для квалифицированных инвесторов"
		// - нет признакак "высокий риск"
		// - уровень листинга 1
		// - валюта номинала - рубль
		// - ИНН эмитента начинается с 77 (чтобы отфильтровать облигации других стран)
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
           AND qualified_only = FALSE
           AND high_risk = TRUE
           AND face_unit = 'RUB'
           AND r.interest_rate > 0
     ) xs
WHERE interest_rate <= (mean + 3 * stddev)
ORDER BY interest_rate DESC
`
		return text
	})
}
