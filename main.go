package main

import (
	"fmt"
	"os"

	"github.com/Enclave-Markets/enclave-go/apiclient"
	"github.com/Enclave-Markets/enclave-go/models"
	"github.com/shopspring/decimal"
)

func main() {

	// the trading pair that will be used, and balance symbol to check
	symbol := models.Symbol("AVAX")
	market := models.Market("AVAX-USDC")

	client, err := apiclient.NewApiClientFromEnv("sandbox")
	if err != nil {
		fmt.Println("failed to create client", err)
		return
	}

	client.WithApiKey(
		os.Getenv("ENCLAVE_KEY"),
		os.Getenv("ENCLAVE_SECRET"),
	)

	client.WaitForEndpoint()
	_, err = client.AuthedHello()
	if err != nil {
		fmt.Println("authed-hello failed:", err)
		return
	}

	// get the balance of AVAX
	balanceResp, err := client.GetBalance(models.GetBalanceReq{Symbol: symbol})
	if err != nil {
		fmt.Println("failed to get balance:", err)
		return
	}
	fmt.Println(symbol, "balance:", balanceResp.Result.TotalBalance)

	// get the AVAX-USDC trading pair to find the min order sizes
	marketsResp, err := client.Markets()
	if err != nil {
		fmt.Println("failed to get markets", err, marketsResp.Success)
		return
	}

	var baseMin, quoteMin decimal.Decimal
	for _, pair := range marketsResp.Result.Spot.TradingPairs {
		if pair.Market != market {
			continue
		}
		baseMin, quoteMin = pair.BaseIncrement, pair.QuoteIncrement
	}
	fmt.Println("base-min:", baseMin, "quote-min:", quoteMin)

	// get top of book for avax usdc
	book, err := client.GetSpotDepthBook(market)
	if err != nil {
		fmt.Println("failed to get market depth:", err)
		return
	}

	if len(book.Result.Asks) == 0 {
		fmt.Println("no asks are resting on the book")
		return
	}
	bestAsk := book.Result.Asks[0]
	fmt.Println("best-ask-price:", bestAsk.Price, "best-ask-size:", bestAsk.Quantity)

	// place a sell limit order of the smallest size one tick above the top of book (so we don't get filled)
	orderResp, err := client.AddSpotOrder(
		models.AddOrderReq{
			Market: market,
			Side:   models.Ask,
			Price:  bestAsk.Price.Add(quoteMin),
			Size:   baseMin,
			Type:   models.OrderTypeLimit,
		},
	)
	if err != nil {
		fmt.Println("failed to place order:", err)
		return
	}
	fmt.Println("order placed, current state:", orderResp.Result.State)

	// cancel all orders in the market
	if err := client.CancelAllSpotOrders(); err != nil {
		fmt.Println("failed to cancel spot orders:", err)
		return
	}

	// get the order state
	orderResp, err = client.GetSpotOrder(orderResp.Result.OrderID)
	if err != nil {
		fmt.Println("failed to get spot order:", err)
		return
	}

	fmt.Println("order state:", orderResp.Result.State)
}
