package scheduler

import (
	"context"
	"log/slog"
)

// resetWeekTodo performs all weekly reset tasks (runs Wednesday 06:02 KST).
func (s *Scheduler) resetWeekTodo(ctx context.Context) error {
	slog.Info("===== weekly todo reset started =====")

	safeExecute("updateTwoCycle", func() (int64, error) {
		return s.updateTwoCycle(ctx)
	})

	safeExecute("resetTodoV2CoolTime2", func() (int64, error) {
		return s.resetTodoV2CoolTime2(ctx)
	})

	safeExecute("resetTodoV2", func() (int64, error) {
		return s.resetTodoV2(ctx)
	})

	safeExecute("updateWeekContent", func() (int64, error) {
		return s.updateWeekContent(ctx)
	})

	safeExecute("updateWeekDayTodoTotalGold", func() (int64, error) {
		return s.updateWeekDayTodoTotalGold(ctx)
	})

	safeExecute("updateCustomWeeklyTodo", func() (int64, error) {
		return s.updateCustomWeeklyTodo(ctx)
	})

	safeExecute("deleteAllRaidBusGold", func() (int64, error) {
		return s.deleteAllRaidBusGold(ctx)
	})

	slog.Info("===== weekly todo reset completed =====")
	return nil
}

// updateTwoCycle toggles the two-cycle key-value entry.
// This alternates between "0" and "1" each week to track 2-week raid cycles.
func (s *Scheduler) updateTwoCycle(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE key_value SET value_name = CASE
			WHEN value_name = '0' THEN '1'
			ELSE '0'
		 END
		 WHERE key_name = 'twoCycle'`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// resetTodoV2CoolTime2 handles 2-cycle raids.
// When cool_time=2 and twoCycle toggles, these raids should be treated differently.
// If the current twoCycle value matches the initial state, reset; otherwise skip.
func (s *Scheduler) resetTodoV2CoolTime2(ctx context.Context) (int64, error) {
	// Get current two-cycle value
	var twoCycleValue string
	err := s.db.QueryRowContext(ctx,
		`SELECT value_name FROM key_value WHERE key_name = 'twoCycle'`).Scan(&twoCycleValue)
	if err != nil {
		return 0, err
	}

	// For 2-cycle raids: if twoCycle is now "1", it means this is the "off" week.
	// Reset checked status but keep the cool_time tracking.
	if twoCycleValue == "1" {
		result, err := s.db.ExecContext(ctx,
			`UPDATE todov2 SET is_checked = false
			 WHERE cool_time = 2`)
		if err != nil {
			return 0, err
		}
		return result.RowsAffected()
	}

	return 0, nil
}

// resetTodoV2 resets all TodoV2 checks and cool_time for the new week.
func (s *Scheduler) resetTodoV2(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE todov2 SET is_checked = false, cool_time = 0
		 WHERE cool_time != 2`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// updateWeekContent resets weekly content counters:
// - week_epona to 0
// - silmael_change to false
// - cube_ticket to 0
// - elysian_count to 0
func (s *Scheduler) updateWeekContent(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET
			week_epona = 0,
			silmael_change = false,
			cube_ticket = 0,
			elysian_count = 0
		 WHERE (is_deleted IS NULL OR is_deleted = false)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// updateWeekDayTodoTotalGold resets the weekly accumulated gold total to 0.
func (s *Scheduler) updateWeekDayTodoTotalGold(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters SET week_total_gold = 0
		 WHERE (is_deleted IS NULL OR is_deleted = false)`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// updateCustomWeeklyTodo resets all weekly custom todo checks to unchecked.
func (s *Scheduler) updateCustomWeeklyTodo(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`UPDATE custom_todo SET is_checked = false WHERE frequency = 'WEEKLY'`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// deleteAllRaidBusGold clears all bus gold entries for the new week.
func (s *Scheduler) deleteAllRaidBusGold(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM raid_bus_gold WHERE fixed = false`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
