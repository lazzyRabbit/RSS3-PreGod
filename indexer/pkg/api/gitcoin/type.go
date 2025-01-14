package gitcoin

import (
	"fmt"
	"math/big"
	"os"
	"time"
)

type GitcoinPlatform string

const (
	Unknown GitcoinPlatform = "unknown"

	ETH     GitcoinPlatform = "eth"
	Polygon GitcoinPlatform = "polygon"
	ZKSYNC  GitcoinPlatform = "zksync"
)

func (p GitcoinPlatform) getContractAddress() string {
	if p == ETH {
		return bulkCheckoutAddressETH
	}

	if p == Polygon {
		return bulkCheckoutAddressPolygon
	}

	return ""
}

type crawlerConfig struct {
	FromHeight    int64
	Step          int64
	MinStep       int64
	Confirmations int64
	SleepInterval time.Duration
	NextRoundTime time.Time
	Interrupt     chan os.Signal
}

var DefaultEthConfig = &crawlerConfig{
	FromHeight:    10245999, // gitcoin bulkCheckout contract was created at block #10245999
	Step:          50,
	MinStep:       10,
	Confirmations: 15,
	SleepInterval: 600 * time.Second,
	NextRoundTime: time.Now(),
	Interrupt:     make(chan os.Signal, 1),
}

var DefaultPolygonConfig = &crawlerConfig{
	FromHeight:    18682002, // gitcoin bulkCheckout contract was created at block #10245999
	Step:          50,
	MinStep:       10,
	Confirmations: 120,
	SleepInterval: 600 * time.Second,
	NextRoundTime: time.Now(),
	Interrupt:     make(chan os.Signal, 1),
}

var DefaultZksyncConfig = &crawlerConfig{
	FromHeight:    2600,
	Step:          50,
	MinStep:       10,
	Confirmations: 15,
	SleepInterval: 600 * time.Second,
	NextRoundTime: time.Now(),
	Interrupt:     make(chan os.Signal, 1),
}

type DonationApproach string

const (
	DonationApproachEthereum = "Standard"
	DonationApproachPolygon  = "Polygon"
	DonationApproachZksync   = "zkSync"
)

type GrantInfo struct {
	Title        string
	AdminAddress string
}

type ProjectInfo struct {
	Active          bool
	Id              int64
	Title           string
	Slug            string
	Description     string
	ReferUrl        string
	Logo            string
	AdminAddress    string
	TokenAddress    string
	TokenSymbol     string
	ContractAddress string
	Network         string
}

type DonationInfo struct {
	Donor          string
	AdminAddress   string
	TokenAddress   string
	Amount         string
	Symbol         string
	FormatedAmount *big.Int
	Decimals       int64
	Timestamp      string
	TxHash         string
	Approach       DonationApproach
}

func (d DonationInfo) String() string {
	return fmt.Sprintf(`Donor: %s, AdminAddress: %s, TokenAddress: %s, Amount: %s, Symbol: %s, TxHash: %s`,
		d.Donor, d.AdminAddress, d.TokenAddress, d.Amount, d.Symbol, d.TxHash)
}

func (d DonationInfo) GetTxTo() string {
	if d.Approach == DonationApproachEthereum {
		return bulkCheckoutAddressETH
	}

	if d.Approach == DonationApproachPolygon {
		return bulkCheckoutAddressPolygon
	}

	if d.Approach == DonationApproachZksync {
		return d.AdminAddress
	}

	return ""
}
