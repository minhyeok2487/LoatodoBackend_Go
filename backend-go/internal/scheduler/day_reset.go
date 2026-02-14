package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"math"
)

// resetDayTodo performs all daily reset tasks.
func (s *Scheduler) resetDayTodo(ctx context.Context) error {
	slog.Info("===== daily todo reset started =====")

	safeExecute("updateDayContentGauge", func() (int64, error) {
		return s.updateDayContentGauge(ctx)
	})

	safeExecute("saveBeforeGauge", func() (int64, error) {
		return s.saveBeforeGauge(ctx)
	})

	safeExecute("updateDayContentCheck", func() (int64, error) {
		return s.updateDayContentCheck(ctx)
	})

	safeExecute("updateDayTodoGold", func() (int64, error) {
		return 0, s.updateDayTodoGold(ctx)
	})

	safeExecute("updateCustomDailyTodo", func() (int64, error) {
		return s.updateCustomDailyTodo(ctx)
	})

	safeExecute("resetServerTodoState", func() (int64, error) {
		return s.resetServerTodoState(ctx)
	})

	slog.Info("===== daily todo reset completed =====")
	return nil
}

// updateDayContentGauge updates the gauge values based on check counts.
// Gauge increases by 10 for each check (max 2 checks per content).
// chaos: check*10, guardian: check*10, epona: check*10
func (s *Scheduler) updateDayContentGauge(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_gauge = LEAST(chaos_gauge + chaos_check * 10, 100),
			guardian_gauge = LEAST(guardian_gauge + guardian_check * 10, 100),
			epona_gauge = LEAST(epona_gauge + epona_check2 * 10, 100)
		 WHERE (is_deleted IS NULL OR is_deleted = false)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// saveBeforeGauge stores the current gauge values as before_gauge for tracking.
func (s *Scheduler) saveBeforeGauge(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET
			before_epona_gauge = epona_gauge,
			before_chaos_gauge = chaos_gauge,
			before_guardian_gauge = guardian_gauge
		 WHERE (is_deleted IS NULL OR is_deleted = false)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// updateDayContentCheck resets all daily check values to 0.
func (s *Scheduler) updateDayContentCheck(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET
			chaos_check = 0,
			guardian_check = 0,
			epona_check2 = 0
		 WHERE (is_deleted IS NULL OR is_deleted = false)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// updateDayTodoGold recalculates guardian gold based on current market prices.
func (s *Scheduler) updateDayTodoGold(ctx context.Context) error {
	// Fetch jewelry prices from market table
	type jewelryPrice struct {
		name        string
		recentPrice int
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT name, recent_price FROM market WHERE name IN ('3티어 1레벨 보석', '4티어 1레벨 보석')`)
	if err != nil {
		return err
	}
	defer rows.Close()

	prices := make(map[string]int)
	for rows.Next() {
		var j jewelryPrice
		if err := rows.Scan(&j.name, &j.recentPrice); err != nil {
			return err
		}
		prices[j.name] = j.recentPrice
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Get guardian day contents and update gold
	contentRows, err := s.db.QueryContext(ctx,
		`SELECT content_id, level, jewelry FROM content
		 WHERE dtype = 'DayContent' AND category = '가디언토벌'`)
	if err != nil {
		return err
	}
	defer contentRows.Close()

	for contentRows.Next() {
		var contentID int64
		var level float64
		var jewelry float64
		if err := contentRows.Scan(&contentID, &level, &jewelry); err != nil {
			return err
		}

		// Determine which jewelry price to use based on item level
		var priceName string
		if level >= 1415 && level < 1640 {
			priceName = "3티어 1레벨 보석"
		} else {
			priceName = "4티어 1레벨 보석"
		}

		price, ok := prices[priceName]
		if !ok {
			continue
		}

		gold := math.Round(float64(price)*jewelry*100) / 100

		_, err = s.db.ExecContext(ctx,
			`UPDATE characters SET guardian_gold = ?
			 WHERE guardian_id = ? AND (is_deleted IS NULL OR is_deleted = false)`,
			gold, contentID)
		if err != nil {
			return err
		}
	}

	return contentRows.Err()
}

// updateCustomDailyTodo resets all daily custom todo checks to unchecked.
func (s *Scheduler) updateCustomDailyTodo(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = false WHERE frequency = 'DAILY'`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// resetServerTodoState resets all server todo state checks.
func (s *Scheduler) resetServerTodoState(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE server_todo_state SET checked = false`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// safeExecute runs a task and logs the result, catching any errors.
func safeExecute(name string, fn func() (int64, error)) {
	count, err := fn()
	if err != nil {
		slog.Error("task failed", "task", name, "error", err)
		return
	}
	if count > 0 {
		slog.Info("task completed", "task", name, "affected", count)
	} else {
		slog.Info("task completed", "task", name)
	}
}

// checkScheduleRaids auto-checks schedule raids that have passed.
func (s *Scheduler) checkScheduleRaids(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE schedule SET checked = true
		 WHERE auto_check = true AND checked = false
		   AND repeat_week = false
		   AND date IS NOT NULL
		   AND CONCAT(date, ' ', COALESCE(time, '00:00')) < NOW()`)
	if err != nil {
		return err
	}
	return nil
}

// addEnergyToAllLifeEnergies adds energy to all life energy records.
// Each 30-minute tick adds energy (30 base, or 45 with beatrice buff).
func (s *Scheduler) addEnergyToAllLifeEnergies(ctx context.Context) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE life_energy SET
			energy = LEAST(
				energy + CASE WHEN beatrice = true THEN 45 ELSE 30 END,
				max_energy
			),
			last_modified_date = NOW()`)
	if err != nil {
		return err
	}

	count, _ := result.RowsAffected()
	slog.Info("life energy updated", "count", count)
	return nil
}

// Helper for nullable int64 scanning
func scanNullInt64(n sql.NullInt64) int64 {
	if n.Valid {
		return n.Int64
	}
	return 0
}
