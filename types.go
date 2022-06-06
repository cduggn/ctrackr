package gocoin

import "time"

const (
	BUY          int = 0
	SELL         int  = 1
)

type Accounts struct {
	Pagination Pagination `json:"pagination"`
	Data       []Data     `json:"data"`
}
type Pagination struct {
	EndingBefore  interface{} `json:"ending_before"`
	StartingAfter interface{} `json:"starting_after"`
	Limit         int         `json:"limit"`
	Order         string      `json:"order"`
	PreviousURI   interface{} `json:"previous_uri"`
	NextURI       interface{} `json:"next_uri"`
}
type Balance struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}
type Data struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Primary      bool      `json:"primary"`
	Type         string    `json:"type"`
	Currency     string    `json:"currency"`
	Balance      Balance   `json:"balance"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Resource     string    `json:"resource"`
	ResourcePath string    `json:"resource_path"`
	Ready        bool      `json:"ready,omitempty"`
}

type Buys struct {
	Pagination struct {
		EndingBefore  interface{} `json:"ending_before"`
		StartingAfter interface{} `json:"starting_after"`
		Limit         int         `json:"limit"`
		Order         string      `json:"order"`
		PreviousURI   interface{} `json:"previous_uri"`
		NextURI       interface{} `json:"next_uri"`
	} `json:"pagination"`
	Data []struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PaymentMethod struct {
			ID           string `json:"id"`
			Resource     string `json:"resource"`
			ResourcePath string `json:"resource_path"`
		} `json:"payment_method"`
		Transaction struct {
			ID           string `json:"id"`
			Resource     string `json:"resource"`
			ResourcePath string `json:"resource_path"`
		} `json:"transaction"`
		Amount struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Total struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"total"`
		Subtotal struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"subtotal"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
		Committed    bool   `json:"committed"`
		Instant      bool   `json:"instant"`
		Fee          struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"fee"`
		PayoutAt string `json:"payout_at"`
	} `json:"data"`
}

type Sells struct {
	Pagination struct {
		EndingBefore  interface{} `json:"ending_before"`
		StartingAfter interface{} `json:"starting_after"`
		Limit         int         `json:"limit"`
		Order         string      `json:"order"`
		PreviousURI   interface{} `json:"previous_uri"`
		NextURI       interface{} `json:"next_uri"`
	} `json:"pagination"`
	Data []struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		PaymentMethod struct {
			ID           string `json:"id"`
			Resource     string `json:"resource"`
			ResourcePath string `json:"resource_path"`
		} `json:"payment_method"`
		Transaction struct {
			ID           string `json:"id"`
			Resource     string `json:"resource"`
			ResourcePath string `json:"resource_path"`
		} `json:"transaction"`
		Amount struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"amount"`
		Total struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"total"`
		Subtotal struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"subtotal"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
		Committed    bool   `json:"committed"`
		Instant      bool   `json:"instant"`
		Fee          struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"fee"`
		PayoutAt string `json:"payout_at"`
	} `json:"data"`
}

type GenericAccount struct {
	Name     string
	ID       string
	Currency string
	Amount   string
	Type     string
	Primary  bool
}

type Message struct {
	Method string
	Path   string
	Body   string
	Secret string
	URL    string
	Epoch  int64
}

type AccountActivity struct {
	Wallet       GenericAccount
	ActiveWallet bool
	BuyOrders    []GenericBuy
	Type         int
	SellOrders   []GenericSell
}

type GenericBuy struct {
	ID            string
	Status        string
	BuyQuantity   string
	BoughtWith    string
	BuyCurrency   string
	BuyTotal      string
	Fee           float64
	FeeCurrency   string
	Resource      string
	Committed     string
	CreatedAt     string
	TransactionID string
}

type GenericSell struct {
	ID            string
	Status        string
	SellQuantity  string
	SoldWith      string
	SellCurrency  string
	SellTotal     string
	Fee           float64
	FeeCurrency   string
	Resource      string
	Committed     string
	CreatedAt     string
	TransactionID string
}

