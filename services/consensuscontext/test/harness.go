package test

import (
	"context"
	"github.com/orbs-network/go-mock"
	"github.com/orbs-network/orbs-network-go/config"
	"github.com/orbs-network/orbs-network-go/crypto/hash"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-network-go/instrumentation/metric"
	"github.com/orbs-network/orbs-network-go/services/consensuscontext"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type harness struct {
	transactionPool *services.MockTransactionPool
	reporting       log.BasicLogger
	service         services.ConsensusContext
	config          config.ConsensusContextConfig
}

func (h *harness) requestTransactionsBlock(ctx context.Context) (*protocol.TransactionsBlockContainer, error) {
	output, err := h.service.RequestNewTransactionsBlock(ctx, &services.RequestNewTransactionsBlockInput{
		BlockHeight:             1,
		MaxBlockSizeKb:          0,
		MaxNumberOfTransactions: 0,
		PrevBlockHash:           hash.CalcSha256([]byte{1}),
	})
	if err != nil {
		return nil, err
	}
	return output.TransactionsBlock, nil
}

func (h *harness) expectTransactionsRequestedFromTransactionPool(numTransactionsToReturn uint32) {

	output := &services.GetTransactionsForOrderingOutput{
		SignedTransactions: nil,
	}

	for i := uint32(0); i < numTransactionsToReturn; i++ {
		targetAddress := builders.AddressForEd25519SignerForTests(2)
		output.SignedTransactions = append(output.SignedTransactions, builders.TransferTransaction().WithAmountAndTargetAddress(uint64(i+1)*10, targetAddress).Build())
	}

	h.transactionPool.When("GetTransactionsForOrdering", mock.Any, mock.Any).Return(output, nil).Times(1)
}

func (h *harness) expectTransactionsNoLongerRequestedFromTransactionPool() {
	h.transactionPool.When("GetTransactionsForOrdering", mock.Any, mock.Any).Return(nil, nil).Times(0)
}

func (h *harness) verifyTransactionsRequestedFromTransactionPool(t *testing.T) {
	ok, _ := h.transactionPool.Verify()

	// TODO: How to print err if it's sometimes nil
	require.True(t, ok)
}

func newHarness() *harness {
	log := log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))

	transactionPool := &services.MockTransactionPool{}
	cfg := config.ForConsensusContextTests()

	metricFactory := metric.NewRegistry()

	service := consensuscontext.NewConsensusContext(transactionPool, nil, nil,
		cfg, log, metricFactory)

	return &harness{
		transactionPool: transactionPool,
		reporting:       log,
		service:         service,
		config:          cfg,
	}
}
