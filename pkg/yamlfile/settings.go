package yamlfile

type UpdateSettingsFunc func(*settings)

type settings struct {
	arrayMergePolicy ArrayMergePolicy
}

func newSettings(opts ...UpdateSettingsFunc) *settings {
	s := settings{
		arrayMergePolicy: Standard,
	}
	for i := range opts {
		opts[i](&s)
	}
	return &s
}

func (s settings) Copy() settings {
	return settings{
		s.arrayMergePolicy,
	}
}

func SetArrayMergePolicy(policy ArrayMergePolicy) UpdateSettingsFunc {
	return func(s *settings) {
		s.arrayMergePolicy = policy
	}
}

type ArrayMergePolicy uint8

const (
	Standard ArrayMergePolicy = 1 << iota
	Strict
)
