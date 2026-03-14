package service

type ColumnDef struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	NotNull bool   `json:"not_null"`
	Default string `json:"default"`
	After   string `json:"after"`
	Comment string `json:"comment"`
}

type TableDef struct {
	Name       string      `json:"name"`
	PrimaryKey string      `json:"primary_key"`
	AutoInc    bool        `json:"auto_increment"`
	Columns    []ColumnDef `json:"columns"`
	UniqueKeys []string    `json:"unique_keys"`
	Engine     string      `json:"engine"`
	Charset    string      `json:"charset"`
}

type CompatCheckResult struct {
	CheckTime      string              `json:"check_time"`
	TotalTables    int                 `json:"total_tables"`
	MissingTables  []string            `json:"missing_tables"`
	ExistingTables []string            `json:"existing_tables"`
	ExtraTables    []string            `json:"extra_tables"`
	MissingColumns []MissingColumnInfo `json:"missing_columns"`
	Summary        string              `json:"summary"`
}

type MissingColumnInfo struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Type   string `json:"type"`
}

type CompatFixResult struct {
	FixTime       string   `json:"fix_time"`
	TablesCreated []string `json:"tables_created"`
	ColumnsAdded  []string `json:"columns_added"`
	Errors        []string `json:"errors"`
	AdminCreated  bool     `json:"admin_created"`
	Summary       string   `json:"summary"`
}

type DBCompatService struct{}

var dbCompatService = &DBCompatService{}

func CheckDBCompat() (*CompatCheckResult, error) {
	return dbCompatService.Check()
}

func FixDBCompat() (*CompatFixResult, error) {
	return dbCompatService.Fix()
}
