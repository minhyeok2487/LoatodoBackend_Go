package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"lostark-todo-backend/internal/client"
	"lostark-todo-backend/internal/webhook"
)

// Scheduler manages all periodic tasks using cron.
type Scheduler struct {
	cron           *cron.Cron
	shedLock       *ShedLock
	db             *sql.DB
	lostarkClient  *client.LostarkClient
	marketClient   *client.LostarkMarketClient
	discord        *webhook.DiscordWebhook
	apiKey         string
}

// NewScheduler creates a new Scheduler with Asia/Seoul timezone.
func NewScheduler(
	shedLock *ShedLock,
	db *sql.DB,
	lostarkClient *client.LostarkClient,
	marketClient *client.LostarkMarketClient,
	discord *webhook.DiscordWebhook,
	apiKey string,
) *Scheduler {
	loc, _ := time.LoadLocation("Asia/Seoul")
	c := cron.New(
		cron.WithLocation(loc),
		cron.WithSeconds(),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	return &Scheduler{
		cron:          c,
		shedLock:      shedLock,
		db:            db,
		lostarkClient: lostarkClient,
		marketClient:  marketClient,
		discord:       discord,
		apiKey:        apiKey,
	}
}

// Start registers all cron jobs and starts the scheduler.
func (s *Scheduler) Start() {
	// Daily 01:00 - Update market data
	s.addJob("0 0 1 * * *", "updateMarketData", 30*time.Minute, 1*time.Minute, s.updateMarketData)

	// Daily 06:00 - Reset day todo
	s.addJob("0 0 6 * * *", "resetDayTodo", 30*time.Minute, 1*time.Minute, s.resetDayTodo)

	// Wednesday 06:02 - Reset week todo
	s.addJob("0 2 6 * * 3", "resetWeekTodo", 30*time.Minute, 1*time.Minute, s.resetWeekTodo)

	// Every 10 minutes - Check schedule raids
	s.addJob("0 */10 * * * *", "checkScheduleRaids", 9*time.Minute, 1*time.Minute, s.checkScheduleRaids)

	// Every 30 minutes (at :05 and :35) - Add energy to all life energies
	s.addJob("0 5,35 * * * *", "addEnergyToAllLifeEnergies", 29*time.Minute, 1*time.Minute, s.addEnergyToAllLifeEnergies)

	s.cron.Start()
	slog.Info("scheduler started")
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	slog.Info("scheduler stopped")
}

// addJob registers a new cron job with ShedLock distributed locking.
func (s *Scheduler) addJob(spec, lockName string, lockAtMost, lockAtLeast time.Duration, fn func(ctx context.Context) error) {
	_, err := s.cron.AddFunc(spec, func() {
		ctx := context.Background()
		err := s.shedLock.RunWithLock(ctx, lockName, lockAtMost, lockAtLeast, fn)
		if err != nil {
			slog.Error("scheduled task failed", "task", lockName, "error", err)
			s.discord.SendError("scheduled task " + lockName + " failed: " + err.Error())
		}
	})
	if err != nil {
		slog.Error("failed to register cron job", "spec", spec, "name", lockName, "error", err)
	}
}
