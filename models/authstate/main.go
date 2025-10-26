package authstate

type StateT struct {
	Short  string `json:"short"`
	Long   string `json:"long"`
	Status int8   `json:"status"`
}

var (
	Fresh = StateT{
		Short:  "FRESH",
		Long:   "Session is fresh and usable for sensitive tasks",
		Status: 15,
	}
	Active = StateT{
		Short:  "ACTIVE",
		Long:   "Session is made and usable",
		Status: 5,
	}
	Stale = StateT{
		Short:  "STALE",
		Long:   "Session is stale and might need refresh",
		Status: -5,
	}
	Unknown = StateT{
		Short:  "UNKNOWN",
		Long:   "Unknown or expired session",
		Status: -15,
	}
	SessionBlocked = StateT{
		Short:  "SESSION_BLOCKED",
		Long:   "Session is blocked",
		Status: -25,
	}
	UserBlocked = StateT{
		Short:  "USER_BLOCKED",
		Long:   "User is blocked",
		Status: -35,
	}
)
