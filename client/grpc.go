package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	liquiditytypes "github.com/crescent-network/crescent/x/liquidity/types"
	liquidstakingtypes "github.com/crescent-network/crescent/x/liquidstaking/types"
)

type GRPCClient struct {
	conn *grpc.ClientConn
}

func ConnectGRPC(ctx context.Context, addr string, opts ...grpc.DialOption) (*GRPCClient, error) {
	conn, err := grpc.DialContext(ctx, addr, append(opts, grpc.WithBlock())...)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &GRPCClient{conn: conn}, nil
}

func ConnectGRPCWithTimeout(ctx context.Context, addr string, timeout time.Duration, opts ...grpc.DialOption) (*GRPCClient, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return ConnectGRPC(ctx, addr, opts...)
}

func (c *GRPCClient) Close() error {
	return c.Close()
}

func (c *GRPCClient) QueryPairs(ctx context.Context) (*liquiditytypes.QueryPairsResponse, error) {
	return liquiditytypes.NewQueryClient(c.conn).Pairs(ctx, &liquiditytypes.QueryPairsRequest{})
}

func (c *GRPCClient) QueryPools(ctx context.Context) (*liquiditytypes.QueryPoolsResponse, error) {
	return liquiditytypes.NewQueryClient(c.conn).Pools(ctx, &liquiditytypes.QueryPoolsRequest{})
}

func (c *GRPCClient) QueryLiquidStakingStates(ctx context.Context) (*liquidstakingtypes.QueryStatesResponse, error) {
	return liquidstakingtypes.NewQueryClient(c.conn).States(ctx, &liquidstakingtypes.QueryStatesRequest{})
}

func (c *GRPCClient) QueryBalances(ctx context.Context, addr string) (*banktypes.QueryAllBalancesResponse, error) {
	return banktypes.NewQueryClient(c.conn).AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address:    addr,
		Pagination: nil,
	})
}
