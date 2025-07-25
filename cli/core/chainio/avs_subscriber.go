package chainio

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"

	"github.com/imua-xyz/imua-avs-sdk/logging"
	avssub "github.com/trigg3rX/imua-contracts/bindings/contracts/TriggerXAvs"
	"github.com/trigg3rX/triggerx-backend-imua/cli/core/chainio/eth"
)

type AvsRegistrySubscriber interface {
	SubscribeToNewTasks(newTaskCreatedChan chan *avssub.TriggerXAvsTaskCreated) event.Subscription
}

type AvsRegistryChainSubscriber struct {
	logger logging.Logger
	avssub avssub.TriggerXAvs
}

// forces EthSubscriber to implement the chainio.Subscriber interface
var _ AvsRegistrySubscriber = (*AvsRegistryChainSubscriber)(nil)

func NewAvsRegistryChainSubscriber(
	avssub avssub.TriggerXAvs,
	logger logging.Logger,
) (*AvsRegistryChainSubscriber, error) {
	return &AvsRegistryChainSubscriber{
		logger: logger,
		avssub: avssub,
	}, nil
}

func BuildAvsRegistryChainSubscriber(
	avssubAddr common.Address,
	ethWsClient eth.EthClient,
	logger logging.Logger,
) (*AvsRegistryChainSubscriber, error) {
	avssub, err := avssub.NewTriggerXAvs(avssubAddr, ethWsClient)
	if err != nil {
		logger.Error("Failed to create BLSApkRegistry contract", "err", err)
		return nil, err
	}
	return NewAvsRegistryChainSubscriber(*avssub, logger)
}

func (s *AvsRegistryChainSubscriber) SubscribeToNewTasks(newTaskCreatedChan chan *avssub.TriggerXAvsTaskCreated) event.Subscription {
	sub, err := s.avssub.WatchTaskCreated(
		&bind.WatchOpts{}, newTaskCreatedChan, []uint64{},
	)
	if err != nil {
		s.logger.Error("Failed to subscribe to new  tasks", "err", err)
	}
	s.logger.Infof("Subscribed to new TaskManager tasks")
	return sub
}
