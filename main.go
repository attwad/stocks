package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"github.com/shopspring/decimal"
)

var (
	ticker       = flag.String("ticker", "GOOGL", "Ticker to look for last close value")
	wantMinPrice = flag.String("want_min_price", "150CHF", "minumum price wanted in to trigger an alert, must finish with wanted currency")
)

func getBaseCurrency(ctx context.Context, c *polygon.Client, ticker string) (string, error) {
	res, err := c.GetTickerDetails(ctx, &models.GetTickerDetailsParams{
		Ticker: ticker,
	})
	if err != nil {
		return "", err
	}
	return strings.ToUpper(res.Results.CurrencyName), nil
}

// parsePrice parses 150CHF into 150.0 and CHF
func parsePrice(price string) (decimal.Decimal, string, error) {
	// Just start from the end, go until something else than an ascii letter is found.
	for i := len(price) - 1; i > 0; i-- {
		if unicode.IsLetter(rune(price[i])) {
			continue
		}
		val, err := strconv.ParseFloat(price[0:i+1], 64)
		if err != nil {
			return decimal.Decimal{}, "", err
		}
		return decimal.NewFromFloat(val), strings.TrimSpace(strings.ToUpper(price[i+1:])), nil
	}
	return decimal.Decimal{}, "", fmt.Errorf("invalid price format: %s", price)
}

func getLastClosePrice(ctx context.Context, c *polygon.Client, ticker string) (decimal.Decimal, error) {
	res, err := c.GetPreviousCloseAgg(ctx, &models.GetPreviousCloseAggParams{Ticker: ticker})
	if err != nil {
		return decimal.Decimal{}, err
	}
	if got, want := len(res.Results), 1; got != want {
		log.Fatalf("got %d results for ticker %s, wanted=%d", got, ticker, want)
	}
	return decimal.NewFromFloat(res.Results[0].Close), nil
}

func main() {
	flag.Parse()
	ctx := context.Background()
	c := polygon.New(os.Getenv("API_KEY"))
	// Parse input.
	wantPriceDestCurrency, destCurrencySymbol, err := parsePrice(*wantMinPrice)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("want price dest curr=", wantPriceDestCurrency)
	fmt.Println("dest curr=", destCurrencySymbol)
	// Get previous close value
	lastClosePrice, err := getLastClosePrice(ctx, c, *ticker)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("last close price=", lastClosePrice)

	// Get currency the stock is traded in.
	baseCurrencySymbol, err := getBaseCurrency(ctx, c, *ticker)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*ticker, "is traded in", baseCurrencySymbol)
	conversionRate := decimal.NewFromFloat(1.0)
	if baseCurrencySymbol != destCurrencySymbol {
		// Get exchange rate between currencies.
		currencyLastClosePrice, err := getLastClosePrice(ctx, c, fmt.Sprintf("C:%s%s", baseCurrencySymbol, destCurrencySymbol))
		if err != nil {
			log.Fatal(err)
		}
		conversionRate = currencyLastClosePrice
	}
	convertedClosePrice := lastClosePrice.Mul(conversionRate)
	fmt.Println(lastClosePrice, baseCurrencySymbol, "is equal to", convertedClosePrice, destCurrencySymbol)

	if convertedClosePrice.Cmp(wantPriceDestCurrency) > 0 {
		fmt.Println("Prices match, exiting the program with an error code")
		os.Exit(1)
	}
}
