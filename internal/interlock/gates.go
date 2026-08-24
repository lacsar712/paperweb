package interlock

import (
	"fmt"

	"github.com/lacsar712/paperweb/internal/model"
)

type PermissiveSet struct {
	fiberOK       bool
	ignitionOK   bool
	couchOK       bool
	pressureOK   bool
	headboxOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetFiber(ok bool)       { p.fiberOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetCouch(ok bool)       { p.couchOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetHeadbox(ok bool) { p.headboxOK = ok }

func (p *PermissiveSet) FiberOK() bool       { return p.fiberOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) CouchOK() bool       { return p.couchOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) HeadboxOK() bool { return p.headboxOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.fiberOK && p.ignitionOK && p.couchOK && p.pressureOK && p.headboxOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.fiberOK {
		return fmt.Errorf("%w", model.ErrFiberPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckSheetLoss(reading model.HeadboxReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.WireframeTempF < 600 {
		return fmt.Errorf("%w", model.ErrSheetLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.couchOK {
		return fmt.Errorf("%w", model.ErrCouchLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.headboxOK {
		return fmt.Errorf("%w", model.ErrHeadboxTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
