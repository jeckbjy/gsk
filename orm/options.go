package orm

import "github.com/jeckbjy/gsk/orm/driver"

type OpenOption func(options *driver.OpenOptions)
type IndexOption func(options *driver.IndexOptions)
type InsertOption func(options *driver.InsertOptions)
type DeleteOption func(options *driver.DeleteOptions)
type UpdateOption func(options *driver.UpdateOptions)
type QueryOption func(options *driver.QueryOptions)
