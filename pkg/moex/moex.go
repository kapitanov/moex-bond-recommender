package moex

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultURL содержит корневой URL сервиса ISS по умолчанию
const DefaultURL = "https://iss.moex.com"

var (
	// EOF обозначает конец выгрузки
	EOF = errors.New("end of data")
)

// Provider является точкой входа в адаптер для ISS
type Provider interface {
	// ListSecurities возвращает итератор на список ценных бумаг
	ListSecurities(query SecurityListQuery) SecurityListIterator

	// ListCoupons возвращает итератор на список купонов
	ListCoupons(query CouponListQuery) CouponListIterator

	// ListAmortizations возвращает итератор на список амортизаций
	ListAmortizations(query AmortizationListQuery) AmortizationListIterator

	// ListOffers возвращает итератор на список оферт
	ListOffers(query OfferListQuery) OfferListIterator

	// GetMarketData возвращает текущие рыночные данные
	GetMarketData() ([]*MarketData, error)

	// GetSecurityDescription возвращает описание ценной бумаги
	GetSecurityDescription(isin string) (*SecurityDescription, error)
}

// Option конфирурирует провайдера
type Option func(p *provider) error

// WithURL задает корневой URL сервиса ISS
// По умолчанию используется DefaultURL
func WithURL(rootURL string) Option {
	rootURL = strings.TrimRight(rootURL, "/")

	return func(opts *provider) error {
		u, err := url.Parse(rootURL)
		if err != nil {
			return fmt.Errorf("\"%s\" is not a valid URL: %s", rootURL, err)
		}

		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("\"%s\" is not a valid URL: only http(s) URLs are supported", rootURL)
		}

		opts.BaseURL = rootURL
		return nil
	}
}

// WithLogger задает логгер для Provider
// По умолчанию логирование отключено
func WithLogger(logger *log.Logger) Option {
	return func(opts *provider) error {
		if logger == nil {
			return fmt.Errorf("logger option is nil")
		}

		opts.Logger = logger
		return nil
	}
}

// WithVerbose включает подробное логирование для Provider
// По умолчанию подробное логирование отключено
func WithVerbose(verbose bool) Option {
	return func(opts *provider) error {
		opts.Verbose = verbose
		return nil
	}
}

// WithHTTPClient задает экземпляр http.Client для Provider
// По умолчанию используется http.DefaultClient
func WithHTTPClient(httpClient *http.Client) Option {
	return func(opts *provider) error {
		if httpClient == nil {
			return fmt.Errorf("httpClient option is nil")
		}

		opts.HTTPClient = httpClient
		return nil
	}
}

// NewProvider создает новый экземпляр интерфейса Provider
func NewProvider(options ...Option) (Provider, error) {
	p := &provider{
		HTTPClient: http.DefaultClient,
		BaseURL:    DefaultURL,
		Logger:     log.New(io.Discard, "", 0),
		Verbose:    false,
	}
	for _, f := range options {
		err := f(p)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

type provider struct {
	HTTPClient *http.Client
	BaseURL    string
	Logger     *log.Logger
	Verbose    bool
}

func (p *provider) getJSON(url string, v interface{}) error {
	url = fmt.Sprintf("%s%s", p.BaseURL, url)
	resp, err := p.HTTPClient.Get(url)
	if err != nil {
		p.Logger.Printf("GET %s: %s", url, err)
		return err
	}

	if p.Verbose {
		p.Logger.Printf("GET %s -> %d", url, resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.Logger.Printf("GET %s: %s", url, err)
		return err
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		p.Logger.Printf("GET %s: %s", url, err)
		return err
	}

	return nil
}

// Date представляет дату в JSON формате
type Date struct {
	time time.Time
}

// NewDate разбирает значение типа Date из строки
func NewDate(str string) (Date, error) {
	time, err := time.Parse("2006-01-02", str)
	if err != nil {
		return Date{}, err
	}

	return Date{time}, nil
}

// Time возвращает дату и время
func (d Date) Time() time.Time {
	return d.time
}

// Format возвращает строковое представление для значения
func (d Date) Format(layout string) string {
	return d.time.Format(layout)
}

// String возвращает строковое представление для значения
func (d Date) String() string {
	return d.Format("2006-01-02")
}

// UnmarshalJSON выполняет десериализацию из JSON
func (d *Date) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	res, err := NewDate(str)
	if err != nil {
		return err
	}

	d.time = res.time
	return nil
}

// NullableDate представляет дату в JSON формате, для которой допускается значение nil
type NullableDate struct {
	time *time.Time
}

// Time возвращает дату и время
func (d NullableDate) Time() *time.Time {
	return d.time
}

// Format возвращает строковое представление для значения
func (d NullableDate) Format(layout string) string {
	if d.time == nil {
		return ""
	}

	return d.time.Format(layout)
}

// String возвращает строковое представление для значения
func (d NullableDate) String() string {
	return d.Format("2006-01-02")
}

// UnmarshalJSON выполняет десериализацию из JSON
func (d *NullableDate) UnmarshalJSON(data []byte) error {
	var str *string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	if str == nil || *str == "" || *str == "0000-00-00" {
		d.time = nil
		return nil
	}

	res, err := NewDate(*str)
	if err != nil {
		return err
	}

	d.time = &res.time
	return nil
}

// DateTime представляет дату и время в JSON формате
type DateTime struct {
	time time.Time
}

// Time возвращает дату и время
func (d DateTime) Time() time.Time {
	return d.time
}

// Format возвращает строковое представление для значения
func (d DateTime) Format(layout string) string {
	return d.time.Format(layout)
}

// String возвращает строковое представление для значения
func (d DateTime) String() string {
	return d.Format("2006-01-02 15:04:05")
}

// UnmarshalJSON выполняет десериализацию из JSON
func (d *DateTime) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	time, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return err
	}

	d.time = time
	return nil
}
