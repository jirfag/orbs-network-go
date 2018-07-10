package publicapi

import (
	"github.com/orbs-network/orbs-network-go/ledger"
	"github.com/orbs-network/orbs-network-go/instrumentation"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/protocol/client"
)

type service struct {
	services.PublicApi
	transactionPool services.TransactionPool
	ledger          ledger.Ledger
	events          instrumentation.Reporting
	isLeader        bool
}

func NewPublicApi(
	transactionPool services.TransactionPool,
	ledger ledger.Ledger,
	events instrumentation.Reporting,
	isLeader bool,
) services.PublicApi {

	return &service{
		transactionPool: transactionPool,
		ledger:          ledger,
		events:          events,
		isLeader:        isLeader,
	}
}

func (s *service) SendTransaction(input *services.SendTransactionInput) (*services.SendTransactionOutput, error) {
	s.events.Info("enter_send_transaction")
	defer s.events.Info("exit_send_transaction")
	//TODO leader should also propagate transactions to other nodes
	tx := input.ClientRequest.SignedTransaction()
	s.transactionPool.AddNewTransaction(&services.AddNewTransactionInput{tx})
	output := &services.SendTransactionOutput{}
	return output, nil
}

func (s *service) CallMethod(input *services.CallMethodInput) (*services.CallMethodOutput, error) {
	s.events.Info("enter_call_method")
	defer s.events.Info("exit_call_method")
	output := &services.CallMethodOutput{ClientResponse: (&client.CallMethodResponseBuilder{
		OutputArguments: []*protocol.MethodArgumentBuilder{
			{Name: "balance", Type: protocol.MethodArgumentTypeUint64, Uint64: uint64(s.ledger.GetState())},
		},
	}).Build()}
	return output, nil
}

func (s *service) GetTransactionStatus(input *services.GetTransactionStatusInput) (*services.GetTransactionStatusOutput, error) {
	panic("Not implemented")
}

func (s *service) HandleTransactionResults(input *handlers.HandleTransactionResultsInput) (*handlers.HandleTransactionResultsOutput, error) {
	panic("Not implemented")
}