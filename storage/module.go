package storage

type ModuleOpt func(db *DatabaseModule)

func SetStorage(s StorageEngine) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Storage = s
	}
}

func SetLocationGetter(l LocationGetter) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Location = l
	}
}

func SetEnvironmentGetter(e EnvironmentGetter) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Environment = e
	}
}

func NewStorageModule(s StorageEngine, moduleOpts ...ModuleOpt) *DatabaseModule {
	d := &DatabaseModule{
		Location:    &DefaultLocation{},
		Environment: &DefaultEnvironment{},
		Storage:     s,
	}

	for _, opts := range moduleOpts {
		opts(d)
	}
	return d
}
