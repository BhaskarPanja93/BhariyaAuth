package users

// User types allowed

type T struct {
	Short     string `json:"short"`
	Authority uint8  `json:"authority"`
}

type allT struct {
	Unknown   T
	Viewer    T
	Moderator T
	Admin     T
	Owner     T
}

var All = allT{
	Unknown: T{
		Short:     "U",
		Authority: 0,
	},
	Viewer: T{
		Short:     "V",
		Authority: 1,
	},
	Moderator: T{
		Short:     "M",
		Authority: 100,
	},
	Admin: T{
		Short:     "A",
		Authority: 200,
	},
	Owner: T{
		Short:     "O",
		Authority: 255,
	},
}

var _ByName map[string]T

func init() {
	_ByName = map[string]T{
		All.Unknown.Short:   All.Unknown,
		All.Viewer.Short:    All.Viewer,
		All.Moderator.Short: All.Moderator,
		All.Admin.Short:     All.Admin,
		All.Owner.Short:     All.Owner,
	}
}

func Find(name string) T {
	if val, ok := _ByName[name]; ok {
		return val
	}
	return All.Unknown
}
