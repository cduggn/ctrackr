# cTracker

A `Go` library which supports fetching transaction history from the [Coinbase API](https://developers.coinbase.com/). The application supports the [API key Authentication method](https://developers.coinbase.com/docs/wallet/api-key-authentication). * Experimental use only.

## Prerequisites

1.  [Enable API Key access](https://www.coinbase.com/settings/api) for your Coinbase account. 

## Installation

Get the latest version of go-coin library:
```
go get github.com/cduggn/ctrackr
```

## Usage

#### Create Client 

Create a new instance of BrokerService by passing in your Coinbase Api credentials. [godotenv](https://github.com/joho/godotenv) or similar can be used to load credentials from a environment file.

>**Note**: Do not store credentials in source control. 

```Go
package main
import (
	gocoin "github.com/cdugga/go-coin"
	"github.com/joho/godotenv"
	"os"
)

func main(){
	godotenv.Load()
	
	k := os.Getenv("<your key>")
    s :=  os.Getenv("<your secret>")
    
    svc := gocoin.NewBrokerService(k, s)
}
 
```

#### Get Accounts

Returns your personal accounts. The list of returned accounts may vary depending on permissions assigned to API Key. Accounts represent Coinbase Wallets, Vault, and Fiat. Data fields returned include name, ID, currency, amount, and type.

```Go
package main

import (
	gocoin "github.com/cdugga/go-coin"
	"go.uber.org/zap"
	"os"
)

var logger *zap.Logger

func main(){
	k := os.Getenv("<your key>")
	s :=  os.Getenv("<your secret>")

	svc := gocoin.NewBrokerService(k, s)
    accounts, err := svc.Client.GetAccounts()
    if err != nil {
        logger.Error("Something went wrong ", zap.Error(err))
    }
	// do something with account IDs
}
```
#### Get Account Activity

Returns buy and sell orders. Uses the `Go worker pool` pattern to fetch buy and sell orders associated with each account ID, passed to the request. Worker concurrency set to 3 by [default](https://github.com/cdugga/go-coin/blob/main/brokerService.go#L15). 

```Go

func main(){
    k := os.Getenv("<your key>")
    s :=  os.Getenv("<your secret>")
    
    svc := gocoin.NewBrokerService(k, s)
    accounts, err := svc.Client.GetAccounts()
    if err != nil {
        logger.Error("Something went wrong ", zap.Error(err))
    }
    ctx, cancel := context.WithCancel(context.Background())
    
    res := svc.Client.GetAccountActivity(ctx, accounts)
	
	// do something with return orders
}
```