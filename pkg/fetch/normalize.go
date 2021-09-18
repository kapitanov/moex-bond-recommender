package fetch

import (
	"database/sql"
	"time"

	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

func normalizeCurrency(str string) string {
	switch str {
	case "SUR":
	case "RUR":
		return "RUB"
	}

	return str
}

func timeToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: *t, Valid: true}
}

func nullableDateToNullTime(d moex.NullableDate) sql.NullTime {
	if !d.HasValue() {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: *d.Time(), Valid: true}
}

func dateToNullTime(d *moex.Date) sql.NullTime {
	if d == nil {
		return sql.NullTime{Valid: false}
	}

	return sql.NullTime{Time: d.Time(), Valid: true}
}
