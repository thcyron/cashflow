package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/fftoml"

	"github.com/thcyron/cashflow/internal/api"
	"github.com/thcyron/cashflow/internal/cf"
	"github.com/thcyron/cashflow/internal/price/cache"
	"github.com/thcyron/cashflow/internal/price/yahoo"
	"github.com/thcyron/cashflow/internal/repository/fs"
	"github.com/thcyron/cashflow/internal/repository/git"
)

func main() {
	logger := makeLogger()
	flagSet := flag.NewFlagSet("cashflow-server", flag.ExitOnError)

	var (
		fsDir = flagSet.String("fs.dir", "", "Path to local portfolio directory")

		gitURL  = flagSet.String("git.url", "", "Git repository URL")
		gitUser = flagSet.String("git.user", "git", "Git SSH username")
		gitKey  = flagSet.String("git.key", "", "Git SSH private key")

		_ = flagSet.String("config", "", "config file (optional)")
	)

	if err := ff.Parse(flagSet, os.Args[1:],
		ff.WithEnvVarPrefix("CASHFLOW"),
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(fftoml.Parser),
	); err != nil {
		logger.Log("msg", "error parsing flags", "err", err)
		os.Exit(1)
	}

	var repo cf.Repository
	if *fsDir != "" {
		repo = fs.NewRepository(*fsDir)
	} else if *gitURL != "" {
		gitRepo, err := git.NewRepositoryWithSSH(*gitURL, *gitUser, []byte(*gitKey))
		if err != nil {
			logger.Log(
				"msg", "error initializing git repository",
				"err", err,
			)
			os.Exit(1)
		}
		repo = gitRepo
	} else {
		logger.Log("err", "neither -fs.dir nor -git.url provided")
		os.Exit(1)
	}

	var (
		yahooClient        = yahoo.NewClient()
		yahooPriceProvider = yahoo.NewProvider(yahooClient)
		priceCache         = cache.New(yahooPriceProvider)
		api                = api.New(log.With(logger, "component", "api"), repo, priceCache.Price)
		runGroup           run.Group
	)

	runGroup.Add(run.SignalHandler(context.Background(), syscall.SIGTERM, syscall.SIGINT))
	runGroup.Add(apiServer(api))
	runGroup.Add(priceUpdater(logger, repo, priceCache))

	if err := runGroup.Run(); err != nil {
		if errors.As(err, &run.SignalError{}) {
			logger.Log("msg", err)
		} else {
			logger.Log("err", err)
			os.Exit(1)
		}
	}
}

func apiServer(server *api.Server) (execute func() error, interrupt func(error)) {
	var listener net.Listener
	return func() error {
			ln, err := net.Listen("tcp", ":8080")
			if err != nil {
				return err
			}
			listener = ln
			return http.Serve(ln, server)
		}, func(error) {
			if listener != nil {
				listener.Close()
			}
		}
}

func priceUpdater(logger log.Logger, repo cf.Repository, cache *cache.Cache) (execute func() error, interrupt func(error)) {
	ctx, cancel := context.WithCancel(context.Background())

	update := func() {
		logger.Log("msg", "updating prices")
		if err := updatePrices(ctx, repo, cache); err != nil {
			logger.Log(
				"msg", "error updating prices",
				"err", err,
			)
			return
		}
		logger.Log("msg", "prices updated")
	}

	return func() error {
			update()
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(15 * time.Minute):
					update()
				}
			}
		}, func(error) {
			cancel()
		}
}

func updatePrices(ctx context.Context, repo cf.Repository, cache *cache.Cache) error {
	stocks, err := repo.Stocks(ctx)
	if err != nil {
		return err
	}
	return cache.UpdateHistory(ctx, stocks)
}

func makeLogger() log.Logger {
	var logger log.Logger
	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}
