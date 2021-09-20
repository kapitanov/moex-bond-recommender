package recommender

func init() {
	register("ofz", "ОФЗ", func(duration Duration) string {
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
SELECT bonds.id
FROM bonds
INNER JOIN issuers ON issuers.id = bonds.issuer_id
WHERE is_traded
  AND maturity_date IS NOT NULL
  AND type = 'ofz_bond'
  AND qualified_only = FALSE
  AND high_risk = FALSE
  AND listing_level = 1
  AND face_unit = 'RUB'
  AND issuers.inn IS NOT NULL
  AND starts_with(issuers.inn::text, '77')
`
		return text
	})
}
