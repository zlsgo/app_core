package common

import (
	"strings"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
)

func Fatal(err error) {
	if err == nil {
		return
	}
	zlog.Fatal(strings.Join(zerror.UnwrapErrors(err), ": "))
}
