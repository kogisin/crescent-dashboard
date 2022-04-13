package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hallazzang/crescent-dashboard/client"
)

func NewRootCmd() *cobra.Command {
	var (
		grpcInsecure bool
	)
	cmd := &cobra.Command{
		Use:   "crescent-dashboard [grpc-addr] [api-base-url]",
		Short: "Crescent Dashboard Server",
		Long: `Crescent Dashboard Server

Examples:
  crescent-dashboard mainnet.crescent.network:9090 https://apigw.crescent.network/
  crescent-dashboard mainnet.crescent.network:9090 https://apigw.crescent.network/ --insecure
`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			grpcAddr := args[0]
			apiBaseURL := args[1]

			var opts []grpc.DialOption
			if grpcInsecure {
				opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
			} else {
				opts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))}
			}

			grpcClient, err := client.ConnectGRPCWithTimeout(context.Background(), grpcAddr, 5*time.Second, opts...)
			if err != nil {
				return fmt.Errorf("connect grpc: %w", err)
			}
			defer grpcClient.Close()

			apiClient, err := client.NewAPIClient(apiBaseURL)
			if err != nil {
				return fmt.Errorf("new api client: %w", err)
			}

			c := NewCollector(grpcClient, apiClient)
			prometheus.MustRegister(c)

			var wg sync.WaitGroup

			launchPeriodicTask := func(interval time.Duration, f func()) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					ticker := time.NewTicker(time.Second)
					defer ticker.Stop()
					for range ticker.C {
						f()
					}
				}()
			}

			ctx := context.Background()

			launchPeriodicTask(time.Minute, func() {
				if err := c.UpdatePairs(ctx); err != nil {
					log.Printf("error: failed to update pairs: %v", err)
				}
			})
			launchPeriodicTask(2*time.Second, func() {
				if err := c.UpdatePools(ctx); err != nil {
					log.Printf("error: failed to update pools: %v", err)
				}
			})
			launchPeriodicTask(2*time.Second, func() {
				if err := c.UpdatePrices(ctx); err != nil {
					log.Printf("error: failed to update prices: %v", err)
				}
			})
			launchPeriodicTask(2*time.Second, func() {
				if err := c.UpdateBTokenState(ctx); err != nil {
					log.Printf("error: failed to update bToken state: %v", err)
				}
			})

			http.Handle("/metrics", promhttp.Handler())
			log.Printf("info: running prometheus server")
			log.Fatal(http.ListenAndServe(":2112", nil))

			return nil
		},
	}
	cmd.Flags().BoolVar(&grpcInsecure, "insecure", false, "Use insecure transport for gRPC")
	return cmd
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
