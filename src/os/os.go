package os

var Args []string

func runtime_args() []string

func init() {
	Args = runtime_args()
}

func Exit(status int)

