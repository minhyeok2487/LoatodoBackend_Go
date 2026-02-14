package service

// CharacterResponse is the full character DTO returned to clients.
// IMPORTANT: This must match the Spring Boot CharacterResponse exactly (flat structure).
type CharacterResponse struct {
	CharacterID   int64   `json:"characterId"`
	MemberID      int64   `json:"memberId"`
	CharacterName string  `json:"characterName"`
	CharacterClassName string `json:"characterClassName"`
	CharacterImage *string `json:"characterImage"`
	ItemLevel     float64 `json:"itemLevel"`
	CombatPower   float64 `json:"combatPower"`
	ServerName    string  `json:"serverName"`
	SortNumber    int     `json:"sortNumber"`

	// DayTodo fields (flat, not nested)
	Chaos        *DayContentResponse `json:"chaos"`
	ChaosCheck   int                 `json:"chaosCheck"`
	ChaosGauge   int                 `json:"chaosGauge"`
	ChaosGold    float64             `json:"chaosGold"`
	Guardian     *DayContentResponse `json:"guardian"`
	GuardianCheck int                `json:"guardianCheck"`
	GuardianGauge int                `json:"guardianGauge"`
	GuardianGold  float64            `json:"guardianGold"`
	EponaCheck   int                 `json:"eponaCheck"`
	EponaGauge   int                 `json:"eponaGauge"`

	GoldCharacter     bool `json:"goldCharacter"`
	ChallengeGuardian bool `json:"challengeGuardian"`
	ChallengeAbyss    bool `json:"challengeAbyss"`

	Settings SettingsResponse `json:"settings"`

	// WeekTodo fields (flat, not nested)
	WeekEpona     int  `json:"weekEpona"`
	SilmaelChange bool `json:"silmaelChange"`
	CubeTicket    int  `json:"cubeTicket"`

	WeekDayTodoGold       float64 `json:"weekDayTodoGold"`
	WeekRaidGold          int     `json:"weekRaidGold"`
	WeekCharacterRaidGold int     `json:"weekCharacterRaidGold"`

	TodoList []TodoResponseDto `json:"todoList"`

	BeforeEponaGauge    int `json:"beforeEponaGauge"`
	BeforeChaosGauge    int `json:"beforeChaosGauge"`
	BeforeGuardianGauge int `json:"beforeGuardianGauge"`

	Memo *string `json:"memo"`

	ElysianCount int `json:"elysianCount"`

	// Internal fields not in JSON (used for gold calculation)
	RaidBusGoldList []RaidBusGoldResponse `json:"-"`
}

// DayContentResponse is the DayContent DTO (from content table).
type DayContentResponse struct {
	ID               int64   `json:"id"`
	Category         string  `json:"category"`
	Name             string  `json:"name"`
	Level            float64 `json:"level"`
	Shilling         float64 `json:"shilling"`
	HonorShard       float64 `json:"honorShard"`
	LeapStone        float64 `json:"leapStone"`
	DestructionStone float64 `json:"destructionStone"`
	GuardianStone    float64 `json:"guardianStone"`
	Jewelry          float64 `json:"jewelry"`
}

// SettingsResponse is the character settings DTO.
type SettingsResponse struct {
	ShowCharacter      bool   `json:"showCharacter"`
	ShowEpona          bool   `json:"showEpona"`
	ThresholdEpona     int    `json:"thresholdEpona"`
	ShowChaos          bool   `json:"showChaos"`
	ThresholdChaos     int    `json:"thresholdChaos"`
	ShowGuardian       bool   `json:"showGuardian"`
	ThresholdGuardian  int    `json:"thresholdGuardian"`
	ShowWeekTodo       bool   `json:"showWeekTodo"`
	ShowWeekEpona      bool   `json:"showWeekEpona"`
	ShowSilmaelChange  bool   `json:"showSilmaelChange"`
	ShowCubeTicket     bool   `json:"showCubeTicket"`
	GoldCheckVersion   bool   `json:"goldCheckVersion"`
	GoldCheckPolicyEnum string `json:"goldCheckPolicyEnum"`
	LinkCubeCal        bool   `json:"linkCubeCal"`
	ShowMoreButton     bool   `json:"showMoreButton"`
	ShowElysian        bool   `json:"showElysian"`
}

// TodoResponseDto is the raid/weekly todo DTO (matching Spring Boot's TodoResponseDto).
type TodoResponseDto struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	CharacterClassName   string   `json:"characterClassName"`
	Gold                 int      `json:"gold"`
	RealGold             int      `json:"realGold"`
	Check                bool     `json:"check"`
	Message              *string  `json:"message"`
	CurrentGate          int      `json:"currentGate"`
	TotalGate            int      `json:"totalGate"`
	WeekCategory         string   `json:"weekCategory"`
	WeekContentCategory  string   `json:"weekContentCategory"`
	SortNumber           int      `json:"sortNumber"`
	GoldCheck            bool     `json:"goldCheck"`
	MoreRewardCheckList  []bool   `json:"moreRewardCheckList"`
	MoreRewardGoldList   []int    `json:"moreRewardGoldList"`
	CharacterGold        int      `json:"characterGold"`
}

// RaidBusGoldResponse is the bus gold DTO.
type RaidBusGoldResponse struct {
	WeekCategory string  `json:"weekCategory"`
	BusGold      float64 `json:"busGold"`
}

// WeekRaidFormResponse is the available raid form DTO.
type WeekRaidFormResponse struct {
	ID                  int64   `json:"id"`
	WeekCategory        string  `json:"weekCategory"`
	WeekContentCategory string  `json:"weekContentCategory"`
	Name                string  `json:"name"`
	Level               float64 `json:"level"`
	Gate                int     `json:"gate"`
	Gold                float64 `json:"gold"`
	MoreRewardCheck     bool    `json:"moreRewardCheck"`
	BusGold             float64 `json:"busGold"`
	Checked             bool    `json:"checked"`
	CoolTime            int     `json:"coolTime"`
	GoldCheck           bool    `json:"goldCheck"`
	BusGoldFixed        bool    `json:"busGoldFixed"`
	CharacterGold       float64 `json:"characterGold"`
}
