package db

import (
	"database/sql"
	"time"
)

// Member represents the member table.
type Member struct {
	ID                     int64          `json:"id"`
	APIKey                 sql.NullString `json:"apiKey"`
	Username               sql.NullString `json:"username"`
	AuthProvider           sql.NullString `json:"authProvider"`
	AccessKey              sql.NullString `json:"accessKey"`
	Password               sql.NullString `json:"password"`
	MainCharacter          sql.NullString `json:"mainCharacter"`
	Role                   string         `json:"role"` // USER, ADMIN
	AdsDate                sql.NullTime   `json:"adsDate"`
	InspectionScheduleHour int            `json:"inspectionScheduleHour"`
	CreatedDate            time.Time      `json:"createdDate"`
	LastModifiedDate       time.Time      `json:"lastModifiedDate"`
}

// Character represents the characters table with embedded DayTodo, WeekTodo, and Settings.
type Character struct {
	ID                 int64          `json:"id"`
	ServerName         string         `json:"serverName"`
	CharacterName      string         `json:"characterName"`
	CharacterLevel     int            `json:"characterLevel"`
	CharacterClassName string         `json:"characterClassName"`
	CharacterImage     sql.NullString `json:"characterImage"`
	ItemLevel          float64        `json:"itemLevel"`
	CombatPower        float64        `json:"combatPower"`
	SortNumber         int            `json:"sortNumber"`
	Memo               sql.NullString `json:"memo"`
	MemberID           int64          `json:"memberId"`
	GoldCharacter      bool           `json:"goldCharacter"`
	ChallengeGuardian  bool           `json:"challengeGuardian"`
	ChallengeAbyss     bool           `json:"challengeAbyss"`
	IsDeleted          bool           `json:"isDeleted"`
	CreatedDate        time.Time      `json:"createdDate"`
	LastModifiedDate   time.Time      `json:"lastModifiedDate"`

	// DayTodo embedded columns
	ChaosName          sql.NullString  `json:"chaosName"`
	ChaosID            sql.NullInt64   `json:"chaosId"`
	GuardianID         sql.NullInt64   `json:"guardianId"`
	ChaosCheck         int             `json:"chaosCheck"`
	ChaosGauge         int             `json:"chaosGauge"`
	ChaosGold          float64         `json:"chaosGold"`
	GuardianName       sql.NullString  `json:"guardianName"`
	GuardianCheck      int             `json:"guardianCheck"`
	GuardianGauge      int             `json:"guardianGauge"`
	GuardianGold       float64         `json:"guardianGold"`
	EponaCheck2        int             `json:"eponaCheck2"`
	EponaGauge         int             `json:"eponaGauge"`
	WeekTotalGold      float64         `json:"weekTotalGold"`
	BeforeEponaGauge   int             `json:"beforeEponaGauge"`
	BeforeChaosGauge   int             `json:"beforeChaosGauge"`
	BeforeGuardianGauge int            `json:"beforeGuardianGauge"`

	// WeekTodo embedded columns
	WeekEpona      int  `json:"weekEpona"`
	SilmaelChange  bool `json:"silmaelChange"`
	CubeTicket     int  `json:"cubeTicket"`
	ElysianCount   int  `json:"elysianCount"`

	// Settings embedded columns
	ShowCharacter        bool   `json:"showCharacter"`
	ShowEpona            bool   `json:"showEpona"`
	ThresholdEpona       int    `json:"thresholdEpona"`
	ShowChaos            bool   `json:"showChaos"`
	ThresholdChaos       int    `json:"thresholdChaos"`
	ShowGuardian         bool   `json:"showGuardian"`
	ThresholdGuardian    int    `json:"thresholdGuardian"`
	ShowWeekTodo         bool   `json:"showWeekTodo"`
	ShowWeekEpona        bool   `json:"showWeekEpona"`
	ShowSilmaelChange    bool   `json:"showSilmaelChange"`
	ShowCubeTicket       bool   `json:"showCubeTicket"`
	GoldCheckVersion     bool   `json:"goldCheckVersion"`
	GoldCheckPolicyEnum  string `json:"goldCheckPolicyEnum"`
	LinkCubeCal          bool   `json:"linkCubeCal"`
	ShowMoreButton       bool   `json:"showMoreButton"`
	ShowElysian          bool   `json:"showElysian"`
}

