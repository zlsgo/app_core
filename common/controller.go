package common

import (
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/ztype"
)

// VarPages retrieves the page and pagesize values from the given Context.
//
// It takes a Context pointer as a parameter and returns the page and pagesize as integers,
// along with an error. The page and pagesize values are retrieved from the request's form or query parameters.
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

// VarUID returns the UID value from the given Context.
//
// It takes a *znet.Context as a parameter.
// It returns a string value.
func VarUID(c *znet.Context) string {
	id, ok := c.Value("uid", "")
	if !ok {
		return ""
	}
	return ztype.ToString(id)
}
