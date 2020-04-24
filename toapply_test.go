package migrate

import (
	. "gopkg.in/check.v1"
	"sort"
)

var toapplyMigrations = []*Migration{
	&Migration{Id: "abc", Up: nil, Down: nil},
	&Migration{Id: "cde", Up: nil, Down: nil},
	&Migration{Id: "efg", Up: nil, Down: nil},
}

var toapplyMigrationsPatch = []*MigrationPatch{
	{Name: "0001_00_abc", VerInt: 1, PatchInt: 0, Up: nil, Down: nil},
	{Name: "0001_01_cde", VerInt: 1, PatchInt: 1, Up: nil, Down: nil},
	{Name: "0002_00_efg", VerInt: 2, PatchInt: 0, Up: nil, Down: nil},
}

type ToApplyMigrateSuite struct {
}

var _ = Suite(&ToApplyMigrateSuite{})

func (s *ToApplyMigrateSuite) TestGetAll(c *C) {
	toApply := ToApply(toapplyMigrations, "", Up)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[0])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetAbc(c *C) {
	toApply := ToApply(toapplyMigrations, "abc", Up)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrations[1])
	c.Assert(toApply[1], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetCde(c *C) {
	toApply := ToApply(toapplyMigrations, "cde", Up)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
}

func (s *ToApplyMigrateSuite) TestGetDone(c *C) {
	toApply := ToApply(toapplyMigrations, "efg", Up)
	c.Assert(toApply, HasLen, 0)

	toApply = ToApply(toapplyMigrations, "zzz", Up)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownDone(c *C) {
	toApply := ToApply(toapplyMigrations, "", Down)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownCde(c *C) {
	toApply := ToApply(toapplyMigrations, "cde", Down)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrations[1])
	c.Assert(toApply[1], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestDownAbc(c *C) {
	toApply := ToApply(toapplyMigrations, "abc", Down)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestDownAll(c *C) {
	toApply := ToApply(toapplyMigrations, "efg", Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[0])

	toApply = ToApply(toapplyMigrations, "zzz", Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrations[2])
	c.Assert(toApply[1], Equals, toapplyMigrations[1])
	c.Assert(toApply[2], Equals, toapplyMigrations[0])
}

func (s *ToApplyMigrateSuite) TestAlphaNumericMigrations(c *C) {
	var migrations = byId([]*Migration{
		&Migration{Id: "10_abc", Up: nil, Down: nil},
		&Migration{Id: "1_abc", Up: nil, Down: nil},
		&Migration{Id: "efg", Up: nil, Down: nil},
		&Migration{Id: "2_cde", Up: nil, Down: nil},
		&Migration{Id: "35_cde", Up: nil, Down: nil},
	})

	sort.Sort(migrations)

	toApplyUp := ToApply(migrations, "2_cde", Up)
	c.Assert(toApplyUp, HasLen, 3)
	c.Assert(toApplyUp[0].Id, Equals, "10_abc")
	c.Assert(toApplyUp[1].Id, Equals, "35_cde")
	c.Assert(toApplyUp[2].Id, Equals, "efg")

	toApplyDown := ToApply(migrations, "2_cde", Down)
	c.Assert(toApplyDown, HasLen, 2)
	c.Assert(toApplyDown[0].Id, Equals, "2_cde")
	c.Assert(toApplyDown[1].Id, Equals, "1_abc")
}

func (s *ToApplyMigrateSuite) TestGetAllPatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, &MigrationPatch{}, Up)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[0])
	c.Assert(toApply[1], Equals, toapplyMigrationsPatch[1])
	c.Assert(toApply[2], Equals, toapplyMigrationsPatch[2])
}

func (s *ToApplyMigrateSuite) TestGetAbcPatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[0], Up)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[1])
	c.Assert(toApply[1], Equals, toapplyMigrationsPatch[2])
}

func (s *ToApplyMigrateSuite) TestGetCdePatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[1], Up)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[2])
}

func (s *ToApplyMigrateSuite) TestGetDonePatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[2], Up)
	c.Assert(toApply, HasLen, 0)

	toApply = ToApplyPatch(toapplyMigrationsPatch, &MigrationPatch{Name:"0005_05_zzz", VerInt: 4, PatchInt: 0}, Up)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownDonePatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, &MigrationPatch{}, Down)
	c.Assert(toApply, HasLen, 0)
}

func (s *ToApplyMigrateSuite) TestDownCdePatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[1], Down)
	c.Assert(toApply, HasLen, 2)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[1])
	c.Assert(toApply[1], Equals, toapplyMigrationsPatch[0])
}

func (s *ToApplyMigrateSuite) TestDownAbcPatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[0], Down)
	c.Assert(toApply, HasLen, 1)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[0])
}

func (s *ToApplyMigrateSuite) TestDownAllPatch(c *C) {
	toApply := ToApplyPatch(toapplyMigrationsPatch, toapplyMigrationsPatch[2], Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[2])
	c.Assert(toApply[1], Equals, toapplyMigrationsPatch[1])
	c.Assert(toApply[2], Equals, toapplyMigrationsPatch[0])

	toApply = ToApplyPatch(toapplyMigrationsPatch, &MigrationPatch{Name:"0005_05_zzz", VerInt: 4, PatchInt: 0}, Down)
	c.Assert(toApply, HasLen, 3)
	c.Assert(toApply[0], Equals, toapplyMigrationsPatch[2])
	c.Assert(toApply[1], Equals, toapplyMigrationsPatch[1])
	c.Assert(toApply[2], Equals, toapplyMigrationsPatch[0])
}

func (s *ToApplyMigrateSuite) TestAlphaNumericMigrationsPatch(c *C) {
	var migrations = byIdPatch([]*MigrationPatch{
		{Name: "0010_00_abc", VerInt: 10, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0001_00_abc", VerInt: 1, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0005_00_efg", VerInt: 5, PatchInt: 0, Up: nil, Down: nil},
		{Name: "0001_01_cde", VerInt: 1, PatchInt: 1, Up: nil, Down: nil},
		{Name: "0035_00_cde", VerInt: 35, PatchInt: 0, Up: nil, Down: nil},
	})

	sort.Sort(migrations)

	current := &MigrationPatch{Name: "0001_01_cde", VerInt: 1, PatchInt: 1, Up: nil, Down: nil}

	toApplyUp := ToApplyPatch(migrations, current, Up)
	c.Assert(toApplyUp, HasLen, 3)
	c.Assert(toApplyUp[0].Name, Equals, "0005_00_efg")
	c.Assert(toApplyUp[1].Name, Equals, "0010_00_abc")
	c.Assert(toApplyUp[2].Name, Equals, "0035_00_cde")

	toApplyDown := ToApplyPatch(migrations, current, Down)
	c.Assert(toApplyDown, HasLen, 2)
	c.Assert(toApplyDown[0].Name, Equals, "0001_01_cde")
	c.Assert(toApplyDown[1].Name, Equals, "0001_00_abc")
}
