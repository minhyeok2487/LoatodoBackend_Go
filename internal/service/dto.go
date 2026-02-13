package service

// CharacterResponse is the full character DTO returned to clients.
type CharacterResponse struct {
	ID                 int64                `json:"id"`
	ServerName         string               `json:"serverName"`
	CharacterName      string               `json:"characterName"`
	CharacterLevel     int                  `json:"characterLevel"`
	CharacterClassName string               `json:"characterClassName"`
	CharacterImage     *string              `json:"characterImage"`
	ItemLevel          float64              `json:"itemLevel"`
	CombatPower        float64              `json:"combatPower"`
	SortNumber         int                  `json:"sortNumber"`
	Memo               *string              `json:"memo"`
	GoldCharacter      bool                 `json:"goldCharacter"`
	ChallengeGuardian  bool                 `json:"challengeGuardian"`
	ChallengeAbyss     bool                 `json:"challengeAbyss"`
	DayTodo            DayTodoResponse      `json:"dayTodo"`
	WeekTodo           WeekTodoResponse     `json:"weekTodo"`
	Settings           SettingsResponse     `json:"settings"`
	TodoV2List         []TodoV2Response     `json:"todoV2List"`
	RaidBusGoldList    []RaidBusGoldResponse `json:"raidBusGoldList"`
}

// DayTodoResponse is the daily todo DTO.
type DayTodoResponse struct {
	ChaosCheck        int     `json:"chaosCheck"`
	ChaosGauge        int     `json:"chaosGauge"`
	ChaosGold         float64 `json:"chaosGold"`
	GuardianCheck     int     `json:"guardianCheck"`
	GuardianGauge     int     `json:"guardianGauge"`
	GuardianGold      float64 `json:"guardianGold"`
	EponaCheck        int     `json:"eponaCheck"`
	EponaGauge        int     `json:"eponaGauge"`
	WeekTotalGold     float64 `json:"weekTotalGold"`
}

// WeekTodoResponse is the weekly todo DTO.
type WeekTodoResponse struct {
	WeekEpona          int  `json:"weekEpona"`
	WeekEponaCheck     bool `json:"weekEponaCheck"`
	SilmaelExchange    bool `json:"silmaelExchange"`
	CubeTicket         int  `json:"cubeTicket"`
	GoldCheckVersion   int  `json:"goldCheckVersion"`
	ElysianCount       int  `json:"elysianCount"`
	ElysianAllCheck    bool `json:"elysianAllCheck"`
}

// SettingsResponse is the character settings DTO.
type SettingsResponse struct {
	ShowCharacter    bool `json:"showCharacter"`
	ShowChaos        bool `json:"showChaos"`
	ShowGuardian     bool `json:"showGuardian"`
	ShowEpona        bool `json:"showEpona"`
	ShowWeekEpona    bool `json:"showWeekEpona"`
	ShowSilmael      bool `json:"showSilmael"`
	ShowCube         bool `json:"showCube"`
	GoldCheckVersion bool `json:"goldCheckVersion"`
}

// TodoV2Response is the raid/weekly todo DTO.
type TodoV2Response struct {
	ID                int64   `json:"id"`
	WeekCategory      string  `json:"weekCategory"`
	WeekContentName   string  `json:"weekContentName"`
	WeekContentGate   int     `json:"weekContentGate"`
	Gold              float64 `json:"gold"`
	GoldCheck         bool    `json:"goldCheck"`
	Check             bool    `json:"check"`
	SortNumber        int     `json:"sortNumber"`
	Message           *string `json:"message"`
	CurrentGate       int     `json:"currentGate"`
	TotalGate         int     `json:"totalGate"`
	MoreReward        bool    `json:"moreReward"`
	MoreRewardGold    float64 `json:"moreRewardGold"`
}

// RaidBusGoldResponse is the bus gold DTO.
type RaidBusGoldResponse struct {
	WeekCategory string  `json:"weekCategory"`
	BusGold      float64 `json:"busGold"`
}

// WeekRaidFormResponse is the available raid form DTO.
type WeekRaidFormResponse struct {
	ID                int64   `json:"id"`
	WeekCategory      string  `json:"weekCategory"`
	WeekContentName   string  `json:"weekContentName"`
	WeekContentGate   int     `json:"weekContentGate"`
	Gold              float64 `json:"gold"`
	ItemLevel         float64 `json:"itemLevel"`
	Selected          bool    `json:"selected"`
}
