package ui

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/jellydn/mole-tui/internal/cleanup"
	"github.com/jellydn/mole-tui/internal/mo"
	"github.com/jellydn/mole-tui/internal/scanner"
)

// operationController owns the UI-facing lifecycle of Scan and Cleanup
// sessions. It translates plain package events into Bubble Tea messages while
// leaving screen state and rendering to Model.
type operationController struct {
	dryRun bool
	runner mo.Runner

	scanCancel    context.CancelFunc
	cleanupCancel context.CancelFunc
	cleanupEvents <-chan cleanup.Event
}

func newOperationController(dryRun bool, runner mo.Runner) operationController {
	return operationController{dryRun: dryRun, runner: runner}
}

func (o *operationController) startScan(sudo bool) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	o.scanCancel = cancel
	return o.scanCmd(ctx, sudo)
}

func (o *operationController) scanCmd(ctx context.Context, sudo bool) tea.Cmd {
	return func() tea.Msg {
		result, err := scanner.Scan(ctx, o.runner, sudo)
		if errors.Is(err, context.Canceled) {
			return scanCancelledMsg{}
		}
		return scanCompleteMsg{result: result, err: err}
	}
}

func (o *operationController) startCleanup(sudo bool) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	o.cleanupCancel = cancel
	o.cleanupEvents = cleanup.NewSession(
		ctx,
		cleanup.Options{DryRun: o.dryRun, Sudo: sudo},
		o.runner,
	).Start().Events
	return o.readCleanupStreamCmd()
}

func (o *operationController) readCleanupStreamCmd() tea.Cmd {
	events := o.cleanupEvents
	return func() tea.Msg {
		if events == nil {
			return nil
		}
		event, ok := <-events
		if !ok {
			return nil
		}
		if event.Kind == cleanup.EventDone {
			return cleanupCompleteMsg{result: event.Done.Result, err: event.Done.Err}
		}
		return cleanupStreamMsg{line: event.Line}
	}
}

func (o *operationController) cancelScan() bool {
	if o.scanCancel == nil {
		return false
	}
	o.scanCancel()
	return true
}

func (o *operationController) cancelCleanup() bool {
	if o.cleanupCancel == nil {
		return false
	}
	o.cleanupCancel()
	return true
}

func (o *operationController) clearScan() { o.scanCancel = nil }

func (o *operationController) clearCleanup() {
	o.cleanupEvents = nil
	o.cleanupCancel = nil
}
