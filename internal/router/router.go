package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"lostark-todo-backend/internal/auth"
	"lostark-todo-backend/internal/handler"
	"lostark-todo-backend/internal/middleware"
)

// NewRouter creates and configures the Chi router with all route definitions.
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
) chi.Router {
	r := chi.NewRouter()

	// Global middleware (order matters)
	r.Use(corsMw)
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(rateLimiter.Middleware)
	r.Use(authMw)

	// ========== Health check ==========
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

	// ========== General Todo routes ==========
	r.Route("/api/v1/general-todo", func(r chi.Router) {
		r.Get("/", generalTodoHandler.GetOverview)
		r.Post("/folder", generalTodoHandler.CreateFolder)
		r.Patch("/folder/{folderId}", generalTodoHandler.UpdateFolder)
		r.Delete("/folder/{folderId}", generalTodoHandler.DeleteFolder)
		r.Patch("/folder/sort", generalTodoHandler.UpdateFolderSort)
		r.Post("/category", generalTodoHandler.CreateCategory)
		r.Patch("/category/{categoryId}", generalTodoHandler.UpdateCategory)
		r.Delete("/category/{categoryId}", generalTodoHandler.DeleteCategory)
		r.Patch("/category/sort", generalTodoHandler.UpdateCategorySort)
		r.Post("/item", generalTodoHandler.CreateItem)
		r.Patch("/item/{itemId}", generalTodoHandler.UpdateItem)
		r.Delete("/item/{itemId}", generalTodoHandler.DeleteItem)
		r.Patch("/item/sort", generalTodoHandler.UpdateItemSort)
		r.Post("/status", generalTodoHandler.CreateStatus)
		r.Patch("/status/{statusId}", generalTodoHandler.ToggleStatus)
		r.Delete("/status/{statusId}", generalTodoHandler.DeleteStatus)
	})

	// ========== Schedule routes ==========
	r.Route("/api/v1/schedule", func(r chi.Router) {
		r.Get("/", scheduleHandler.GetSchedules)
		r.Post("/", scheduleHandler.CreateSchedule)
		r.Patch("/{scheduleId}", scheduleHandler.UpdateSchedule)
		r.Delete("/{scheduleId}", scheduleHandler.DeleteSchedule)
		r.Get("/raid-categories", scheduleHandler.GetRaidCategories)
		r.Get("/characters", scheduleHandler.GetCharacters)
		r.Get("/week", scheduleHandler.GetWeekSchedule)
	})

	// ========== Notification routes ==========
	r.Route("/api/v1/notification", func(r chi.Router) {
		r.Get("/", notificationHandler.GetNotifications)
		r.Patch("/{notificationId}", notificationHandler.MarkAsRead)
		r.Delete("/{notificationId}", notificationHandler.DeleteNotification)
		r.Delete("/", notificationHandler.DeleteAllRead)
	})

	// ========== Cube routes ==========
	r.Route("/api/v1/cube", func(r chi.Router) {
		r.Get("/", cubeHandler.GetCubeData)
		r.Post("/", cubeHandler.CreateOrUpdateCube)
		r.Patch("/{cubeId}", cubeHandler.UpdateCube)
	})

	// ========== Logs routes ==========
	r.Route("/api/v1/logs", func(r chi.Router) {
		r.Get("/", logsHandler.GetLogs)
		r.Post("/", logsHandler.CreateLog)
	})

	// ========== Server Todo routes ==========
	r.Route("/api/v1/server-todo", func(r chi.Router) {
		r.Get("/", serverTodoHandler.GetServerTodos)
		r.Post("/", serverTodoHandler.CreateServerTodo)
		r.Patch("/{serverTodoId}", serverTodoHandler.UpdateServerTodo)
		r.Delete("/{serverTodoId}", serverTodoHandler.DeleteServerTodo)
		r.Post("/state/{stateId}", serverTodoHandler.ToggleCheck)
	})

	// ========== Life Energy routes ==========
	r.Route("/api/v1/life-energy", func(r chi.Router) {
		r.Get("/", lifeEnergyHandler.GetLifeEnergies)
		r.Post("/", lifeEnergyHandler.CreateLifeEnergy)
		r.Patch("/{lifeEnergyId}", lifeEnergyHandler.UpdateLifeEnergy)
		r.Delete("/{lifeEnergyId}", lifeEnergyHandler.DeleteLifeEnergy)
	})

	// ========== Custom Todo routes ==========
	r.Route("/api/v1/custom-todo", func(r chi.Router) {
		r.Get("/", customTodoHandler.GetCustomTodos)
		r.Post("/", customTodoHandler.CreateCustomTodo)
		r.Patch("/{customTodoId}", customTodoHandler.UpdateCustomTodo)
		r.Delete("/{customTodoId}", customTodoHandler.DeleteCustomTodo)
		r.Post("/{customTodoId}/check", customTodoHandler.ToggleCheck)
	})

	// ========== Community routes ==========
	r.Route("/api/v1/community", func(r chi.Router) {
		r.Get("/", communityHandler.ListPosts)
		r.Get("/{communityId}", communityHandler.GetPost)
		r.Post("/", communityHandler.CreatePost)
		r.Patch("/{communityId}", communityHandler.UpdatePost)
		r.Delete("/{communityId}", communityHandler.DeletePost)
		r.Post("/{communityId}/like", communityHandler.ToggleLike)
	})

	// ========== Comment routes ==========
	r.Route("/api/v1/comments", func(r chi.Router) {
		r.Get("/", commentsHandler.GetComments)
		r.Post("/", commentsHandler.CreateComment)
		r.Patch("/{commentId}", commentsHandler.UpdateComment)
		r.Delete("/{commentId}", commentsHandler.DeleteComment)
	})

	// ========== Content routes ==========
	r.Get("/api/v1/content/week", contentHandler.GetWeekContent)

	// ========== MyGame routes ==========
	r.Get("/api/v1/games", myGameHandler.ListGames)
	r.Get("/api/v1/events", myGameHandler.ListEvents)
	r.Post("/api/v1/suggestions", myGameHandler.CreateSuggestion)

	// ========== Admin routes ==========
	r.Route("/admin", func(r chi.Router) {
		r.Get("/members", adminHandler.ListMembers)
		r.Get("/members/{memberId}", adminHandler.GetMemberDetail)
		r.Patch("/members/{memberId}", adminHandler.UpdateMember)
		r.Delete("/members/{memberId}", adminHandler.DeleteMember)
		r.Get("/members/search", adminHandler.ListMembers) // search uses same handler with query params
		r.Get("/dashboard", adminHandler.GetDashboard)
	})

	return r
}
