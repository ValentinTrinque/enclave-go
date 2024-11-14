package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type FillID string

type FillParams struct {
	StartTime *time.Time
	EndTime   *time.Time
	Market    string
	Limit     int
	Cursor    string
}

func (fp *FillParams) IsEmpty() bool {
	return fp.StartTime == nil && fp.EndTime == nil && fp.Market == "" && fp.Limit == 0 && fp.Cursor == ""
}

func (fp *FillParams) GetFillPathParams() string {
	if fp.IsEmpty() {
		return ""
	}

	pathParams := "?"

	if fp.StartTime != nil {
		pathParams += fmt.Sprintf("startTime=%d&", fp.StartTime.UnixMilli())
	}

	if fp.EndTime != nil {
		pathParams += fmt.Sprintf("endTime=%d&", fp.EndTime.UnixMilli())
	}

	if fp.Market != "" {
		pathParams += fmt.Sprintf("market=%s&", fp.Market)
	}

	if fp.Limit > 0 {
		pathParams += fmt.Sprintf("limit=%d&", fp.Limit)
	}

	if fp.Cursor != "" {
		pathParams += fmt.Sprintf("cursor=%s&", fp.Cursor)
	}

	pathParams = strings.TrimSuffix(pathParams, "&")

	return pathParams
}

type ApiFill struct {
	FillID        FillID           `json:"id"`
	OrderID       OrderID          `json:"orderId"`
	ClientOrderID OrderID          `json:"clientOrderId,omitempty"`
	Market        Market           `json:"market"`
	Price         decimal.Decimal  `json:"price"`
	Size          decimal.Decimal  `json:"size"` // size of fill (base currency)
	Side          BidAsk           `json:"side"`
	Cost          decimal.Decimal  `json:"filledCost"` // total cost of fill (quote currency)
	Fee           decimal.Decimal  `json:"fee"`
	FeeRebate     *decimal.Decimal `json:"feeRebate,omitempty"`
	CreatedAt     time.Time        `json:"time"`
	ADL           *bool            `json:"isADL,omitempty"`
}
