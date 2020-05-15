package migrate

import (
	. "gopkg.in/check.v1"
	"sort"
)

type SortSuite struct{}

var _ = Suite(&SortSuite{})

func (s *SortSuite) TestSortMigrations(c *C) {
	var migrations = byId([]*Migration{
		&Migration{Id: "10_abc", Up: nil, Down: nil},
		&Migration{Id: "120_cde", Up: nil, Down: nil},
		&Migration{Id: "1_abc", Up: nil, Down: nil},
		&Migration{Id: "efg", Up: nil, Down: nil},
		&Migration{Id: "2_cde", Up: nil, Down: nil},
		&Migration{Id: "35_cde", Up: nil, Down: nil},
		&Migration{Id: "3_efg", Up: nil, Down: nil},
		&Migration{Id: "4_abc", Up: nil, Down: nil},
	})

	sort.Sort(migrations)
	c.Assert(migrations, HasLen, 8)
	c.Assert(migrations[0].Id, Equals, "1_abc")
	c.Assert(migrations[1].Id, Equals, "2_cde")
	c.Assert(migrations[2].Id, Equals, "3_efg")
	c.Assert(migrations[3].Id, Equals, "4_abc")
	c.Assert(migrations[4].Id, Equals, "10_abc")
	c.Assert(migrations[5].Id, Equals, "35_cde")
	c.Assert(migrations[6].Id, Equals, "120_cde")
	c.Assert(migrations[7].Id, Equals, "efg")
}

func (s *SortSuite) TestSortMigrationsPatch(c *C) {
	var migrations = byIdPatch([]*MigrationPatch{
		{Name: "0010_00_abc", VerInt: 10, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0120_00_cde", VerInt: 120, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0001_01_cde", VerInt: 1, PatchInt: 1, Up: nil, Down: nil},
		{Name: "0999_00_efg", VerInt: 999, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0001_00_abc", VerInt: 1, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0035_78_cde", VerInt: 35, PatchInt: 78, Up: nil, Down: nil},
		{Name: "0003_00_efg", VerInt: 3, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0004_00_abc", VerInt: 4, PatchInt: 0, Up: nil, Down: nil},
	})

	sort.Sort(migrations)
	c.Assert(migrations, HasLen, 8)
	c.Assert(migrations[0].Name, Equals, "0001_00_abc")
	c.Assert(migrations[1].Name, Equals, "0001_01_cde")
	c.Assert(migrations[2].Name, Equals, "0003_00_efg")
	c.Assert(migrations[3].Name, Equals, "0004_00_abc")
	c.Assert(migrations[4].Name, Equals, "0010_00_abc")
	c.Assert(migrations[5].Name, Equals, "0035_78_cde")
	c.Assert(migrations[6].Name, Equals, "0120_00_cde")
	c.Assert(migrations[7].Name, Equals, "0999_00_efg")
}
