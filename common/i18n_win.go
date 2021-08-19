// +build windows
// +build go1.9

package common

import (
	"golang.org/x/sys/windows"
)

func init(){
	lang, _ = windows.GetUserPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
}