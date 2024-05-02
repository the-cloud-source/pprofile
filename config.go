package pprofile

import "os"

var EnableHTTP = "0"
var EnableTmpSocket = "1"
var EnableAbstractSocket = "1"
var ListenHTTP = "localhost:6000"
var TmpSocketTemplate = "{{ .TMP }}" + string(os.PathSeparator) + ".go_pid{{ .PID }}"
var AbstractSocketTemplate = "@{{ .Executable }}"

var FailOnError = ""
var DebugLog = ""

func fail() bool {
	return (FailOnError != "")
}

func dbglog() bool {
	return (DebugLog != "")
}
