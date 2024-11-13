package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type OrderID string

type AddOrderReq struct {
	Side      BidAsk          `json:"side"`
	Price     decimal.Decimal `json:"price"`
	Size      decimal.Decimal `json:"size"`
	QuoteSize decimal.Decimal `json:"quoteSize"`
	Market    Market          `json:"market"`

	ClientOrderID OrderID          `json:"clientOrderId,omitempty"`
	Type          OrderType        `json:"type"`
	TimeInForce   OrderTimeInForce `json:"timeInForce,omitempty"`
	ReduceOnly    bool             `json:"reduceOnly,omitempty"`

	PostOnly bool `json:"postOnly,omitempty"`
}

type BidAsk bool

const Bid BidAsk = true
const Ask BidAsk = false

func (b BidAsk) Opposite() BidAsk {
	return !b
}

func (b BidAsk) String() string {
	if b {
		return "buy"
	} else {
		return "sell"
	}
}

func (b *BidAsk) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case `"buy"`:
		*b = Bid
	case `"sell"`:
		*b = Ask
	default:
		return fmt.Errorf("invalid bid/ask: %s", string(data))
	}
	return nil
}

func (b BidAsk) MarshalJSON() ([]byte, error) {
	if b {
		return []byte(`"buy"`), nil
	} else {
		return []byte(`"sell"`), nil
	}
}

type OrderType bool

const OrderTypeLimit OrderType = false
const OrderTypeMarket OrderType = true

func (o OrderType) String() string {
	if o {
		return "market"
	} else {
		return "limit"
	}
}

func (o OrderType) MarshalJSON() ([]byte, error) {
	if o {
		return []byte(`"market"`), nil
	} else {
		return []byte(`"limit"`), nil
	}
}

func (o *OrderType) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case `"market"`:
		*o = OrderTypeMarket
	case `"limit"`:
		*o = OrderTypeLimit
	default:
		return fmt.Errorf("invalid order type: %s", string(data))
	}
	return nil
}

type OrderTimeInForce string

const OrderTimeInForceGoodUntilCancelled OrderTimeInForce = "GTC"
const OrderTimeInForceImmediateOrCancel OrderTimeInForce = "IOC"

type ApiOrder struct {
	OrderID        OrderID          `json:"orderId"`
	ClientOrderID  OrderID          `json:"clientOrderId,omitempty"`
	Side           BidAsk           `json:"side"`
	Price          decimal.Decimal  `json:"price"`
	OrderQuantity  decimal.Decimal  `json:"size"`
	Market         Market           `json:"market"`
	FilledQuantity decimal.Decimal  `json:"filledSize"`
	FilledCost     decimal.Decimal  `json:"filledCost"`
	Fee            decimal.Decimal  `json:"fee"`
	FeeRebate      *decimal.Decimal `json:"feeRebate,omitempty"`
	State          OrderState       `json:"status"`
	CreatedAt      time.Time        `json:"createdAt"`
	FilledAt       *time.Time       `json:"filledAt,omitempty"`

	CanceledAt   *time.Time       `json:"canceledAt,omitempty"`
	CancelReason CancelReason     `json:"cancelReason,omitempty"`
	Type         OrderType        `json:"type"`
	TimeInForce  OrderTimeInForce `json:"timeInForce,omitempty"`
	ReduceOnly   bool             `json:"reduceOnly,omitempty"`
}

type CancelReason int

const (
	// This is a default cancel state and is for when an order has been canceled by a user.
	User CancelReason = iota

	// When the cancellation is due to liquidation.
	Liquidation

	// Canceled due to self match prevention.
	SelfMatchPrevention

	// Canceled due to websocket disconnect and cancel on disconnect being enabled.
	CancelAfterTimeout

	// Canceled due to bad price on startup.
	StartupBadPrices

	// Canceled due to being an IOC order and not being able to fill.
	ImmediateOrCancel

	// Canceled due to the monolith shutting down while a user has a subscription to the CancelOnDisconnect websocket
	CancelAfterTimeoutOnShutdown

	// Canceled due to the monolith starting up and the user's id being in the feature flag: "cancel_all_open_orders_on_startup"
	CancelOnStartup

	// Canceled via the admin dashboard.
	CancelByAdmin
)

func (s CancelReason) String() string {
	switch s {
	case User:
		return ""
	case Liquidation:
		return "liquidation"
	case SelfMatchPrevention:
		return "selfMatchPrevention"
	case CancelAfterTimeout:
		return "cancelAfterTimeout"
	case StartupBadPrices:
		return "startupBadPrice"
	case ImmediateOrCancel:
		return "immediateOrCancel"
	case CancelAfterTimeoutOnShutdown:
		return "cancelAfterTimeoutOnShutdown"
	case CancelOnStartup:
		return "cancelOnStartup"
	case CancelByAdmin:
		return "cancelByAdmin"
	default:
		panic(fmt.Sprintf("invalid CancelReason: %d", s))
	}
}

func (s CancelReason) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, strings.ToLower(s.String()))), nil
}

func (s *CancelReason) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case "":
		*s = User
	case `"liquidation"`:
		*s = Liquidation
	case `"selfmatchprevention"`:
		*s = SelfMatchPrevention
	case `"cancelaftertimeout"`:
		*s = CancelAfterTimeout
	case `"startupbadprice"`:
		*s = StartupBadPrices
	case `"immediateorcancel"`:
		*s = ImmediateOrCancel
	case `"cancelaftertimeoutonshutdown"`:
		*s = CancelAfterTimeoutOnShutdown
	case `"cancelonstartup"`:
		*s = CancelOnStartup
	case `"cancelbyadmin"`:
		*s = CancelByAdmin
	default:
		return fmt.Errorf("invalid CancelReason: %s", string(data))
	}
	return nil
}

// swagger:type string
type OrderState int

const (
	// This is the default state and is for when an order is created but hasn't been added to the matching engine
	// An order with this state should never be visible to a user
	New OrderState = iota

	// After accepted by the matching engine
	Open

	// The order is closed because it was fully filled
	FullyFilled

	// The order was canceled.  It may or may not have been partially filled
	Canceled

	// If the matching engine rejects a cancel request.  This should never happen because there is validation before
	// sending the cancel to the matching engine
	CancelRejected

	// Not used
	Rejected
)

func (s OrderState) String() string {
	switch s {
	case New:
		return "new"
	case Open:
		return "open"
	case FullyFilled:
		return "fullyFilled"
	case Canceled:
		return "canceled"
	case CancelRejected:
		return "cancelRejected"
	case Rejected:
		return "rejected"
	default:
		return "unknown"
	}
}

var ErrStatusQuery = fmt.Errorf("indicated empty order state for filter")

func OrderStateFromQueryParam(s string) (OrderState, error) {
	switch s {
	case "open":
		return Open, nil
	case "fullyFilled":
		return FullyFilled, nil
	case "canceled":
		return Canceled, nil
	case "":
		return -1, ErrStatusQuery
	default:
		return -1, fmt.Errorf("invalid order state query %s", s)

	}
}

func (s OrderState) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, strings.ToLower(s.String()))), nil
}

func (s *OrderState) UnmarshalJSON(data []byte) error {
	switch strings.ToLower(string(data)) {
	case `"new"`:
		*s = New
	case `"open"`:
		*s = Open
	case `"fullyfilled"`:
		*s = FullyFilled
	case `"canceled"`:
		*s = Canceled
	case `"cancelrejected"`:
		*s = CancelRejected
	case `"rejected"`:
		*s = Rejected
	default:
		return fmt.Errorf("invalid OrderState: %s", string(data))
	}
	return nil
}
