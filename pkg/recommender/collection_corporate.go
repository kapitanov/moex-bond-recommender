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
		text := `
SELECT bonds.id
FROM bonds
INNER JOIN issuers ON issuers.id = bonds.issuer_id
WHERE is_traded
  AND maturity_date IS NOT NULL
  AND type = 'corporate_bond'
  AND qualified_only = FALSE
  AND high_risk = FALSE
  AND face_unit = 'RUB'
`
		return text
	})
}
