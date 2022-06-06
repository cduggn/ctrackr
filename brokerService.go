package gocoin

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

const (
	GET_ACCOUNTS_PATH = "/v2/accounts"
	BASE_URL          = "https://api.coinbase.com"
	workerConcurrency = 3
)

var env *Environment

var logger *zap.Logger

type Environment struct {
	APIKey    string
	APISecret string
}

type Broker interface {
	GetAccounts() ([]GenericAccount, error)
	GetAccountActivity(ctx context.Context, acc []GenericAccount) (t []AccountActivity)
	GetTransactions(nextUrl string)
}

type Account interface {
	getPaginatedAccountResults(ctx context.Context, nextUrl string) (<-chan GenericAccount, <-chan error)
	getAccountByURI(ctx context.Context, nextURL string) (Accounts, error)
}

type AccountService struct {
}

type BrokerService struct {
	Client Broker
}

type BrokerImpl struct {
	AccountService Account
}

func NewBrokerService(apiKey, apiSecret string) *BrokerService {

	logger, _ = zap.NewProduction()

	env = &Environment{
		APIKey:    apiKey,
		APISecret: apiSecret,
	}

	return &BrokerService{
		Client: &BrokerImpl{
			AccountService: &AccountService{},
		},
	}
}

func (b *BrokerImpl) GetAccounts() (uAcc []GenericAccount, err error) {
	logger.Info("GetAccounts.... ")
	ctx := context.Background()

	out, errs := b.AccountService.getPaginatedAccountResults(ctx, GET_ACCOUNTS_PATH)

Loop:
	for {
		select {
		case acc, ok := <-out: // using two value receive to detect when channel is closed
			{
				if ok {
					uAcc = append(uAcc, acc)
				} else {
					break Loop
				}
			}
		case r, ok := <-errs:
			{
				if ok {
					if r != nil {
						return nil, r
					}
				} else {
					break Loop
				}
			}
		default:
		}
	}
	return
}

func (a *AccountService) getPaginatedAccountResults(ctx context.Context, nextUrl string) (<-chan GenericAccount, <-chan error) {
	out := make(chan GenericAccount)
	errs := make(chan error, 1)
	logger.Info("getPaginatedAccountResults.... ")
	go func() {
		defer close(out)
		defer close(errs)
		for {
			accounts, err := a.getAccountByURI(ctx, nextUrl)
			if err != nil {
				errs <- fmt.Errorf("getPaginatedAccountResults aborted unexpectedly : %s", err)
				return
			}
			for _, account := range accounts.Data {
				out <- GenericAccount{
					Name:     account.Name,
					ID:       account.ID,
					Currency: account.Currency,
					Amount:   account.Balance.Amount,
					Type:     account.Type,
					Primary:  account.Primary,
				}
			}
			if accounts.Pagination.NextURI == nil {
				break
			}
			nextUrl = fmt.Sprintf("%v", accounts.Pagination.NextURI)
		}
	}()
	logger.Info("Finished pulling account details...")
	return out, errs
}

func (a *AccountService) getAccountByURI(ctx context.Context, nextURL string) (Accounts, error) {

	logger.Info("getAccountByURI", zap.String("url", nextURL))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var accounts Accounts
	message := Message{
		Method: http.MethodGet,
		Path:   nextURL,
		Body:   "",
		URL:    fmt.Sprintf("%s%s", BASE_URL, nextURL),
	}

	res, err := NewClient().NewRequest(ctx, message, env.APIKey, env.APISecret)
	if err != nil {
		logger.Error("GetAccountByURI request failure.. ", zap.String("url", nextURL), zap.Error(err))
		return accounts, err
	}

	if err := json.Unmarshal(res, &accounts); err != nil { // Parse []byte to go struct pointer
		logger.Error("GetAccountByURI: Failed to parse response: ", zap.String("url", nextURL), zap.Error(err))
		return Accounts{}, fmt.Errorf("GetAccountByURI: Failed to parse response %g", err)
	}
	return accounts, nil
}

func (b *BrokerImpl) GetAccountActivity(ctx context.Context, acc []GenericAccount) (t []AccountActivity) {

	var tasks []*Task
	for _, account := range acc {
		tasks = append(tasks, NewTask(getPaginatedBuyOrders, account))
		tasks = append(tasks, NewTask(getPaginatedSellOrders, account))
	}
	wp := NewWorkerPool(3)
	wp.GenerateFrom(tasks)
	go wp.Run(ctx)

	for {
		select {
		case r, ok := <-wp.ResultSet():
			if !ok {
				return
			}
			t = append(t, r)
		default:
		}
	}
	return
}

