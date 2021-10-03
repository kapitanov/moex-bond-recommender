package pages

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/goodsign/monday"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/recommender"
)

const (
	euroSymbol = "\u20ac"
	rubSymbol  = "\u20bd"
)

// DefineFunctions создает набор функций для отрисовки представлений
func DefineFunctions(googleAnalyticsID string) template.FuncMap {
	fns := make(template.FuncMap)
	fns["formatDate"] = formatDate
	fns["formatPercent"] = formatPercent
	fns["formatPercentNoScale"] = formatPercentNoScale
	fns["formatMoney"] = formatMoney
	fns["formatDuration"] = formatDuration
	fns["formatDaysTillMaturity"] = formatDaysTillMaturity
	fns["formatCashFlowItemType"] = formatCashFlowItemType
	fns["getFullOpenValue"] = getFullOpenValue
	fns["getFullRevenue"] = getFullRevenue
	fns["formatBool"] = formatBool
	fns["formatBondType"] = formatBondType
	fns["formatPercentWithSign"] = formatPercentWithSign
	fns["formatMoneyWithSign"] = formatMoneyWithSign
	fns["json"] = formatJSON

	fns["googleAnalyticsID"] = func() (string, error) {
		return googleAnalyticsID, nil
	}

	return fns
}

func formatDate(v interface{}) (template.HTML, error) {
	str := ""
	switch t := v.(type) {
	case time.Time:
		str = monday.Format(t, "2 Jan 2006", monday.LocaleRuRU)
		break
	case *time.Time:
		if t != nil {
			str = monday.Format(*t, "2 Jan 2006", monday.LocaleRuRU)
		}
		break
	case sql.NullTime:
		if t.Valid {
			str = monday.Format(t.Time, "2 Jan 2006", monday.LocaleRuRU)
		}
		break
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatPercent(v interface{}) (template.HTML, error) {
	str := ""

	switch t := v.(type) {
	case float64:
		str = fmt.Sprintf("%0.2f%%", t)
		break
	case *float64:
		if t != nil {
			str = fmt.Sprintf("%0.2f%%", *t)
		}
		break
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatPercentNoScale(v interface{}) (template.HTML, error) {
	switch t := v.(type) {
	case float64:
		return formatPercent(t * 100.0)
	case *float64:
		if t != nil {
			return formatPercent((*t) * 100.0)
		}
		break
	}

	return "", nil
}

func formatMoney(currency string, v interface{}) (template.HTML, error) {
	str := ""
	switch t := v.(type) {
	case float64:
		str = fmt.Sprintf("%0.2f", t)
		break
	case *float64:
		if t != nil {
			str = fmt.Sprintf("%0.2f", *t)
		}
		break
	}

	if str != "" {
		switch currency {
		case "USD":
			str = fmt.Sprintf("$%s", str)
			break
		case "EUR":
			str = fmt.Sprintf("%s %s", str, euroSymbol)
			break
		case "RUB":
			str = fmt.Sprintf("%s %s", str, rubSymbol)
			break
		default:
			str = fmt.Sprintf("%s %s", str, currency)
			break
		}
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatDuration(v interface{}) (template.HTML, error) {
	str := ""
	switch t := v.(type) {
	case recommender.Duration:
		switch t {
		case recommender.Duration1Year:
			str = "1 год"
			break
		case recommender.Duration2Year:
			str = "2 года"
			break
		case recommender.Duration3Year:
			str = "3 года"
			break
		case recommender.Duration4Year:
			str = "4 года"
			break
		case recommender.Duration5Year:
			str = "5 лет"
			break
		default:
			str = string(t)
			break
		}
		break
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatDaysTillMaturity(v interface{}) (template.HTML, error) {
	var d time.Time
	switch t := v.(type) {
	case time.Time:
		d = t
		break
	case *time.Time:
		if t == nil {
			return "", nil
		}
		d = *t
		break
	case sql.NullTime:
		if !t.Valid {
			return "", nil
		}
		d = t.Time
		break
	default:
		return "", nil
	}

	days := int(math.Round(d.Sub(time.Now().UTC()).Hours() / 24.0))

	str := fmt.Sprintf("%d", days)
	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatCashFlowItemType(v interface{}) (template.HTML, error) {
	str := ""

	switch t := v.(type) {
	case recommender.CashFlowItemType:
		switch t {
		case recommender.Coupon:
			str = "Купон"
			break
		case recommender.Amortization:
			str = "Амортизация"
			break
		case recommender.Maturity:
			str = "Погашение"
			break
		}
		break
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func getFullOpenValue(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case *recommender.Report:
		return t.OpenValue + t.OpenFee, nil
	}

	return v, nil
}

func getFullRevenue(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case *recommender.Report:
		return t.Revenue - t.Taxes, nil
	}

	return v, nil
}

func formatBool(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case bool:
		if t {
			return "Да", nil
		}
		return "Нет", nil
	}

	return v, nil
}

func formatBondType(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case data.BondType:
		switch t {
		case data.SubfederalBond:
			return "Субфедеральная облигация", nil
		case data.OFZBond:
			return "Облигация федерального займа", nil
		case data.ExchangeBond:
			return "Биржевая облигация", nil
		case data.CBBond:
			return "Облигация Центрального банка", nil
		case data.MunicipalBond:
			return "Мунициальная облигация", nil
		case data.CorporateBond:
			return "Корпоративная облигация", nil
		case data.IFIBond:
			return "Облигация международной финансовой организации", nil
		case data.EuroBond:
			return "Еврооблигация", nil
		default:
			return string(t), nil
		}
	}

	return v, nil
}

func formatPercentWithSign(v interface{}) (template.HTML, error) {
	str := ""

	switch t := v.(type) {
	case float64:
		sign := "+"
		if t < 0 {
			sign = "-"
		}
		str = fmt.Sprintf("%s %0.2f%%", sign, math.Abs(t))
		break
	case *float64:
		if t != nil {
			sign := "+"
			if *t < 0 {
				sign = "-"
			}
			str = fmt.Sprintf("%s %0.2f%%", sign, math.Abs(*t))
		}
		break
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatMoneyWithSign(currency string, v interface{}) (template.HTML, error) {
	str := ""
	switch t := v.(type) {
	case float64:
		sign := "+"
		if t < 0 {
			sign = "-"
		}
		str = fmt.Sprintf("%s %0.2f", sign, math.Abs(t))
		break
	case *float64:
		if t != nil {
			sign := "+"
			if *t < 0 {
				sign = "-"
			}
			str = fmt.Sprintf("%s %0.2f", sign, math.Abs(*t))
		}
		break
	}

	if str != "" {
		switch currency {
		case "USD":
			str = fmt.Sprintf("$%s", str)
			break
		case "EUR":
			str = fmt.Sprintf("%s %s", str, euroSymbol)
			break
		case "RUB":
			str = fmt.Sprintf("%s %s", str, rubSymbol)
			break
		default:
			str = fmt.Sprintf("%s %s", str, currency)
			break
		}
	}

	str = template.HTMLEscapeString(str)
	return template.HTML(str), nil
}

func formatJSON(v interface{}) (interface{}, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
