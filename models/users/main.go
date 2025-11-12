package users

type T struct {
	Short  string   `json:"short"`
	Claims []string `json:"claims"`
}

type allT struct {
	Unknown   T
	Viewer    T
	Technical T
	Admin     T
}

var All = allT{
	Unknown: T{
		Short:  "Unknown",
		Claims: []string{},
	},
	Viewer: T{
		Short:  "Viewer",
		Claims: []string{"viewer"},
	},
	Technical: T{
		Short:  "Technical",
		Claims: []string{"technical", "viewer"},
	},
	Admin: T{
		Short:  "Admin",
		Claims: []string{"admin", "technical", "viewer"},
	},
}
var _ByName map[string]T

func init() {
	_ByName = map[string]T{
		All.Unknown.Short:   All.Unknown,
		All.Viewer.Short:    All.Viewer,
		All.Technical.Short: All.Technical,
		All.Admin.Short:     All.Admin,
	}
}

func Find(name string) T {
	if val, ok := _ByName[name]; ok {
		return val
	}
	return All.Unknown
}
