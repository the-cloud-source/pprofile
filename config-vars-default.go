//go:build !pprofile_legacy
// +build !pprofile_legacy

package pprofile

import "os"

var EnableHTTP = "0"
var EnableTmpSocket = "0"
var EnableAbstractSocket = "1"
var ListenHTTP = "localhost:6000"
var TmpSocketTemplate = "{{ .TMP }}" + string(os.PathSeparator) + ".{{ .Executable }}.{{ .PID }}"
var AbstractSocketTemplate = "@{{ .Executable }}"
