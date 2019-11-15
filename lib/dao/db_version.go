package dao

func NewDbVersionPostgresDao(postgresStorage PostgresStorage) DbVersionDao {
	return &dbVersionStorage{
		PostgresStorage: postgresStorage,
		dbVersionEquey:  "SELECT version FROM schema_migrations",
	}
}

type dbVersionStorage struct {
	PostgresStorage
	dbVersionEquey string
}

func (this dbVersionStorage) GetDbVersion() (int, error) {
	ver, found, err := this.PostgresStorage.doFindAndReturn(this.dbVersionEquey, IntColumnMapper)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, nil
	}
	return ver.(int), nil
}
