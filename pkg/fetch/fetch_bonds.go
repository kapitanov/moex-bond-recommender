package fetch

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

type bondFetchWorker struct {
	provider moex.Provider
	tx       *data.TX
	log      *log.Logger
	stats    *BondFetchStats
}

// FetchBonds выполняет выгрузку облигаций из ISS
func (w *bondFetchWorker) FetchBonds(ctx context.Context) error {
	tradingStatus := moex.IsTraded
	query := moex.SecurityListQuery{
		Engine:        moex.StockEngine,
		Market:        moex.BondMarket,
		TradingStatus: &tradingStatus,
	}
	iter := w.provider.ListSecurities(ctx, query)
	count := 0
	for {
		w.log.Printf("fetch bonds: %d item(s) processed", count)

		securities, err := iter.Next()
		if err != nil {
			if err == moex.EOF {
				break
			}
			return err
		}

		for _, security := range securities {
			err = ctx.Err()
			if err != nil {
				return err
			}

			issuer, err := w.CreateOrUpdateIssuer(security)
			if err != nil {
				return err
			}

			err = w.CreateOrUpdateBond(ctx, security, issuer)
			if err != nil {
				return err
			}

			count++
		}
	}

	return nil
}

// CreateOrUpdateBond выполняет выгрузку отдельно взятой облигации
func (w *bondFetchWorker) CreateOrUpdateBond(ctx context.Context, security *moex.Security, issuer *data.Issuer) error {
	_, err := w.tx.Bonds.GetByMoexID(security.ID)
	if err == nil {
		return nil
	}
	if err != data.ErrNotFound {
		return err
	}

	_, err = w.tx.Bonds.GetBySecurityID(security.SecurityID)
	if err == nil {
		return nil
	}
	if err != data.ErrNotFound {
		return err
	}

	props, err := w.GetSecurityProps(ctx, security)
	if err != nil {
		return err
	}

	args := data.CreateBondArgs{
		IssuerID:           issuer.ID,
		MoexID:             security.ID,
		SecurityID:         security.SecurityID,
		ShortName:          security.ShortName,
		FullName:           security.Name,
		ISIN:               security.ISIN,
		IsTraded:           security.IsTraded == moex.IsTraded,
		QualifiedOnly:      props.QualifiedOnly,
		IsHighRisk:         props.IsHighRisk,
		Type:               data.BondType(security.Type),
		PrimaryBoardID:     security.PrimaryBoardID,
		MarketPriceBoardID: security.MarketPriceBoardID,
		InitialFaceValue:   props.InitialFaceValue,
		FaceUnit:           normalizeCurrency(props.FaceUnit),
		IssueDate:          props.IssueDate,
		MaturityDate:       props.MaturityDate,
		ListingLevel:       props.ListingLevel,
		CouponFrequency:    props.CouponFrequency,
	}
	b, err := w.tx.Bonds.Create(args)
	if err != nil {
		return err
	}

	w.log.Printf("new bond: #%d %s \"%s\"", b.MoexID, b.ISIN, b.ShortName)
	w.stats.NewBonds++
	return nil
}

// GetSecurityProps запрашивает параметры облигации из провайдера
func (w *bondFetchWorker) GetSecurityProps(ctx context.Context, security *moex.Security) (*securityProps, error) {
	desc, err := w.provider.GetSecurityDescription(ctx, security.SecurityID)
	if err != nil {
		return nil, err
	}

	props := securityProps{}

	// QualifiedOnly
	props.QualifiedOnly, err = desc.IsForQualifiedInvestorsOnly()
	if err != nil {
		return nil, err
	}

	// IsHighRisk
	props.IsHighRisk, err = desc.IsHighRisk()
	if err != nil {
		return nil, err
	}

	// InitialFaceValue
	initialFaceValue, err := desc.InitialFaceValue()
	if err != nil {
		return nil, err
	}
	if initialFaceValue == nil {
		return nil, fmt.Errorf("missing property %s for %s", moex.InitialFaceValueProperty, security.ISIN)
	}
	props.InitialFaceValue = *initialFaceValue

	// FaceUnit
	faceUnit, err := desc.FaceUnit()
	if err != nil {
		return nil, err
	}
	if faceUnit == nil {
		return nil, fmt.Errorf("missing property %s for %s", moex.FaceUnitProperty, security.ISIN)
	}
	props.FaceUnit = *faceUnit

	// IssueDate
	issueDate, err := desc.IssueDate()
	if err != nil {
		return nil, err
	}
	props.IssueDate = dateToNullTime(issueDate)

	maturityDate, err := desc.MaturityDate()
	if err != nil {
		return nil, err
	}
	props.MaturityDate = dateToNullTime(maturityDate)

	listingLevel, err := desc.ListingLevel()
	if err != nil {
		return nil, err
	}
	if listingLevel == nil {
		return nil, fmt.Errorf("missing property %s for %s", moex.ListingLevelProperty, security.ISIN)
	}
	props.ListingLevel = int(*listingLevel)

	couponFrequency, err := desc.CouponFrequency()
	if err != nil {
		return nil, err
	}
	if couponFrequency == nil {
		props.CouponFrequency = 0
	} else {
		props.CouponFrequency = int(*couponFrequency)
	}

	return &props, nil
}

type securityProps struct {
	QualifiedOnly    bool
	IsHighRisk       bool
	InitialFaceValue float64
	FaceUnit         string
	IssueDate        sql.NullTime
	MaturityDate     sql.NullTime
	ListingLevel     int
	CouponFrequency  int
}
