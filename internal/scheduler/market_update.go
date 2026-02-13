package scheduler

import (
	"context"
	"log/slog"
	"time"

	"lostark-todo-backend/internal/client"
)

// updateMarketData fetches current market and auction data from the Lostark API
// and upserts it into the market table.
func (s *Scheduler) updateMarketData(ctx context.Context) error {
	slog.Info("===== market data update started =====")

	// Fetch market items for category 50000 (재련재료)
	if err := s.updateMarketItems(ctx); err != nil {
		slog.Error("failed to update market items", "error", err)
		s.discord.SendError("거래소 데이터 갱신 실패: " + err.Error())
	}

	// Fetch auction items for tier 3 & 4 gems
	if err := s.updateAuctionItems(ctx); err != nil {
		slog.Error("failed to update auction items", "error", err)
		s.discord.SendError("경매장 데이터 갱신 실패: " + err.Error())
	}

	slog.Info("===== market data update completed =====")
	return nil
}

// updateMarketItems fetches all market items for 재련재료 (category 50000) and upserts them.
func (s *Scheduler) updateMarketItems(ctx context.Context) error {
	items, err := s.marketClient.GetMarketData(client.CategoryCodeMarketMaterials, s.apiKey)
	if err != nil {
		return err
	}

	slog.Info("fetched market items", "count", len(items))

	now := time.Now()
	for _, item := range items {
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO market (lostark_market_id, name, grade, icon, bundle_count,
				y_day_avg_price, recent_price, current_min_price, category_code,
				created_date, last_modified_date)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
				name = VALUES(name),
				grade = VALUES(grade),
				icon = VALUES(icon),
				bundle_count = VALUES(bundle_count),
				y_day_avg_price = VALUES(y_day_avg_price),
				recent_price = VALUES(recent_price),
				current_min_price = VALUES(current_min_price),
				last_modified_date = VALUES(last_modified_date)`,
			item.ID, item.Name, item.Grade, item.Icon, item.BundleCount,
			item.YDayAvgPrice, item.RecentPrice, item.CurrentMinPrice,
			client.CategoryCodeMarketMaterials, now, now,
		)
		if err != nil {
			slog.Error("failed to upsert market item", "name", item.Name, "error", err)
		}
	}

	return nil
}

// updateAuctionItems fetches the cheapest tier 3 and tier 4 level-5 gems from auction
// and upserts them into the market table.
func (s *Scheduler) updateAuctionItems(ctx context.Context) error {
	auctionRequests := []client.AuctionRequestDto{
		{ItemTier: 3, CategoryCode: client.CategoryCodeGems, ItemName: "5레벨"},
		{ItemTier: 4, CategoryCode: client.CategoryCodeGems, ItemName: "5레벨"},
	}

	gemNames := []string{"3티어 1레벨 보석", "4티어 1레벨 보석"}

	now := time.Now()
	for i, req := range auctionRequests {
		item, err := s.marketClient.GetAuctionItems(req, s.apiKey)
		if err != nil {
			slog.Error("failed to fetch auction item", "tier", req.ItemTier, "error", err)
			continue
		}

		buyPrice := 0
		if item.AuctionInfo != nil {
			buyPrice = item.AuctionInfo.BuyPrice
		}

		_, err = s.db.ExecContext(ctx,
			`INSERT INTO market (lostark_market_id, name, grade, icon, bundle_count,
				y_day_avg_price, recent_price, current_min_price, category_code,
				created_date, last_modified_date)
			 VALUES (0, ?, ?, ?, 1, 0, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
				grade = VALUES(grade),
				icon = VALUES(icon),
				recent_price = VALUES(recent_price),
				current_min_price = VALUES(current_min_price),
				last_modified_date = VALUES(last_modified_date)`,
			gemNames[i], item.Grade, item.Icon, buyPrice, buyPrice,
			client.CategoryCodeGems, now, now,
		)
		if err != nil {
			slog.Error("failed to upsert auction item", "name", gemNames[i], "error", err)
		}
	}

	return nil
}
