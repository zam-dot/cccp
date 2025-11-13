package generators

var initializers []func()

func Register(initFunc func()) {
	initializers = append(initializers, initFunc)
}

func InitAll() {
	for _, init := range initializers {
		init()
	}
}
