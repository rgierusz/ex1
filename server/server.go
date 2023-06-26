package server

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rgierusz/ex1/metric"
	"log"
	"math/big"
	"net/http"
)

const (
	serverAddr = ":8081"

	pathVariableGetBalance = "address"
	pathSegmentWei         = "wei"

	pathMetric      = "/metrics"
	pathHealthCheck = "/healthz"

	pathGetBalance          = "/eth/balance/{" + pathVariableGetBalance + "}"
	pathGetBalanceWeiSuffix = "/{" + pathSegmentWei + ":(?i)" + pathSegmentWei + "}" //optional suffix indicating wei response format
)

var gatewayURLs = [3]string{
	"https://mainnet.infura.io/v3/b3bd9456e8d44150b963248668023317",         // recommended by Alluvial
	"https://mainnet.infura.io/v3/62b943de85fe49078034886e04feeb81",         // first backup, created just for the exercise
	"https://eth-mainnet.g.alchemy.com/v2/qa0ADnNx_f93xDT6ZLKcPKv84uu-Nj5T", // second backup, created just for the exercise
}

type BalanceResponse struct {
	Balance    string `json:"balance,omitempty"`
	WeiBalance string `json:"weiBalance,omitempty"`
}

// client is cached for performance reasons
var ethClient *ethclient.Client

func InitServer() {
	router := mux.NewRouter()

	router.HandleFunc(pathMetric, promhttp.Handler().ServeHTTP)
	router.HandleFunc(pathHealthCheck, metric.HandlerMetricsWrapper(metric.HealthCheckStartedCounter, metric.HealthCheckCompletedCounter, livenessHandler))

	balanceHandle := metric.HandlerMetricsWrapper(metric.GetBalanceStartedCounter, metric.GetBalanceCompletedCounter, balanceHandler)
	router.HandleFunc(pathGetBalance, balanceHandle)
	router.PathPrefix(pathGetBalance).Path(pathGetBalanceWeiSuffix).HandlerFunc(balanceHandle)

	log.Fatal(http.ListenAndServe(serverAddr, router))
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	hexAddress := vars[pathVariableGetBalance]
	if !common.IsHexAddress(hexAddress) {
		http.Error(w, "Incorrect ETH address", http.StatusBadRequest)
		return
	}

	if balance, e, c := getBalance(r, common.HexToAddress(hexAddress)); e != nil {
		http.Error(w, e.Error(), c)
	} else {
		var resp BalanceResponse

		if _, weiRequest := vars[pathSegmentWei]; weiRequest {
			resp = BalanceResponse{WeiBalance: balance.String()}
		} else {
			balanceEther := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(params.Ether))
			resp = BalanceResponse{Balance: fmt.Sprintf("%.18f", balanceEther)}
		}

		if jsonResp, jsonErr := json.Marshal(resp); jsonErr != nil {
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")

			if _, e := fmt.Fprintf(w, string(jsonResp)); e != nil {
				metric.GenericResponseErrorCounter.WithLabelValues(e.Error()).Inc()
				log.Printf("Error writing balance response: %v", e)
			}
		}
	}
}

func getBalance(r *http.Request, account common.Address) (*big.Int, error, int) {
	if ethClient == nil {
		if e, c := refreshClient(); e != nil {
			return nil, e, c
		}
	}

	if balance, e := ethClient.BalanceAt(r.Context(), account, nil); e != nil {
		metric.ETHCallCounter.WithLabelValues(e.Error()).Inc()

		ethClient.Close()
		ethClient = nil

		return nil, e, http.StatusInternalServerError
	} else {
		metric.ETHCallCounter.WithLabelValues("").Inc()

		return balance, nil, 0
	}
}

func refreshClient() (error, int) {
	for i, url := range gatewayURLs {
		if c, e := ethclient.Dial(url); e == nil && c != nil {
			// client initiated, all good
			ethClient = c // update global client
			break
		} else {
			metric.ETHCallCounter.WithLabelValues(e.Error()).Inc()

			if i == len(gatewayURLs)-1 {
				return e, http.StatusInternalServerError
			}
		}
	}

	return nil, 0
}