// Content represents the content table (single table inheritance with dtype).
type Content struct {
	ID                  int64          `json:"id"`
	DType               string         `json:"dtype"` // DayContent or WeekContent
	Category            sql.NullString `json:"category"`
	Name                sql.NullString `json:"name"`
	Level               float64        `json:"level"`

	// DayContent specific fields
	Shilling         float64 `json:"shilling"`
	HonorShard       float64 `json:"honorShard"`
	LeapStone        float64 `json:"leapStone"`
	DestructionStone float64 `json:"destructionStone"`
	GuardianStone    float64 `json:"guardianStone"`
	Jewelry          float64 `json:"jewelry"`

	// WeekContent specific fields
	WeekCategory        sql.NullString `json:"weekCategory"`
	WeekContentCategory sql.NullString `json:"weekContentCategory"`
	Gate                int            `json:"gate"`
	Gold                int            `json:"gold"`
	CharacterGold       int            `json:"characterGold"`
	CoolTime            int            `json:"coolTime"`
	MoreRewardGold      int            `json:"moreRewardGold"`
}

// TodoV2 represents the todo_v2 table.
type TodoV2 struct {
	ID               int64          `json:"id"`
	ContentID        sql.NullInt64  `json:"contentId"`
	IsChecked        bool           `json:"isChecked"`
	Gold             int            `json:"gold"`
	CharacterGold    int            `json:"characterGold"`
	Message          sql.NullString `json:"message"`
	CharacterID      int64          `json:"characterId"`
	CoolTime         int            `json:"coolTime"`
	SortNumber       int            `json:"sortNumber"`
	GoldCheck        bool           `json:"goldCheck"`
	MoreRewardCheck  bool           `json:"moreRewardCheck"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// Friends represents the friends table.
type Friends struct {
	ID               int64     `json:"id"`
	MemberID         int64     `json:"memberId"`
	FromMember       int64     `json:"fromMember"`
	AreWeFriend      bool      `json:"areWeFriend"`
	Ordering         int       `json:"ordering"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`

	// FriendSettings embedded columns
	ShowDayTodo  bool `json:"showDayTodo"`
	ShowRaid     bool `json:"showRaid"`
	ShowWeekTodo bool `json:"showWeekTodo"`
	CheckDayTodo bool `json:"checkDayTodo"`
	CheckRaid    bool `json:"checkRaid"`
	CheckWeekTodo bool `json:"checkWeekTodo"`
	UpdateGauge  bool `json:"updateGauge"`
	UpdateRaid   bool `json:"updateRaid"`
	Setting      bool `json:"setting"`
}

// Schedule represents the schedule table.
type Schedule struct {
	ID                    int64          `json:"id"`
	CharacterID           int64          `json:"characterId"`
	ScheduleRaidCategory  string         `json:"scheduleRaidCategory"`
	ScheduleCategory      string         `json:"scheduleCategory"`
	RaidName              sql.NullString `json:"raidName"`
	RaidLevel             sql.NullString `json:"raidLevel"`
	DayOfWeek             sql.NullInt64  `json:"dayOfWeek"`
	Time                  sql.NullString `json:"time"`
	Memo                  sql.NullString `json:"memo"`
	RepeatWeek            bool           `json:"repeatWeek"`
	Leader                bool           `json:"leader"`
	LeaderScheduleID      int64          `json:"leaderScheduleId"`
	Checked               bool           `json:"checked"`
	Date                  sql.NullString `json:"date"`
	AutoCheck             bool           `json:"autoCheck"`
	CreatedDate           time.Time      `json:"createdDate"`
	LastModifiedDate      time.Time      `json:"lastModifiedDate"`
}

// Notification represents the notification table.
type Notification struct {
	ID                     int64          `json:"id"`
	Content                sql.NullString `json:"content"`
	IsRead                 bool           `json:"isRead"`
	NotificationType       string         `json:"notificationType"`
	BoardID                int64          `json:"boardId"`
	CommentID              int64          `json:"commentId"`
	FriendID               int64          `json:"friendId"`
	FriendUsername         sql.NullString `json:"friendUsername"`
	FriendCharacterName    sql.NullString `json:"friendCharacterName"`
	CommunityID            int64          `json:"communityId"`
	InspectionCharacterID  int64          `json:"inspectionCharacterId"`
	ReceiverID             int64          `json:"receiverId"`
	CreatedDate            time.Time      `json:"createdDate"`
	LastModifiedDate       time.Time      `json:"lastModifiedDate"`
}

// Market represents the market table.
type Market struct {
	ID               int64          `json:"id"`
	LostarkMarketID  int64          `json:"lostarkMarketId"`
	Name             sql.NullString `json:"name"`
	Grade            sql.NullString `json:"grade"`
	Icon             sql.NullString `json:"icon"`
	BundleCount      int            `json:"bundleCount"`
	YDayAvgPrice     float64        `json:"yDayAvgPrice"`
	RecentPrice      int            `json:"recentPrice"`
	CurrentMinPrice  int            `json:"currentMinPrice"`
	CategoryCode     int            `json:"categoryCode"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// Cubes represents the cubes table.
type Cubes struct {
	ID          int64 `json:"id"`
	CharacterID int64 `json:"characterId"`
	Ban1        int   `json:"ban1"`
	Ban2        int   `json:"ban2"`
	Ban3        int   `json:"ban3"`
	Ban4        int   `json:"ban4"`
	Ban5        int   `json:"ban5"`
	Unlock1     int   `json:"unlock1"`
	Unlock2     int   `json:"unlock2"`
	Unlock3     int   `json:"unlock3"`
	Unlock4     int   `json:"unlock4"`
}

// Logs represents the logs table.
type Logs struct {
	ID               int64          `json:"id"`
	LocalDate        sql.NullString `json:"localDate"`
	MemberID         int64          `json:"memberId"`
	CharacterID      int64          `json:"characterId"`
	LogType          sql.NullString `json:"logType"`
	LogContent       sql.NullString `json:"logContent"`
	Name             sql.NullString `json:"name"`
	Message          sql.NullString `json:"message"`
	Profit           float64        `json:"profit"`
	Deleted          bool           `json:"deleted"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// LifeEnergy represents the life_energy table.
type LifeEnergy struct {
	ID               int64     `json:"id"`
	MemberID         int64     `json:"memberId"`
	CharacterName    string    `json:"characterName"`
	Energy           int       `json:"energy"`
	MaxEnergy        int       `json:"maxEnergy"`
	Beatrice         bool      `json:"beatrice"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// GeneralTodoFolder represents the general_todo_folder table.
type GeneralTodoFolder struct {
	ID               int64     `json:"id"`
	MemberID         int64     `json:"memberId"`
	Name             string    `json:"name"`
	SortOrder        int       `json:"sortOrder"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// GeneralTodoCategory represents the general_todo_category table.
type GeneralTodoCategory struct {
	ID               int64          `json:"id"`
	FolderID         int64          `json:"folderId"`
	MemberID         int64          `json:"memberId"`
	Name             string         `json:"name"`
	Color            sql.NullString `json:"color"`
	SortOrder        int            `json:"sortOrder"`
	ViewMode         string         `json:"viewMode"` // LIST, BOARD, etc.
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// GeneralTodoItem represents the general_todo_item table.
type GeneralTodoItem struct {
	ID               int64          `json:"id"`
	FolderID         int64          `json:"folderId"`
	CategoryID       int64          `json:"categoryId"`
	MemberID         int64          `json:"memberId"`
	StatusID         int64          `json:"statusId"`
	Title            string         `json:"title"`
	Description      sql.NullString `json:"description"`
	StartDate        sql.NullTime   `json:"startDate"`
	DueDate          sql.NullTime   `json:"dueDate"`
	AllDay           bool           `json:"allDay"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// GeneralTodoStatus represents the general_todo_status table.
type GeneralTodoStatus struct {
	ID         int64  `json:"id"`
	CategoryID int64  `json:"categoryId"`
	MemberID   int64  `json:"memberId"`
	Name       string `json:"name"`
	SortOrder  int    `json:"sortOrder"`
}

// Community represents the community table.
type Community struct {
	ID                 int64          `json:"id"`
	MemberID           int64          `json:"memberId"`
	CharacterImage     sql.NullString `json:"characterImage"`
	CharacterClassName sql.NullString `json:"characterClassName"`
	Name               sql.NullString `json:"name"`
	ShowName           bool           `json:"showName"`
	Body               sql.NullString `json:"body"`
	Category           sql.NullString `json:"category"`
	RootParentID       int64          `json:"rootParentId"`
	CommentParentID    int64          `json:"commentParentId"`
	Deleted            bool           `json:"deleted"`
	CreatedDate        time.Time      `json:"createdDate"`
	LastModifiedDate   time.Time      `json:"lastModifiedDate"`
}

// CommunityImages represents the community_images table.
type CommunityImages struct {
	ID               int64          `json:"id"`
	CommunityID      int64          `json:"communityId"`
	URL              sql.NullString `json:"url"`
	FileName         sql.NullString `json:"fileName"`
	Ordering         int64          `json:"ordering"`
	Deleted          bool           `json:"deleted"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// CommunityLike represents the community_like table.
type CommunityLike struct {
	ID               int64     `json:"id"`
	MemberID         int64     `json:"memberId"`
	CommunityID      int64     `json:"communityId"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// Comments represents the comments table.
type Comments struct {
	ID               int64          `json:"id"`
	Body             sql.NullString `json:"body"`
	MemberID         int64          `json:"memberId"`
	ParentID         int64          `json:"parentId"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// Follow represents the follow table.
type Follow struct {
	ID          int64 `json:"id"`
	FollowerID  int64 `json:"followerId"`
	FollowingID int64 `json:"followingId"`
}

// ServerTodo represents the server_todo table.
type ServerTodo struct {
	ID               int64          `json:"id"`
	ContentName      string         `json:"contentName"`
	DefaultEnabled   bool           `json:"defaultEnabled"`
	MemberID         sql.NullInt64  `json:"memberId"`
	Frequency        sql.NullString `json:"frequency"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// ServerTodoVisibleWeekday represents the server_todo_visible_weekday join table.
type ServerTodoVisibleWeekday struct {
	ServerTodoID   int64  `json:"serverTodoId"`
	VisibleWeekday string `json:"visibleWeekday"`
}

// ServerTodoState represents the server_todo_state table.
type ServerTodoState struct {
	ID               int64     `json:"id"`
	ServerTodoID     int64     `json:"serverTodoId"`
	MemberID         int64     `json:"memberId"`
	ServerName       string    `json:"serverName"`
	Enabled          bool      `json:"enabled"`
	Checked          bool      `json:"checked"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// CustomTodo represents the custom_todo table.
type CustomTodo struct {
	ID               int64     `json:"id"`
	ContentName      string    `json:"contentName"`
	IsChecked        bool      `json:"isChecked"`
	Frequency        string    `json:"frequency"` // DAILY, WEEKLY
	CharacterID      int64     `json:"characterId"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// RaidBusGold represents the raid_bus_gold table.
type RaidBusGold struct {
	ID               int64     `json:"id"`
	CharacterID      int64     `json:"characterId"`
	WeekCategory     string    `json:"weekCategory"`
	BusGold          int       `json:"busGold"`
	Fixed            bool      `json:"fixed"`
	CreatedDate      time.Time `json:"createdDate"`
	LastModifiedDate time.Time `json:"lastModifiedDate"`
}

// Ads represents the ads table.
type Ads struct {
	ID               int64          `json:"id"`
	Name             sql.NullString `json:"name"`
	ProposerEmail    sql.NullString `json:"proposerEmail"`
	MemberID         int64          `json:"memberId"`
	Checked          bool           `json:"checked"`
	CreatedDate      time.Time      `json:"createdDate"`
	LastModifiedDate time.Time      `json:"lastModifiedDate"`
}

// KeyValue represents the key_value table.
type KeyValue struct {
	ID        int64  `json:"id"`
	KeyName   string `json:"keyName"`
	ValueName string `json:"valueName"`
}

// ShedLock represents the shedlock table.
type ShedLock struct {
	Name      string    `json:"name"`
	LockUntil time.Time `json:"lockUntil"`
	LockedAt  time.Time `json:"lockedAt"`
	LockedBy  string    `json:"lockedBy"`
}
