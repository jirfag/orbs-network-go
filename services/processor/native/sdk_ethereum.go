package native

import (
	"context"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/orbs-network/orbs-contract-sdk/go/sdk"
	"github.com/orbs-network/orbs-spec/types/go/primitives"
	"github.com/orbs-network/orbs-spec/types/go/protocol"
	"github.com/orbs-network/orbs-spec/types/go/services/handlers"
	"github.com/pkg/errors"
	"strings"
)

type ethereumSdk struct {
	handler         handlers.ContractSdkCallHandler
	permissionScope protocol.ExecutionPermissionScope
}

const SDK_OPERATION_NAME_ETHEREUM = "Sdk.Ethereum"

func (s *ethereumSdk) CallMethod(executionContextId sdk.Context, contractAddress string, jsonAbi string, methodName string, out interface{}, args ...interface{}) error {
	packedInput, err := ethereumPackInputArguments(jsonAbi, methodName, args...)
	if err != nil {
		return err
	}

	output, err := s.handler.HandleSdkCall(context.TODO(), &handlers.HandleSdkCallInput{
		ContextId:     primitives.ExecutionContextId(executionContextId),
		OperationName: SDK_OPERATION_NAME_ETHEREUM,
		MethodName:    "callMethod",
		InputArguments: []*protocol.MethodArgument{
			(&protocol.MethodArgumentBuilder{
				Name:        "contractAddress",
				Type:        protocol.METHOD_ARGUMENT_TYPE_STRING_VALUE,
				StringValue: contractAddress,
			}).Build(),
			(&protocol.MethodArgumentBuilder{
				Name:        "jsonAbi",
				Type:        protocol.METHOD_ARGUMENT_TYPE_STRING_VALUE,
				StringValue: jsonAbi,
			}).Build(),
			(&protocol.MethodArgumentBuilder{
				Name:        "methodName",
				Type:        protocol.METHOD_ARGUMENT_TYPE_STRING_VALUE,
				StringValue: methodName,
			}).Build(),
			(&protocol.MethodArgumentBuilder{
				Name:       "ethereumPackedInputArguments",
				Type:       protocol.METHOD_ARGUMENT_TYPE_BYTES_VALUE,
				BytesValue: packedInput,
			}).Build(),
		},
		PermissionScope: s.permissionScope,
	})
	if err != nil {
		return err
	}
	if len(output.OutputArguments) != 1 || !output.OutputArguments[0].IsTypeBytesValue() {
		return errors.Errorf("callMethod Sdk.Ethereum returned corrupt output value")
	}

	return ethereumUnpackOutput(out, output.OutputArguments[0].BytesValue())
}

func ethereumPackInputArguments(jsonAbi string, method string, args ...interface{}) ([]byte, error) {
	if parsedABI, err := abi.JSON(strings.NewReader(jsonAbi)); err != nil {
		return nil, errors.WithStack(err)
	} else {
		return parsedABI.Pack(method, args...)
	}
}

// TODO: @jlevison add
func ethereumUnpackOutput(out interface{}, packedOutput []byte) error {
	return nil
}