package types

type Hibernate struct {
	Start int64 `json:"start"` // This may be redundant.
	End   int64 `json:"end"`
}
