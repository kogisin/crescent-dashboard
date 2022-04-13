package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/hallazzang/crescent-dashboard/client"
)

type Pair struct {
	ID             uint64
	BaseCoinDenom  string
	QuoteCoinDenom string
	NumOrders      uint64
	LastPrice      *float64
	CurrentBatchID uint64
}

type Pool struct {
	ID                  uint64
	PairID              uint64
	NumDepositRequests  uint64
	NumWithdrawRequests uint64
	Price               float64
	Value               float64
}

type LiquidStakingState struct {
	MintRate     float64
	BTokenSupply float64
}

type Collector struct {
	grpcClient *client.GRPCClient
	apiClient  *client.APIClient

	pairs      map[uint64]Pair
	pairsMux   sync.RWMutex
	pools      map[uint64]Pool
	poolsMux   sync.RWMutex
	prices     map[string]float64
	pricesMux  sync.RWMutex
	lsState    *LiquidStakingState
	lsStateMux sync.RWMutex

	numOrders       *prometheus.Desc
	numDepositReqs  *prometheus.Desc
	numWithdrawReqs *prometheus.Desc
	lastPrice       *prometheus.Desc
	poolPrice       *prometheus.Desc
	poolValue       *prometheus.Desc
	mintRate        *prometheus.Desc
	bTokenSupply    *prometheus.Desc
	price           *prometheus.Desc
}

