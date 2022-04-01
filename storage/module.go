package storage

type ModuleOpt func(db *DatabaseModule)

func SetStorage(s StorageStreamer) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Storage = s
	}
}

func SetLocationGetter(l LocationGetter) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Locator = l
	}
}

func SetEnvironmentGetter(e EnvironmentGetter) ModuleOpt {
	return func(db *DatabaseModule) {
		db.Environment = e
	}
}

func NewStorageModule(s StorageStreamer, moduleOpts ...ModuleOpt) *DatabaseModule {
	d := &DatabaseModule{
		Locator:     &DefaultLocation{},
		Environment: &DefaultEnvironment{},
		Storage:     s,
	}

	for _, opts := range moduleOpts {
		opts(d)
	}
	return d
}
