package mod

// ModECSAPIFactoryImpl はModECSAPIFactoryの基本実装
type ModECSAPIFactoryImpl struct {
	apis map[string]ModECSAPI
}

// NewModECSAPIFactory 新しいファクトリーを作成
func NewModECSAPIFactory() ModECSAPIFactory {
	return &ModECSAPIFactoryImpl{
		apis: make(map[string]ModECSAPI),
	}
}

func (f *ModECSAPIFactoryImpl) Create(modID string, config ModConfig) (ModECSAPI, error) {
	if _, exists := f.apis[modID]; exists {
		return nil, ErrModAlreadyExists
	}

	api := NewModECSAPI(modID, config)
	f.apis[modID] = api
	return api, nil
}

func (f *ModECSAPIFactoryImpl) Destroy(modID string) error {
	if _, exists := f.apis[modID]; !exists {
		return ErrModNotFound
	}

	delete(f.apis, modID)
	return nil
}
