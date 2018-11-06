package benchmarkconsensus

import (
	"github.com/orbs-network/orbs-network-go/crypto/digest"
	"github.com/orbs-network/orbs-network-go/crypto/logic"
	"github.com/orbs-network/orbs-network-go/crypto/signature"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/services"
	"github.com/pkg/errors"
	"math"

	//"math"
	"context"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
)

func (s *service) getLastCommittedBlock() (primitives.BlockHeight, *protocol.BlockPairContainer) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if s.lastCommittedBlockUnderMutex == nil {
		return 0, nil
	}
	return s.lastCommittedBlockUnderMutex.TransactionsBlock.Header.BlockHeight(), s.lastCommittedBlockUnderMutex
}

func (s *service) setLastCommittedBlock(blockPair *protocol.BlockPairContainer, expectedLastCommittedBlockBefore *protocol.BlockPairContainer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.lastCommittedBlockUnderMutex != expectedLastCommittedBlockBefore {
		return errors.New("aborting shared state update due to inconsistency")
	}
	s.lastCommittedBlockUnderMutex = blockPair
	s.lastCommittedBlockVotersUnderMutex = make(map[string]bool) // leader only
	s.lastCommittedBlockVotersReachedQuorumUnderMutex = false    // leader only

	return nil
}

func (s *service) requiredQuorumSize() int {
	return int(math.Ceil(float64(s.config.NetworkSize(0)) * float64(s.config.ConsensusRequiredQuorumPercentage()/100)))
}

func (s *service) saveToBlockStorage(ctx context.Context, blockPair *protocol.BlockPairContainer) error {
	if blockPair.TransactionsBlock.Header.BlockHeight() == 0 {
		return nil
	}
	s.logger.Info("saving block to storage", log.BlockHeight(blockPair.TransactionsBlock.Header.BlockHeight()))
	_, err := s.blockStorage.CommitBlock(ctx, &services.CommitBlockInput{
		BlockPair: blockPair,
	})
	return err
}

func (s *service) validateBlockConsensus(blockPair *protocol.BlockPairContainer, prevCommittedBlockPair *protocol.BlockPairContainer) error {
	// correct block type
	if !blockPair.TransactionsBlock.BlockProof.IsTypeBenchmarkConsensus() {
		return errors.Errorf("incorrect block proof type: %s", blockPair.TransactionsBlock.BlockProof.Type())
	}
	if !blockPair.ResultsBlock.BlockProof.IsTypeBenchmarkConsensus() {
		return errors.Errorf("incorrect block proof type: %s", blockPair.ResultsBlock.BlockProof.Type())
	}

	// prev block hash ptr (if given)
	if prevCommittedBlockPair != nil {
		prevTxHash := digest.CalcTransactionsBlockHash(prevCommittedBlockPair.TransactionsBlock)
		if !blockPair.TransactionsBlock.Header.PrevBlockHashPtr().Equal(prevTxHash) {
			return errors.Errorf("transactions prev block hash does not match prev block: %s", prevTxHash)
		}
		prevRxHash := digest.CalcResultsBlockHash(prevCommittedBlockPair.ResultsBlock)
		if !blockPair.ResultsBlock.Header.PrevBlockHashPtr().Equal(prevRxHash) {
			return errors.Errorf("results prev block hash does not match prev block: %s", prevRxHash)
		}
	}

	// block proof
	blockProof := blockPair.ResultsBlock.BlockProof.BenchmarkConsensus()
	if !blockProof.Sender().SenderPublicKey().Equal(s.config.ConstantConsensusLeader()) {
		return errors.Errorf("block proof not from leader: %s", blockProof.Sender().SenderPublicKey())
	}
	signedData := s.signedDataForBlockProof(blockPair)
	if !signature.VerifyEd25519(blockProof.Sender().SenderPublicKey(), signedData, blockProof.Sender().Signature()) {
		return errors.Errorf("block proof signature is invalid: %s", blockProof.Sender().Signature())
	}

	return nil
}

func (s *service) signedDataForBlockProof(blockPair *protocol.BlockPairContainer) []byte {
	txHash := digest.CalcTransactionsBlockHash(blockPair.TransactionsBlock)
	rxHash := digest.CalcResultsBlockHash(blockPair.ResultsBlock)
	xorHash := logic.CalcXor(txHash, rxHash)
	return xorHash
}

func (s *service) handleBlockConsensusFromHandler(mode handlers.HandleBlockConsensusMode, blockType protocol.BlockType, blockPair *protocol.BlockPairContainer, prevCommittedBlockPair *protocol.BlockPairContainer) error {
	if blockType != protocol.BLOCK_TYPE_BLOCK_PAIR {
		return errors.Errorf("handler received unsupported block type %s", blockType)
	}

	// validate the block consensus
	if mode == handlers.HANDLE_BLOCK_CONSENSUS_MODE_VERIFY_AND_UPDATE || mode == handlers.HANDLE_BLOCK_CONSENSUS_MODE_VERIFY_ONLY {
		err := s.validateBlockConsensus(blockPair, prevCommittedBlockPair)
		if err != nil {
			return err
		}
	}

	// update lastCommitted to reflect this if newer
	if mode == handlers.HANDLE_BLOCK_CONSENSUS_MODE_VERIFY_AND_UPDATE || mode == handlers.HANDLE_BLOCK_CONSENSUS_MODE_UPDATE_ONLY {
		_lastCommittedBlockHeight, _lastCommittedBlock := s.getLastCommittedBlock()

		if blockPair.TransactionsBlock.Header.BlockHeight() > _lastCommittedBlockHeight {
			err := s.setLastCommittedBlock(blockPair, _lastCommittedBlock)
			if err != nil {
				return err
			}
			// don't forget to update internal vars too since they may be used later on in the function
			_lastCommittedBlock = blockPair
			_lastCommittedBlockHeight = _lastCommittedBlock.TransactionsBlock.Header.BlockHeight()
		}
	}

	return nil
}
