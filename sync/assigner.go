package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/micromdm/nanodep/godep"

	"github.com/micromdm/nanolib/log"
	"github.com/micromdm/nanolib/log/ctxlog"
)

type AssignerProfileRetriever interface {
	// RetrieveAssignerProfile retrieves the assigner profile UUID and timestamp for name (DEP name).
	// If name or assigner profile UUID do not exist returns an empty profile and nil error.
	RetrieveAssignerProfile(ctx context.Context, name string) (profileUUID string, modTime time.Time, err error)
}

// Assigner assigns devices synced from the Apple DEP APIs to a profile UUID.
type Assigner struct {
	client *godep.Client
	name   string
	store  AssignerProfileRetriever
	logger log.Logger
	debug  bool
}

type AssignerOption func(*Assigner)

// NewAssigner creates a new Assigner from client and uses store to lookup
// assigner profile UUIDs. DEP name is specified with name.
func NewAssigner(client *godep.Client, name string, store AssignerProfileRetriever, opts ...AssignerOption) *Assigner {
	assigner := &Assigner{
		client: client,
		name:   name,
		store:  store,
		logger: log.NopLogger,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(assigner)
		}
	}
	assigner.logger = assigner.logger.With("name", assigner.name)
	return assigner
}

// WithAssignerLogger configures logger for the assigner.
func WithAssignerLogger(logger log.Logger) AssignerOption {
	return func(a *Assigner) {
		a.logger = logger
	}
}

// WithAssignerDebug enables additional assigner-specific debug logging for troubleshooting.
func WithAssignerDebug() AssignerOption {
	return func(a *Assigner) {
		a.debug = true
	}
}

// ProcessDeviceResponse processes the device response from the device sync
// DEP API endpoints and assigns the profile UUID associated with the DEP
// client DEP name.
func (a *Assigner) ProcessDeviceResponse(ctx context.Context, resp *godep.FetchDeviceResponseJson) error {
	if len(resp.Devices) < 1 {
		// no devices means we can't assign anything
		return nil
	}
	profileUUID, _, err := a.store.RetrieveAssignerProfile(ctx, a.name)
	if err != nil {
		return fmt.Errorf("retrieve profile: %w", err)
	}
	logger := ctxlog.Logger(ctx, a.logger)
	if profileUUID == "" {
		// empty UUID means we can't assign anything
		if a.debug {
			// the user could simply have not setup an assigner profile
			// UUID yet. so hide this debug log behind the 'extra' debug
			// setting to avoid unnecessary cause for concern.
			logger.Debug("msg", "empty assigner profile UUID")
		}
		return nil
	}

	var serialsToAssign []string
	for _, device := range resp.Devices {
		if a.debug {
			logs := []interface{}{"msg", "device"}
			logs = append(logs, logDevice(device)...)
			logger.Debug(logs...)
		}
		// note that we may see multiple serial number "events"
		if shouldAssignDevice(device) {
			serialsToAssign = append(serialsToAssign, deref(device.SerialNumber))
		}
	}

	logger = logger.With("profile_uuid", profileUUID)

	if len(serialsToAssign) < 1 {
		if a.debug {
			logger.Debug(
				"msg", "no serials to assign",
				"devices", len(resp.Devices),
			)
		}
		return nil
	}

	apiResp, err := a.client.AssignProfile(ctx, a.name, profileUUID, serialsToAssign...)
	if err != nil {
		logger.Info(
			"msg", "assign profile",
			"devices", len(serialsToAssign),
			"err", err,
		)
		return fmt.Errorf("assign profile: %w", err)
	}

	logs := []interface{}{
		"msg", "profile assigned",
		"devices", len(serialsToAssign),
	}
	logs = append(logs, logCountsForResults(apiResp.Devices)...)
	logger.Info(logs...)

	return nil
}

// shouldAssignDevice decides whether a device "event" should be passed
// off to the assigner.
func shouldAssignDevice(device godep.DeviceJson) bool {
	// we currently only listen for an op_type of "added." the other
	// op_types are ambiguous and it would be needless to assign the
	// profile UUID every single time we get an update.
	if strings.ToLower(string(deref(device.OpType))) == "added" {
		return true
	}
	return false
}

// logCountsForResults tries to aggregate the result types and log the counts.
func logCountsForResults(deviceResults map[string]string) (out []interface{}) {
	results := map[string]int{"success": 0, "not_accessible": 0, "failed": 0, "other": 0}
	for _, result := range deviceResults {
		l := strings.ToLower(result)
		if _, ok := results[l]; !ok {
			l = "other"
		}
		results[l] += 1
	}
	for k, v := range results {
		if v > 0 {
			out = append(out, k, v)
		}
	}
	return
}
