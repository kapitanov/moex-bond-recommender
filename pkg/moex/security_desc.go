package moex

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// SecurityDescription содержит набор параметров ценной бумаги
type SecurityDescription struct {
	Properties map[PropertyID]*Property
}

// IsForQualifiedInvestorsOnly возвращает значение параметра IsForQualifiedInvestorsOnlyProperty
func (desc SecurityDescription) IsForQualifiedInvestorsOnly() (bool, error) {
	prop, exists := desc.Properties[IsForQualifiedInvestorsOnlyProperty]
	if !exists {
		return false, nil
	}

	value, err := prop.AsBool()
	return value, err
}

// InitialFaceValue возвращает значение параметра InitialFaceValueProperty
func (desc SecurityDescription) InitialFaceValue() (*float64, error) {
	prop, exists := desc.Properties[InitialFaceValueProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsFloat64()
	return &value, err
}

// FaceUnit возвращает значение параметра FaceUnitProperty
func (desc SecurityDescription) FaceUnit() (*string, error) {
	prop, exists := desc.Properties[FaceUnitProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsString()
	return &value, err
}

// IssueDate возвращает значение параметра IssueDateProperty
func (desc SecurityDescription) IssueDate() (*Date, error) {
	prop, exists := desc.Properties[IssueDateProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsDate()
	return &value, err
}

// MaturityDate возвращает значение параметра MaturityDateProperty
func (desc SecurityDescription) MaturityDate() (*Date, error) {
	prop, exists := desc.Properties[MaturityDateProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsDate()
	return &value, err
}

// ListingLevel возвращает значение параметра ListingLevelProperty
func (desc SecurityDescription) ListingLevel() (*float64, error) {
	prop, exists := desc.Properties[ListingLevelProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsFloat64()
	return &value, err
}

// CouponFrequency возвращает значение параметра CouponFrequencyProperty
func (desc SecurityDescription) CouponFrequency() (*float64, error) {
	prop, exists := desc.Properties[CouponFrequencyProperty]
	if !exists {
		return nil, nil
	}

	value, err := prop.AsFloat64()
	return &value, err
}

// IsHighRisk возвращает значение параметра IsHighRiskProperty
func (desc SecurityDescription) IsHighRisk() (bool, error) {
	prop, exists := desc.Properties[IsHighRiskProperty]
	if !exists {
		return false, nil
	}

	value, err := prop.AsBool()
	return value, err
}

// PropertyID содержит тип параметра ценной бумаги
type PropertyID string

const (
	IssueDateProperty                   PropertyID = "ISSUEDATE"
	MaturityDateProperty                PropertyID = "MATDATE"
	InitialFaceValueProperty            PropertyID = "INITIALFACEVALUE"
	FaceUnitProperty                    PropertyID = "FACEUNIT"
	ListingLevelProperty                PropertyID = "LISTLEVEL"
	IsForQualifiedInvestorsOnlyProperty PropertyID = "ISQUALIFIEDINVESTORS"
	CouponFrequencyProperty             PropertyID = "COUPONFREQUENCY"
	IsHighRiskProperty                  PropertyID = "HIGHRISK"
)

// PropertyType содержит тип значения параметра ценной бумаги
type PropertyType string

const (
	StringPropertyType  PropertyType = "string"
	DatePropertyType    PropertyType = "date"
	NumberPropertyType  PropertyType = "number"
	BooleanPropertyType PropertyType = "boolean"
)

var ErrWrongPropertyType = errors.New("wrong property type")

// Property содержит отдельный параметр ценной бумаги
type Property struct {
	Name  PropertyID   `json:"name"`
	Value string       `json:"value"`
	Type  PropertyType `json:"type"`
}

// AsString возвращает значение для свойств типа "string"
func (p Property) AsString() (string, error) {
	if p.Type != StringPropertyType {
		return "", ErrWrongPropertyType
	}

	return p.Value, nil
}

// AsDate возвращает значение для свойств типа "date"
func (p Property) AsDate() (Date, error) {
	if p.Type != DatePropertyType {
		return Date{}, ErrWrongPropertyType
	}

	return NewDate(p.Value)
}

// AsFloat64 возвращает значение для свойств типа "number"
func (p Property) AsFloat64() (float64, error) {
	if p.Type != NumberPropertyType {
		return float64(0), ErrWrongPropertyType
	}

	return strconv.ParseFloat(p.Value, 64)
}

// AsBool возвращает значение для свойств типа "boolean"
func (p Property) AsBool() (bool, error) {
	if p.Type != BooleanPropertyType {
		return false, ErrWrongPropertyType
	}

	value, err := strconv.Atoi(p.Value)
	if err != nil {
		return false, err
	}

	return value != 0, nil
}

// GetSecurityDescription возвращает описание ценной бумаги
func (p *provider) GetSecurityDescription(ctx context.Context, isin string) (*SecurityDescription, error) {
	values := make(url.Values)

	values.Set("iss.only", "description")
	values.Set("iss.json", "extended")
	values.Set("iss.meta", "off")

	u := fmt.Sprintf("/iss/securities/%s.json?%s", url.PathEscape(isin), values.Encode())

	var resp []descriptionResponse
	err := p.getJSON(ctx, u, &resp)
	if err != nil {
		return nil, err
	}

	desc := SecurityDescription{
		Properties: make(map[PropertyID]*Property),
	}

	for _, respItem := range resp {
		for _, prop := range respItem.Properties {
			desc.Properties[prop.Name] = prop
		}
	}

	return &desc, nil
}

type descriptionResponse struct {
	Properties []*Property `json:"description"`
}
