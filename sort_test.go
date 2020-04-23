package migrate

import (
	. "gopkg.in/check.v1"
	"sort"
)

type SortSuite struct{}

var _ = Suite(&SortSuite{})

func (s *SortSuite) TestSortMigrations(c *C) {
	var migrations = byId([]*Migration{
		&Migration{Name: "0010_00_abc", verInt: 10, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0120_00_cde", verInt: 120, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0001_01_cde", verInt: 1, patchInt: 1, Up: nil, Down: nil},
		&Migration{Name: "0999_00_efg", verInt: 999, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0001_00_abc", verInt: 1, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0035_00_cde", verInt: 35, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0003_00_efg", verInt: 3, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0004_00_abc", verInt: 4, patchInt: 0, Up: nil, Down: nil},
	})

	sort.Sort(migrations)
	c.Assert(migrations, HasLen, 8)
	c.Assert(migrations[0].Name, Equals, "0001_00_abc")
	c.Assert(migrations[1].Name, Equals, "0001_01_cde")
	c.Assert(migrations[2].Name, Equals, "0003_00_efg")
	c.Assert(migrations[3].Name, Equals, "0004_00_abc")
	c.Assert(migrations[4].Name, Equals, "0010_00_abc")
	c.Assert(migrations[5].Name, Equals, "0035_00_cde")
	c.Assert(migrations[6].Name, Equals, "0120_00_cde")
	c.Assert(migrations[7].Name, Equals, "0999_00_efg")
}
