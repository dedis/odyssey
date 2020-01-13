package catalogc

import (
	"go.dedis.ch/cothority/v3/byzcoin"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
)

// This service is used because we need to register our contracts to the ByzCoin
// service. So we create this stub and add contracts to it.

func init() {
	_, err := onet.RegisterNewService("OdysseyCatalogContract", newService)
	log.ErrFatal(err)
	byzcoin.RegisterGlobalContract(ContractCatalogID, contractCatalogFromBytes)
}

// Service is only used to being able to store our contracts
type Service struct {
	// We need to embed the ServiceProcessor, so that incoming messages
	// are correctly handled.
	*onet.ServiceProcessor
}

func newService(c *onet.Context) (onet.Service, error) {
	s := &Service{
		ServiceProcessor: onet.NewServiceProcessor(c),
	}
	return s, nil
}
