package client

import (
	"encoding/json"
	"fmt"
)

// CategoryCode constants matching the Java enum.
const (
	CategoryCodeMarketMaterials = 50000  // 재련재료
	CategoryCodeGems            = 210000 // 보석
)

// MarketSearchRequest is the request body for the market items API.
type MarketSearchRequest struct {
	Sort          string `json:"Sort"`
	CategoryCode  int    `json:"CategoryCode"`
	PageNo        int    `json:"PageNo"`
	SortCondition string `json:"SortCondition"`
}

// MarketSearchResponse is the response from the market items API.
type MarketSearchResponse struct {
	PageNo     int          `json:"PageNo"`
	PageSize   int          `json:"PageSize"`
	TotalCount int          `json:"TotalCount"`
	Items      []MarketItem `json:"Items"`
}

// MarketItem represents a single item from the market search results.
type MarketItem struct {
	ID              int64   `json:"Id"`
	Name            string  `json:"Name"`
	Grade           string  `json:"Grade"`
	Icon            string  `json:"Icon"`
	BundleCount     int     `json:"BundleCount"`
	TradeRemainCount *int   `json:"TradeRemainCount"`
	YDayAvgPrice    float64 `json:"YDayAvgPrice"`
	RecentPrice     int     `json:"RecentPrice"`
	CurrentMinPrice int     `json:"CurrentMinPrice"`
}

// AuctionSearchRequest is the request body for the auction items API.
type AuctionSearchRequest struct {
	Sort          string `json:"Sort"`
	CategoryCode  int    `json:"CategoryCode"`
	ItemTier      int    `json:"ItemTier"`
	ItemName      string `json:"ItemName"`
	PageNo        int    `json:"PageNo"`
	SortCondition string `json:"SortCondition"`
}

// AuctionSearchResponse is the response from the auction items API.
type AuctionSearchResponse struct {
	PageNo     int           `json:"PageNo"`
	PageSize   int           `json:"PageSize"`
	TotalCount int           `json:"TotalCount"`
	Items      []AuctionItem `json:"Items"`
}

// AuctionItem represents a single item from the auction search results.
type AuctionItem struct {
	Name            string  `json:"Name"`
	Grade           string  `json:"Grade"`
	Tier            int     `json:"Tier"`
	Icon            string  `json:"Icon"`
	AuctionInfo     *AuctionInfo `json:"AuctionInfo"`
}

// AuctionInfo holds the auction-specific info for an auction item.
type AuctionInfo struct {
	StartPrice   int    `json:"StartPrice"`
	BuyPrice     int    `json:"BuyPrice"`
	BidPrice     int    `json:"BidPrice"`
	EndDate      string `json:"EndDate"`
	BidCount     int    `json:"BidCount"`
	BidStartPrice int   `json:"BidStartPrice"`
	IsCompetitive bool  `json:"IsCompetitive"`
	TradeAllowCount int `json:"TradeAllowCount"`
}

// LostarkMarketClient handles market and auction API calls.
type LostarkMarketClient struct {
	client *LostarkClient
}

// NewLostarkMarketClient creates a new LostarkMarketClient.
func NewLostarkMarketClient(client *LostarkClient) *LostarkMarketClient {
	return &LostarkMarketClient{client: client}
}

// GetMarketData fetches all market items for the given category code, paginating through all pages.
// POST https://developer-lostark.game.onstove.com/markets/items
func (c *LostarkMarketClient) GetMarketData(categoryCode int, apiKey string) ([]MarketItem, error) {
	const apiURL = "https://developer-lostark.game.onstove.com/markets/items"
	const pageSize = 10

	// First request to get total count
	firstReq := MarketSearchRequest{
		Sort:          "RECENT_PRICE",
		CategoryCode:  categoryCode,
		PageNo:        0,
		SortCondition: "DESC",
	}

	body, err := c.client.Post(apiURL, firstReq, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetching market data page 0: %w", err)
	}

	var firstResp MarketSearchResponse
	if err := json.Unmarshal(body, &firstResp); err != nil {
		return nil, fmt.Errorf("parsing market response: %w", err)
	}

	totalCount := firstResp.TotalCount
	var allItems []MarketItem

	// Paginate through all pages (pageNo starts at 1 for actual data)
	pageNo := 0
	for pageNo*pageSize < totalCount {
		pageNo++
		req := MarketSearchRequest{
			Sort:          "RECENT_PRICE",
			CategoryCode:  categoryCode,
			PageNo:        pageNo,
			SortCondition: "DESC",
		}

		body, err := c.client.Post(apiURL, req, apiKey)
		if err != nil {
			return nil, fmt.Errorf("fetching market data page %d: %w", pageNo, err)
		}

		var resp MarketSearchResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parsing market response page %d: %w", pageNo, err)
		}

		allItems = append(allItems, resp.Items...)
	}

	return allItems, nil
}

// AuctionRequestDto holds the parameters for an auction item search.
type AuctionRequestDto struct {
	ItemTier     int
	CategoryCode int
	ItemName     string
}

// GetAuctionItems fetches the cheapest auction item matching the given criteria.
// POST https://developer-lostark.game.onstove.com/auctions/items
func (c *LostarkMarketClient) GetAuctionItems(request AuctionRequestDto, apiKey string) (*AuctionItem, error) {
	const apiURL = "https://developer-lostark.game.onstove.com/auctions/items"

	req := AuctionSearchRequest{
		Sort:          "BUY_PRICE",
		CategoryCode:  request.CategoryCode,
		ItemTier:      request.ItemTier,
		ItemName:      request.ItemName,
		PageNo:        1,
		SortCondition: "ASC",
	}

	body, err := c.client.Post(apiURL, req, apiKey)
	if err != nil {
		return nil, fmt.Errorf("fetching auction data: %w", err)
	}

	var resp AuctionSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing auction response: %w", err)
	}

	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no auction items found for %s", request.ItemName)
	}

	return &resp.Items[0], nil
}
