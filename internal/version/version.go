package version

var version = "development"
var commit = "0000000000000000000000000000000000000000"

func Version() string {
	return version
}

func Commit() string {
	return commit
}
