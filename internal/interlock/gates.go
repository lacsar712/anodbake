package interlock

import (
	"fmt"

	"github.com/lacsar712/anodbake/internal/model"
)

type PermissiveSet struct {
	pitchOK       bool
	ignitionOK   bool
	anodepitOK       bool
	pressureOK   bool
	pitfireOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetPitch(ok bool)       { p.pitchOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetAnodepit(ok bool)       { p.anodepitOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetPitfire(ok bool) { p.pitfireOK = ok }

func (p *PermissiveSet) PitchOK() bool       { return p.pitchOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) AnodepitOK() bool       { return p.anodepitOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) PitfireOK() bool { return p.pitfireOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.pitchOK && p.ignitionOK && p.anodepitOK && p.pressureOK && p.pitfireOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.pitchOK {
		return fmt.Errorf("%w", model.ErrPitchPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckBakeLoss(reading model.PitfireReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.PitframeTempF < 600 {
		return fmt.Errorf("%w", model.ErrBakeLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.anodepitOK {
		return fmt.Errorf("%w", model.ErrAnodepitLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.pitfireOK {
		return fmt.Errorf("%w", model.ErrPitfireTrip)
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
