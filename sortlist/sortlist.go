package sortlist

type Sortable interface {
	GetUniqueID() uint64
	GetScore() uint64
}

type SortList interface {
	GetUser(uniqueID uint64) Sortable
	RemoveUser(uniqueID uint64)
	UpdateUser(uniqueID uint64, score uint64)

	GetRank(uniqueID uint64) uint64
	GetRankUser(rank uint64) Sortable
	GetRankUsers(fromRnk int64, toRank int64) []Sortable
	GetUserCount() uint64

	Load()
	Save()
}
