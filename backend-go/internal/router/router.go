package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/handler"
	"lostark-todo-backend/internal/middleware"
)

// NewRouter creates and configures the Chi router with all route definitions.
// All paths match the Spring Boot API exactly so the frontend works without changes.
func NewRouter(
	authMw func(http.Handler) http.Handler,
	corsMw func(http.Handler) http.Handler,
	rateLimiter *middleware.RateLimiter,
	oauth2Handler *auth.OAuth2Handler,
	authHandler *handler.AuthHandler,
	memberHandler *handler.MemberHandler,
	characterHandler *handler.CharacterHandler,
	characterListHandler *handler.CharacterListHandler,
	characterDayHandler *handler.CharacterDayHandler,
	characterWeekHandler *handler.CharacterWeekHandler,
	friendHandler *handler.FriendHandler,
	generalTodoHandler *handler.GeneralTodoHandler,
	scheduleHandler *handler.ScheduleHandler,
	notificationHandler *handler.NotificationHandler,
	cubeHandler *handler.CubeHandler,
	logsHandler *handler.LogsHandler,
	serverTodoHandler *handler.ServerTodoHandler,
	lifeEnergyHandler *handler.LifeEnergyHandler,
	customTodoHandler *handler.CustomTodoHandler,
	communityHandler *handler.CommunityHandler,
	commentsHandler *handler.CommentsHandler,
	contentHandler *handler.ContentHandler,
	myGameHandler *handler.MyGameHandler,
	adminHandler *handler.AdminHandler,
	followHandler *handler.FollowHandler,
	analysisHandler *handler.AnalysisHandler,
	emailHandler *handler.EmailHandler,
) chi.Router {
	r := chi.NewRouter()

	// Global middleware (order matters)
	r.Use(corsMw)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(rateLimiter.Middleware)
	r.Use(authMw)

	// ========== Root & Health check ==========
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/manage/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"UP"}`))
	})
	r.Get("/manage/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"app":"lostark-todo-backend","lang":"go"}`))
	})

	// ========== Auth routes ==========
	r.Post("/api/v1/auth/signup", authHandler.SignUp)
	r.Post("/api/v1/auth/login", authHandler.Login)
	r.Get("/api/v1/auth/logout", authHandler.Logout)

	// ========== Email routes ==========
	r.Route("/api/v1/mail", func(r chi.Router) {
		r.Post("/", emailHandler.SendSignUpMail)
		r.Post("/password", emailHandler.SendResetPasswordMail)
		r.Post("/auth", emailHandler.VerifyEmailCode)
	})

	// ========== OAuth2 routes ==========
	r.Get("/auth/authorize", oauth2Handler.HandleAuthorize)
	r.Get("/login/oauth2/code/google", oauth2Handler.HandleCallback)

	// ========== Member routes ==========
	r.Route("/api/v1/member", func(r chi.Router) {
		r.Get("/", memberHandler.GetMember)
		r.Post("/character", memberHandler.SaveCharacter)
		r.Post("/password", memberHandler.UpdatePassword)
		r.Patch("/main-character", memberHandler.EditMainCharacter)
		r.Patch("/provider", memberHandler.ChangeProvider)
		r.Post("/ads", memberHandler.SaveAds)
		r.Delete("/characters", memberHandler.DeleteAllCharacters)
		r.Patch("/api-key", memberHandler.UpdateAPIKey)
	})

	// ========== Character routes ==========
	r.Route("/api/v1/character", func(r chi.Router) {
		// Character CRUD
		r.Put("/", characterHandler.UpdateCharacter)
		r.Post("/", characterHandler.AddCharacter)
		r.Patch("/settings", characterHandler.UpdateSettings)
		r.Patch("/gold-character", characterHandler.ToggleGoldCharacter)
		r.Post("/memo", characterHandler.UpdateMemo)
		r.Patch("/deleted", characterHandler.ToggleDeleted)
		r.Patch("/name", characterHandler.ChangeCharacterName)

		// Day todo
		r.Post("/day/check", characterDayHandler.CheckDayContent)
		r.Post("/day/gauge", characterDayHandler.UpdateGauge)
		r.Post("/day/check/all", characterDayHandler.CheckAllDayContent)
		r.Post("/day/check/all-characters", characterDayHandler.CheckAllDayCharacters)

		// Week todo
		r.Post("/week/raid", characterWeekHandler.AddRemoveWeekRaid)
		r.Get("/week/raid/form", characterWeekHandler.GetWeekRaidForm)
		r.Post("/week/raid/bus", characterWeekHandler.UpdateBusGold)
		r.Post("/week/raid/check", characterWeekHandler.CheckRaid)
		r.Post("/week/raid/message", characterWeekHandler.UpdateRaidMessage)
		r.Post("/week/raid/sort", characterWeekHandler.SortRaids)
		r.Post("/week/epona", characterWeekHandler.CheckWeekEpona)
		r.Post("/week/silmael", characterWeekHandler.ToggleSilmael)
		r.Post("/week/cube", characterWeekHandler.UpdateCubeTicket)
		r.Patch("/week/raid/gold-check", characterWeekHandler.ToggleGoldCheck)
		r.Patch("/week/gold-check-version", characterWeekHandler.ToggleGoldCheckVersion)
		r.Post("/week/raid/more-reward", characterWeekHandler.ToggleMoreReward)
		r.Post("/week/elysian", characterWeekHandler.UpdateElysian)
		r.Post("/week/elysian/all", characterWeekHandler.ToggleElysianAll)
	})

	// ========== Character list routes ==========
	r.Route("/api/v1/character-list", func(r chi.Router) {
		r.Get("/", characterListHandler.GetCharacterList)
		r.Patch("/sorting", characterListHandler.UpdateSorting)
		r.Get("/deleted", characterListHandler.GetDeletedCharacters)
	})

	// ========== Friend routes ==========
	r.Route("/api/v1/friend", func(r chi.Router) {
		r.Get("/", friendHandler.GetFriends)
		r.Get("/character/{characterName}", friendHandler.SearchCharacter)
		r.Post("/", friendHandler.SendFriendRequest)
		r.Post("/request", friendHandler.HandleFriendRequest)
		r.Patch("/settings", friendHandler.UpdateSettings)
		r.Delete("/{friendId}", friendHandler.DeleteFriend)
		r.Put("/sort", friendHandler.UpdateSortOrder)
	})

	// ========== General Todo routes (Spring Boot: /api/v1/general-todos) ==========
	r.Route("/api/v1/general-todos", func(r chi.Router) {
		r.Get("/", generalTodoHandler.GetOverview)

		// Folders
		r.Post("/folders", generalTodoHandler.CreateFolder)
		r.Patch("/folders/{folderId}", generalTodoHandler.UpdateFolder)
		r.Patch("/folders/reorder", generalTodoHandler.UpdateFolderSort)
		r.Delete("/folders/{folderId}", generalTodoHandler.DeleteFolder)

		// Categories
		r.Post("/categories/folders/{folderId}", generalTodoHandler.CreateCategoryForFolder)
		r.Patch("/categories/{categoryId}", generalTodoHandler.UpdateCategory)
		r.Patch("/categories/folders/{folderId}/reorder", generalTodoHandler.UpdateCategorySort)
		r.Delete("/categories/{categoryId}", generalTodoHandler.DeleteCategory)

		// Items
		r.Post("/items", generalTodoHandler.CreateItem)
		r.Patch("/items/{itemId}", generalTodoHandler.UpdateItem)
		r.Patch("/items/{itemId}/status", generalTodoHandler.ToggleStatus)
		r.Delete("/items/{itemId}", generalTodoHandler.DeleteItem)

		// Statuses (under categories)
		r.Post("/categories/{categoryId}/statuses", generalTodoHandler.CreateStatusForCategory)
		r.Patch("/categories/{categoryId}/statuses/{statusId}", generalTodoHandler.ToggleStatusInCategory)
		r.Patch("/categories/{categoryId}/statuses/reorder", generalTodoHandler.UpdateItemSort)
		r.Delete("/categories/{categoryId}/statuses/{statusId}", generalTodoHandler.DeleteStatusInCategory)
	})

	// ========== Schedule routes ==========
	r.Route("/api/v1/schedule", func(r chi.Router) {
		r.Get("/", scheduleHandler.GetSchedules)
		r.Post("/", scheduleHandler.CreateSchedule)
		r.Get("/raid/category", scheduleHandler.GetRaidCategories)
		r.Get("/{scheduleId}", scheduleHandler.GetSchedules) // individual schedule uses same list handler
		r.Patch("/{scheduleId}", scheduleHandler.UpdateSchedule)
		r.Delete("/{scheduleId}", scheduleHandler.DeleteSchedule)
		r.Post("/{scheduleId}/friend", scheduleHandler.EditFriend)
	})

	// ========== Notification routes ==========
	r.Route("/api/v1/notification", func(r chi.Router) {
		r.Get("/", notificationHandler.GetNotifications)
		r.Get("/status", notificationHandler.GetRecentNotification)
		r.Post("/{notificationId}", notificationHandler.MarkAsRead) // POST not PATCH
		r.Post("/all", notificationHandler.MarkAllAsRead)
	})

	// ========== Cube routes ==========
	r.Route("/api/v1/cube", func(r chi.Router) {
		r.Get("/", cubeHandler.GetCubeData)
		r.Get("/statistics", cubeHandler.CalculateProfits)
		r.Post("/", cubeHandler.CreateOrUpdateCube)
		r.Put("/", cubeHandler.UpdateCube)
		r.Delete("/{characterId}", cubeHandler.DeleteCube)
		r.Post("/spend", cubeHandler.SpendCubeTicket)
	})

	// ========== Logs routes ==========
	r.Route("/api/v1/logs", func(r chi.Router) {
		r.Get("/", logsHandler.GetLogs)
		r.Get("/profit", logsHandler.GetLogsProfit)
		r.Delete("/{logId}", logsHandler.DeleteLog)
		r.Post("/", logsHandler.CreateLog)
	})

	// ========== Server Todo routes (Spring Boot: /api/v1/server-todos) ==========
	r.Route("/api/v1/server-todos", func(r chi.Router) {
		r.Get("/", serverTodoHandler.GetServerTodos)
		r.Post("/", serverTodoHandler.CreateServerTodo)
		r.Delete("/{todoId}", serverTodoHandler.DeleteServerTodo)
		r.Patch("/{todoId}/toggle-enabled", serverTodoHandler.UpdateServerTodo)
		r.Post("/{todoId}/check", serverTodoHandler.ToggleCheck)
	})

	// ========== Life Energy routes ==========
	r.Route("/api/v1/life-energy", func(r chi.Router) {
		r.Post("/", lifeEnergyHandler.CreateLifeEnergy)
		r.Put("/", lifeEnergyHandler.UpdateLifeEnergy)
		r.Delete("/{characterName}", lifeEnergyHandler.DeleteLifeEnergy)
		r.Post("/spend", lifeEnergyHandler.SpendLifeEnergy)
	})

	// ========== Custom Todo routes (Spring Boot: /api/v1/custom) ==========
	r.Route("/api/v1/custom", func(r chi.Router) {
		r.Get("/", customTodoHandler.GetCustomTodos)
		r.Post("/", customTodoHandler.CreateCustomTodo)
		r.Patch("/{customTodoId}", customTodoHandler.UpdateCustomTodo)
		r.Post("/check", customTodoHandler.ToggleCheck)
		r.Delete("/{customTodoId}", customTodoHandler.DeleteCustomTodo)
	})

	// ========== Community routes ==========
	r.Route("/api/v1/community", func(r chi.Router) {
		r.Get("/category", communityHandler.GetCategories)
		r.Get("/", communityHandler.ListPosts)
		r.Get("/{communityId}", communityHandler.GetPost)
		r.Post("/", communityHandler.CreatePost)
		r.Post("/image", communityHandler.UploadCommunityImage)
		r.Patch("/", communityHandler.UpdatePost)
		r.Delete("/{communityId}", communityHandler.DeletePost)
		r.Post("/like/{communityId}", communityHandler.ToggleLike)
	})

	// ========== Comment routes ==========
	r.Route("/api/v1/comments", func(r chi.Router) {
		r.Get("/", commentsHandler.GetComments)
		r.Post("/", commentsHandler.CreateComment)
		r.Patch("/{commentId}", commentsHandler.UpdateComment)
		r.Delete("/{commentId}", commentsHandler.DeleteComment)
	})

	// ========== Content routes ==========
	r.Get("/api/v1/content/week/raid/category", contentHandler.GetWeekRaidCategory)

	// ========== Follow routes ==========
	r.Route("/api/v1/follow", func(r chi.Router) {
		r.Get("/", followHandler.GetFollowers)
		r.Post("/", followHandler.ToggleFollow)
	})

	// ========== Analysis routes ==========
	r.Route("/api/v1/analysis", func(r chi.Router) {
		r.Get("/", analysisHandler.GetAnalyses)
		r.Post("/", analysisHandler.CreateAnalysis)
	})

	// ========== MyGame routes ==========
	r.Route("/api/v1/games", func(r chi.Router) {
		r.Get("/", myGameHandler.ListGames)
		r.Get("/all", myGameHandler.GetAllGames)
		r.Get("/{id}", myGameHandler.GetGame)
		r.Post("/", myGameHandler.CreateGame)
		r.Post("/images", myGameHandler.UploadGameImage)
	})
	r.Route("/api/v1/events", func(r chi.Router) {
		r.Get("/", myGameHandler.ListEvents)
		r.Get("/{id}", myGameHandler.GetEvent)
		r.Post("/", myGameHandler.CreateEvent)
		r.Post("/images", myGameHandler.UploadEventImage)
	})
	r.Post("/api/v1/suggestions", myGameHandler.CreateSuggestion)

	// ========== Admin routes (Spring Boot: /admin/api/v1/*) ==========
	r.Route("/admin/api/v1", func(r chi.Router) {
		// Members
		r.Route("/members", func(r chi.Router) {
			r.Get("/", adminHandler.ListMembers)
			r.Get("/{memberId}", adminHandler.GetMemberDetail)
			r.Put("/{memberId}", adminHandler.UpdateMember)
			r.Delete("/{memberId}", adminHandler.DeleteMember)
		})

		// Dashboard
		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/member", adminHandler.GetDashboard)
			r.Get("/daily-members", adminHandler.GetDashboard)
			r.Get("/daily-characters", adminHandler.GetDashboard)
			r.Get("/summary", adminHandler.GetDashboard)
		})

		// Characters
		r.Route("/characters", func(r chi.Router) {
			r.Get("/", adminHandler.ListCharacters)
			r.Get("/{characterId}", adminHandler.GetCharacterDetail)
			r.Put("/{characterId}", adminHandler.UpdateCharacter)
			r.Delete("/{characterId}", adminHandler.DeleteCharacter)
		})

		// Friends
		r.Route("/friends", func(r chi.Router) {
			r.Get("/", adminHandler.ListFriends)
			r.Delete("/{friendId}", adminHandler.DeleteFriend)
		})

		// Notifications
		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", adminHandler.ListNotifications)
			r.Post("/broadcast", adminHandler.CreateNotification)
			r.Delete("/{notificationId}", adminHandler.DeleteNotificationAdmin)
		})

		// Contents
		r.Route("/contents", func(r chi.Router) {
			r.Get("/", adminHandler.ListContent)
			r.Get("/{contentId}", adminHandler.GetContentDetail)
			r.Post("/", adminHandler.CreateContent)
			r.Put("/{contentId}", adminHandler.UpdateContent)
			r.Delete("/{contentId}", adminHandler.DeleteContent)
		})

		// Ads
		r.Route("/ads", func(r chi.Router) {
			r.Get("/", adminHandler.ManageAds)
			r.Post("/date", adminHandler.ManageAds)
		})

		// Comments
		r.Route("/comments", func(r chi.Router) {
			r.Get("/", adminHandler.ListComments)
			r.Delete("/{commentId}", adminHandler.DeleteComment)
		})
	})

	return r
}
