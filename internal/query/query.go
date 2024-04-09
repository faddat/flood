package query

import (
	"google.golang.org/grpc"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	clquery "github.com/osmosis-labs/osmosis/v24/x/concentrated-liquidity/client/queryproto"
	pmquery "github.com/osmosis-labs/osmosis/v24/x/poolmanager/client/queryproto"
)

// QueryClient is a wrapper with Cosmos and Osmosis grpc query clients
type QueryClient struct {
	// cosmos-sdk query clients
	Authz authz.QueryClient
	Bank  banktypes.QueryClient

	// wasmd query clients
	Wasm wasmtypes.QueryClient

	// osmosis query clients
	ConcentratedLiquidity clquery.QueryClient
	PoolManager           pmquery.QueryClient
}

// NewQueryClient creates a new QueryClient and initializes all the module query clients
func NewQueryClient(conn *grpc.ClientConn) (*QueryClient, error) {
	client := &QueryClient{
		Authz:                 authz.NewQueryClient(conn),
		Bank:                  banktypes.NewQueryClient(conn),
		Wasm:                  wasmtypes.NewQueryClient(conn),
		ConcentratedLiquidity: clquery.NewQueryClient(conn),
		PoolManager:           pmquery.NewQueryClient(conn),
	}
	return client, nil
}
