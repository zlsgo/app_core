package utils

import (
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/ztype"
)

func VarPages(c *znet.Context) (page, pagesize int, err error) {
	p := c.DefaultFormOrQuery("page", "1")
	s := c.DefaultFormOrQuery("pagesize", "10")
	page = ztype.ToInt(p)
	if page < 1 {
		page = 1
	}
	pagesize = ztype.ToInt(s)
	if pagesize < 1 {
		pagesize = 10
	}
	return
}
