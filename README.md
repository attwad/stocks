# Stock price (+currency) alerts

Simple utility that uses the polygon.io API to check for the last day's traded value of a stock, and fail if the price is above a certain threshold for a given currency.

## Why?

I wanted a simple tool that would alert me when some stock was above a certain price for a given currency and couldn't find anything online that did it.

## How?

```shell
API_KEY=<your polygon.io API key (free tier limited to 5 RPM)> go run . --ticker=GOOGL --want_min_prices=170CHF,150USD
```

Here `GOOGL` is traded in USD, when the stock goes over 170CFH or 150USD according to the same polygon API (last close price), the program will exit with a non 0 status code.

You can couple that exit behavior with a crontab or monit cron job and get alerted this way.