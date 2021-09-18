package fetch

import (
	"github.com/kapitanov/moex-bond-recommender/pkg/data"
	"github.com/kapitanov/moex-bond-recommender/pkg/moex"
)

// CreateOrUpdateIssuer выполняет синхронизацию отдельно взятого эмитента
func (w *bondFetchWorker) CreateOrUpdateIssuer(security *moex.Security) (*data.Issuer, error) {
	issuer, err := w.tx.Issuers.GetByMoexID(security.IssuerId)
	if err == nil {
		return issuer, nil
	}
	if err != data.ErrNotFound {
		return nil, err
	}

	args := data.CreateIssuerArgs{
		MoexID: security.IssuerId,
		Name:   security.IssuerName,
		INN:    security.IssuerINN,
		OKPO:   security.IssuerOKPO,
	}

	issuer, err = w.tx.Issuers.Create(args)
	if err != nil {
		return nil, err
	}

	w.stats.NewIssuers++
	w.log.Printf("new issuer #%d \"%s\"", issuer.MoexID, issuer.Name)
	return issuer, nil
}