func NewCollector(grpcClient *client.GRPCClient, apiClient *client.APIClient) *Collector {
	return &Collector{
		grpcClient:      grpcClient,
		apiClient:       apiClient,
		numOrders:       prometheus.NewDesc("crescent_num_orders", "Number of orders", []string{"pair_id"}, nil),
		numDepositReqs:  prometheus.NewDesc("crescent_num_deposit_requests", "Number of deposit requests", []string{"pool_id"}, nil),
		numWithdrawReqs: prometheus.NewDesc("crescent_num_withdraw_requests", "Number of withdraw requests", []string{"pool_id"}, nil),
		lastPrice:       prometheus.NewDesc("crescent_last_price", "Pair's last price", []string{"pair_id"}, nil),
		poolPrice:       prometheus.NewDesc("crescent_pool_price", "Pool price", []string{"pool_id"}, nil),
		poolValue:       prometheus.NewDesc("crescent_pool_value", "Pool value", []string{"pool_id"}, nil),
		mintRate:        prometheus.NewDesc("crescent_mint_rate", "bToken mint rate", nil, nil),
		bTokenSupply:    prometheus.NewDesc("crescent_btoken_supply", "bToken total supply", nil, nil),
		price:           prometheus.NewDesc("crescent_price", "Coin price", []string{"denom"}, nil),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numOrders
	ch <- c.numDepositReqs
	ch <- c.numWithdrawReqs
	ch <- c.lastPrice
	ch <- c.poolPrice
	ch <- c.poolValue
	ch <- c.mintRate
	ch <- c.bTokenSupply
	ch <- c.price
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	withRLock := func(mux *sync.RWMutex, f func()) {
		mux.RLock()
		defer mux.RUnlock()
		f()
	}
	withRLock(&c.pairsMux, func() {
		for _, pair := range c.pairs {
			ch <- prometheus.MustNewConstMetric(c.numOrders, prometheus.GaugeValue, float64(pair.NumOrders), strconv.FormatUint(pair.ID, 10))
			if pair.LastPrice != nil {
				ch <- prometheus.MustNewConstMetric(c.lastPrice, prometheus.GaugeValue, *pair.LastPrice, strconv.FormatUint(pair.ID, 10))
			}
		}
	})
	withRLock(&c.poolsMux, func() {
		for _, pool := range c.pools {
			ch <- prometheus.MustNewConstMetric(c.numDepositReqs, prometheus.GaugeValue, float64(pool.NumDepositRequests), strconv.FormatUint(pool.ID, 10))
			ch <- prometheus.MustNewConstMetric(c.numWithdrawReqs, prometheus.GaugeValue, float64(pool.NumWithdrawRequests), strconv.FormatUint(pool.ID, 10))
			ch <- prometheus.MustNewConstMetric(c.poolPrice, prometheus.GaugeValue, pool.Price, strconv.FormatUint(pool.ID, 10))
			ch <- prometheus.MustNewConstMetric(c.poolValue, prometheus.GaugeValue, pool.Value, strconv.FormatUint(pool.ID, 10))
		}
	})
	withRLock(&c.pricesMux, func() {
		for denom, price := range c.prices {
			ch <- prometheus.MustNewConstMetric(c.price, prometheus.GaugeValue, price, denom)
		}
	})
	withRLock(&c.lsStateMux, func() {
		if c.lsState != nil {
			ch <- prometheus.MustNewConstMetric(c.mintRate, prometheus.GaugeValue, c.lsState.MintRate)
			ch <- prometheus.MustNewConstMetric(c.bTokenSupply, prometheus.GaugeValue, c.lsState.BTokenSupply)
		}
	})
}

func (c *Collector) UpdatePairs(ctx context.Context) error {
	resp, err := c.grpcClient.QueryPairs(ctx)
	if err != nil {
		return err
	}
	c.pairsMux.Lock()
	defer c.pairsMux.Unlock()
	c.pairs = map[uint64]Pair{}
	for _, pair := range resp.Pairs {
		var lastPrice *float64
		if pair.LastPrice != nil {
			p := pair.LastPrice.MustFloat64()
			lastPrice = &p
		}
		c.pairs[pair.Id] = Pair{
			ID:             pair.Id,
			BaseCoinDenom:  pair.BaseCoinDenom,
			QuoteCoinDenom: pair.QuoteCoinDenom,
			NumOrders:      pair.LastOrderId,
			LastPrice:      lastPrice,
			CurrentBatchID: pair.CurrentBatchId,
		}
	}
	return nil
}

func (c *Collector) UpdatePools(ctx context.Context) error {
	resp, err := c.grpcClient.QueryPools(ctx)
	if err != nil {
		return err
	}
	c.pairsMux.RLock()
	defer c.pairsMux.RUnlock()
	if len(c.pairs) == 0 { // no pairs yet
		return nil
	}
	c.pricesMux.RLock()
	defer c.pricesMux.RUnlock()
	if len(c.prices) == 0 { // no prices yet
		return nil
	}
	c.poolsMux.Lock()
	defer c.poolsMux.Unlock()
	c.pools = map[uint64]Pool{}
	for _, pool := range resp.Pools {
		pair, ok := c.pairs[pool.PairId]
		if !ok {
			return fmt.Errorf("pair not found: %d", pool.PairId)
		}
		price := pool.Balances.AmountOf(pair.QuoteCoinDenom).ToDec().Quo(
			pool.Balances.AmountOf(pair.BaseCoinDenom).ToDec()).MustFloat64()
		value := 0.0
		for _, coin := range pool.Balances {
			p, ok := c.prices[coin.Denom]
			if !ok {
				return fmt.Errorf("price not found: %s", coin.Denom)
			}
			value += p * (float64(coin.Amount.Int64()) / 1000000) // TODO: is it right type conversion?
		}
		c.pools[pool.Id] = Pool{
			ID:                  pool.Id,
			PairID:              pool.PairId,
			NumDepositRequests:  pool.LastDepositRequestId,
			NumWithdrawRequests: pool.LastWithdrawRequestId,
			Price:               price,
			Value:               value,
		}
	}
	return nil
}

func (c *Collector) UpdatePrices(ctx context.Context) error {
	c.pricesMux.Lock()
	defer c.pricesMux.Unlock()
	prices, err := c.apiClient.Prices(ctx)
	if err != nil {
		return err
	}
	c.prices = prices
	return nil
}

func (c *Collector) UpdateBTokenState(ctx context.Context) error {
	c.lsStateMux.Lock()
	defer c.lsStateMux.Unlock()
	resp, err := c.grpcClient.QueryLiquidStakingStates(ctx)
	if err != nil {
		return err
	}
	c.lsState = &LiquidStakingState{
		MintRate:     resp.NetAmountState.MintRate.MustFloat64(),
		BTokenSupply: float64(resp.NetAmountState.BtokenTotalSupply.Int64()) / 1000000,
	}
	return nil
}
