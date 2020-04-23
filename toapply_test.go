package migrate

import (
	. "gopkg.in/check.v1"
	"sort"
)

var toapplyMigrations = []*Migration{
	&Migration{Name: "0001_00_abc", verInt: 1, patchInt: 0, Up: nil, Down: nil},
	&Migration{Name: "0001_01_cde", verInt: 1, patchInt: 1, Up: nil, Down: nil},
	&Migration{Name: "0002_00_efg", verInt: 2, patchInt: 0, Up: nil, Down: nil},
}

type ToApplyMigrateSuite struct {
}

var _ = Suite(&ToApplyMigrateSuite{})

func (s *ToApplyMigrateSuite) TestGetAll(c *C) {
	toApply := ToApply(toapplyMigrations, &Migration{}, Up)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[0])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetAbc(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[0], Up)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrations[1])
	c.Assert(toApply[1], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetCde(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[1], Up)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetDone(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[2], Up)
	c.Assert(toApply, HasLen, 0)

	toApply = ToApply(toapplyMigrations, &Migration{Name:"0005_05_zzz", verInt: 4, patchInt: 0}, Up)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownDone(c *C) {
	toApply := ToApply(toapplyMigrations, &Migration{}, Down)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownCde(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[1], Down)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrations[1])
	c.Assert(toApply[1], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestDownAbc(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[0], Down)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestDownAll(c *C) {
	toApply := ToApply(toapplyMigrations, toapplyMigrations[2], Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[0])

	toApply = ToApply(toapplyMigrations, &Migration{Name:"0005_05_zzz", verInt: 4, patchInt: 0}, Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestAlphaNumericMigrations(c *C) {
	var migrations = byId([]*Migration{
		&Migration{Name: "0010_00_abc", verInt: 10, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0001_00_abc", verInt: 1, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0005_00_efg", verInt: 5, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0002_00_cde", verInt: 2, patchInt: 0, Up: nil, Down: nil},
		&Migration{Name: "0035_00_cde", verInt: 35, patchInt: 0, Up: nil, Down: nil},
	})

	sort.Sort(migrations)

	current := &Migration{Name: "0002_00_cde", verInt: 2, patchInt: 0, Up: nil, Down: nil}

	toApplyUp := ToApply(migrations, current, Up)
	c.Assert(toApplyUp, HasLen, 3)
	c.Assert(toApplyUp[0].Name, Equals, "0005_00_efg")
	c.Assert(toApplyUp[1].Name, Equals, "0010_00_abc")
	c.Assert(toApplyUp[2].Name, Equals, "0035_00_cde")

	toApplyDown := ToApply(migrations, current, Down)
	c.Assert(toApplyDown, HasLen, 2)
	c.Assert(toApplyDown[0].Name, Equals, "0002_00_cde")
	c.Assert(toApplyDown[1].Name, Equals, "0001_00_abc")
}
