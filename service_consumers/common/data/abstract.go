package data

type IDbCollection[TEntity any] interface {
	GenerateNewId() string
	Upsert(id string, document TEntity) (TEntity, error)
	FirstOrDefault(filter any, params ...any) (TEntity, error)
	Where(filter any, params ...any) (TEntity, error)
	Paginate(filter any, page int, take int, params ...any) (PagingModel[TEntity], error)
}

type IDbProvider interface {
	Connect(connectionString string)
	Disconnect()
	GetClient() any
}

type IMigratable interface {
	Migrate(connectionString string)
}

type PagingModel[T any] struct {
	TotalPage   int `json:"totalPage"`
	CurrentPage int `json:"currentPage"`
	Take        int `json:"take"`
	TotalCount  int `json:"totalCount"`
	Data        []T `json:"data"`
}
