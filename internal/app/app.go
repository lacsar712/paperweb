package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/paperweb/internal/wireline"
	"github.com/lacsar712/paperweb/internal/clock"
	"github.com/lacsar712/paperweb/internal/headbox"
	"github.com/lacsar712/paperweb/internal/config"
	"github.com/lacsar712/paperweb/internal/couch"
	"github.com/lacsar712/paperweb/internal/fsm"
	"github.com/lacsar712/paperweb/internal/interlock"
	"github.com/lacsar712/paperweb/internal/model"
	"github.com/lacsar712/paperweb/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.WirelineFSM
	wireline        *wireline.Controller
	headbox    *headbox.Coordinator
	couch          *couch.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	dewaterWindow   *clock.DewaterWindow
	warmupWindow  *clock.HeadboxWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	fiberLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewWirelineFSM(cfg.UnitID),
		wireline:       wireline.NewController(clk),
		headbox:   headbox.NewCoordinator(clk),
		couch:         couch.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		dewaterWindow:  clock.NewDewaterWindow(clk),
		warmupWindow: clock.NewHeadboxWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		fiberLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.WirelineFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetCouch(a.couch.Level().WithinLimits(snap.Couch.LevelPercent))
	a.permissives.SetPressure(a.wireline.Pressure().WithinTripLimits(snap.Wireline.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetHeadbox(a.headbox.Burner().SheetStable(snap.Headbox))
	a.permissives.SetFiber(snap.Headbox.FiberFlowTPH > 0 || snap.State == model.StateDewater)
	a.permissives.SetIgnition(snap.Headbox.BurnerPhase == model.BurnerStable || snap.Headbox.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetFiberPermissive(a.permissives.FiberOK())
	a.fsm.SetDewaterComplete(a.dewaterWindow.Ready(snap.Headbox.DewaterStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
