package users

type T struct {
	Short  string   `json:"short"`
	Long   string   `json:"long"`
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
		Long:   "User hasn't logged in yet or of unknown type",
		Claims: []string{},
	},
	Viewer: T{
		Short:  "Viewer",
		Long:   "User has logged in",
		Claims: []string{"viewer"},
	},
	Technical: T{
		Short:  "Technical",
		Long:   "User is logged in as Tech Support",
		Claims: []string{"technical", "viewer"},
	},
	Admin: T{
		Short:  "Admin",
		Long:   "User is logged in as Admin and can access everything",
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
