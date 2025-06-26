package pb

type PocketbaseAdmin struct {
	Token  string `json:"token"`
	Record struct {
		ID string `json:"id"`
	} `json:"record"`
	BaseDomain string
}

type PbResponse[T any] struct {
	Items      []T `json:"items"`
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
}

type BackupCollection struct {
	Key      string `json:"key"`
	Modified string `json:"modified"`
}

type CollegeCollection struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Alias string `json:"alias"`
}

type BranchCollection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Ciwg bool   `json:"ciwg"`
}

type RankCollection struct {
	ID         string `json:"id"`
	Year       int    `json:"year"`
	Round      int    `json:"round"`
	JEE_OPEN   int    `json:"jee_open"`
	JEE_CLOSE  int    `json:"jee_close"`
	DASA_OPEN  int    `json:"dasa_open"`
	DASA_CLOSE int    `json:"dasa_close"`
	College    string `json:"college"`
	Branch     string `json:"branch"`
	Expand     struct {
		College CollegeCollection `json:"college"`
		Branch  BranchCollection  `json:"branch"`
	} `json:"expand"`
}

type BranchCreateRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Ciwg bool   `json:"ciwg"`
}

type RankCreateRequest struct {
	Year       int    `json:"year"`
	Round      int    `json:"round"`
	JEE_OPEN   int    `json:"jee_open"`
	JEE_CLOSE  int    `json:"jee_close"`
	DASA_OPEN  int    `json:"dasa_open"`
	DASA_CLOSE int    `json:"dasa_close"`
	College    string `json:"college"`
	Branch     string `json:"branch"`
}
