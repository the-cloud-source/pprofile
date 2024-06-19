package pprofile

var FailOnError = ""
var DebugLog = ""

func fail() bool {
	return (FailOnError != "")
}

func dbglog() bool {
	return (DebugLog != "")
}