// function should return the wallets and allow some calling entity to look after the actual printing
func getPaginatedBuyOrders(ctx context.Context, wallet GenericAccount) (AccountActivity, error) {

	var t AccountActivity
	logger.Info("Executing getPaginatedBuyOrders for account: ", zap.String("Wallet ID", wallet.ID))
	t.Wallet = wallet
	var allBuys []GenericBuy

	buildUrl := func(id string) string {
		return fmt.Sprintf("/v2/accounts/%s/buys", id)
	}
	nextUrl := buildUrl(wallet.ID) // first time per wallet this URL is manually constructed, if pagination in place the value is updated based on responses
	for {
		buys, err := getBuyOrderByURI(ctx, nextUrl)

		if err != nil {
			logger.Error("Failure calling getBuyOrderByURI: ", zap.String("Wallet ID", wallet.ID), zap.Error(err))
			return t, err
		}

		logger.Info("getPaginatedBuyOrders: ", zap.Int("Number buy order results", len(buys.Data)))
		for _, b := range buys.Data {
			allBuys = append(allBuys, GenericBuy{
				ID:            b.ID,
				Status:        b.Status,
				BuyQuantity:   b.Amount.Amount,
				BuyTotal:      b.Total.Amount,
				BuyCurrency:   b.Amount.Currency,
				BoughtWith:    b.Total.Currency,
				FeeCurrency:   b.Subtotal.Currency,
				Resource:      b.Resource,
				Committed:     strconv.FormatBool(b.Committed),
				CreatedAt:     b.CreatedAt,
				TransactionID: b.Transaction.ID,
			})
		}

		if buys.Pagination.NextURI == nil {
			break
		}
		nextUrl = fmt.Sprintf("%v", buys.Pagination.NextURI)
	}
	t.BuyOrders = allBuys
	t.Type = BUY
	return t, nil
}

// function should return the wallets and allow some calling entity to look after the actual printing
func getPaginatedSellOrders(ctx context.Context, wallet GenericAccount) (AccountActivity, error) {

	var t AccountActivity
	logger.Info("Executing getPaginatedSellOrders for account: ", zap.String("Wallet ID", wallet.ID))
	t.Wallet = wallet
	var allSales []GenericSell

	buildUrl := func(id string) string {
		return fmt.Sprintf("/v2/accounts/%s/sells", id)
	}
	nextUrl := buildUrl(wallet.ID) // first time per wallet this URL is manually constructed, if pagination in place the value is updated based on responses

	for {
		sells, err := getSellOrderByURI(ctx, nextUrl)

		if err != nil {
			logger.Error("Failure calling getSellOrderByURI: ", zap.Error(err))
			return t, err
		}

		logger.Info("Number sell order results", zap.Int("Count", len(sells.Data)))
		for _, b := range sells.Data {
			allSales = append(allSales, GenericSell{
				ID:            b.ID,
				Status:        b.Status,
				SellQuantity:  b.Amount.Amount,
				SellTotal:     b.Total.Amount,
				SellCurrency:  b.Amount.Currency,
				SoldWith:      b.Total.Currency,
				FeeCurrency:   b.Subtotal.Currency,
				Resource:      b.Resource,
				Committed:     strconv.FormatBool(b.Committed),
				CreatedAt:     b.CreatedAt,
				TransactionID: b.Transaction.ID,
			})
		}

		if sells.Pagination.NextURI == nil {
			break
		}
		nextUrl = fmt.Sprintf("%v", sells.Pagination.NextURI)
	}

	t.SellOrders = allSales
	t.Type = SELL
	return t, nil
}

func getBuyOrderByURI(ctx context.Context, nextURL string) (Buys, error) {

	logger.Info("Executing getBuyOrderByURI for URL ", zap.String("url", nextURL))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var buys Buys

	m := Message{
		Method: http.MethodGet,
		Path:   nextURL,
		Body:   "",
		URL:    fmt.Sprintf("%s%s", BASE_URL, nextURL),
	}

	c := NewClient()
	res, err := c.NewRequest(ctx, m, env.APIKey, env.APISecret)
	if err != nil {
		logger.Error("getBuyOrderByURI request failure.. ", zap.Error(err))
		return Buys{}, err
	}
	if err := json.Unmarshal(res, &buys); err != nil { // Parse []byte to go struct pointer
		logger.Error("Can not unmarshal JSON", zap.Error(err))
	}
	return buys, nil
}

func getSellOrderByURI(ctx context.Context, nextURL string) (Sells, error) {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var sellOrders Sells

	m := Message{
		Method: http.MethodGet,
		Path:   nextURL,
		Body:   "",
		URL:    fmt.Sprintf("%s%s", BASE_URL, nextURL),
	}

	res, err := NewClient().NewRequest(ctx, m, env.APIKey, env.APISecret)
	if err != nil {
		return Sells{}, fmt.Errorf("getSellOrderByURI failed Sending request : %s", err)
	}

	if err := json.Unmarshal(res, &sellOrders); err != nil { // Parse []byte to go struct pointer
		return sellOrders, fmt.Errorf("getSellOrderByURI failed unmarshalling response : %s", err)
	}
	logger.Info("Number sell orders returned for token ... ", zap.String("url", nextURL), zap.Int("Sell Orders", len(sellOrders.Data)))

	return sellOrders, nil
}

// TODO
func (b *BrokerImpl) GetTransactions(nextUrl string) {}
