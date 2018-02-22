package app

import (
	"fmt"

	"github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/TIBCOSoftware/flogo-lib/core/trigger"
)

//InstanceHelper helps to create the instances for a given id
type InstanceHelper struct {
	app        *Config
	tFactories map[string]trigger.Factory
	aFactories map[string]action.Factory
}

//NewInstanceManager creates a new instance manager
func NewInstanceHelper(app *Config, tFactories map[string]trigger.Factory, aFactories map[string]action.Factory) *InstanceHelper {
	return &InstanceHelper{app: app, tFactories: tFactories, aFactories: aFactories}
}

//CreateTriggers creates new instances for triggers in the configuration
func (h *InstanceHelper) CreateTriggers() (map[string]*trigger.TriggerInstance, error) {

	// Get Trigger instances from configuration
	tConfigs := h.app.Triggers

	instances := make(map[string]*trigger.TriggerInstance, len(tConfigs))

	for _, tConfig := range tConfigs {
		if tConfig == nil {
			continue
		}

		_, ok := instances[tConfig.Id]
		if ok {
			return nil, fmt.Errorf("Trigger with id '%s' already registered, trigger ids have to be unique", tConfig.Id)
		}

		factory, ok := h.tFactories[tConfig.Ref]
		if !ok {
			return nil, fmt.Errorf("Trigger Factory '%s' not registered", tConfig.Ref)
		}

		newInterface := factory.New(tConfig)

		if newInterface == nil {
			return nil, fmt.Errorf("Cannot create Trigger nil for id '%s'", tConfig.Id)
		}

		tConfig.FixUp(newInterface.Metadata())

		instances[tConfig.Id] = &trigger.TriggerInstance{Config: tConfig, Interf: newInterface}
	}

	return instances, nil
}

//CreateActions creates new instances for actions in the configuration
func (h *InstanceHelper) CreateActions() (map[string]action.Action, error) {

	// Get Action instances from configuration
	aConfigs := h.app.Actions

	actions := make(map[string]action.Action, len(aConfigs))

	for _, aConfig := range aConfigs {
		if aConfig == nil {
			continue
		}

		_, ok := actions[aConfig.Id]
		if ok {
			return nil, fmt.Errorf("Action with id '%s' already registered, action ids have to be unique", aConfig.Id)
		}

		factory, ok := h.aFactories[aConfig.Ref]
		if !ok {
			return nil, fmt.Errorf("Action Factory '%s' not registered", aConfig.Ref)
		}

		newAction := factory.New(aConfig)

		if newAction == nil {
			return nil, fmt.Errorf("Cannot create Action nil for id '%s'", aConfig.Id)
		}

		actions[aConfig.Id] = newAction
	}

	return actions, nil
}

func CreateTriggers(tConfigs []*trigger.Config, runner action.Runner) (map[string]trigger.Trigger, error) {

	triggers := make(map[string]trigger.Trigger, len(tConfigs))
	legacyRunner := trigger.NewLegacyRunner(runner)

	for _, tConfig := range tConfigs {

		_, exists := triggers[tConfig.Id]
		if exists {
			return nil, fmt.Errorf("Trigger with id '%s' already registered, trigger ids have to be unique", tConfig.Id)
		}

		factory := trigger.GetFactory(tConfig.Ref)

		if factory == nil {
			return nil, fmt.Errorf("Trigger Factory '%s' not registered", tConfig.Ref)
		}

		tgr := factory.New(tConfig)

		if tgr == nil {
			return nil, fmt.Errorf("Cannot create Trigger nil for id '%s'", tConfig.Id)
		}

		initCtx := &initContext{handlers:make([]*trigger.Handler, 0, len(tConfig.Handlers))}

		//create handlers for that trigger and init
		for _, hConfig := range tConfig.Handlers {

			h := trigger.NewHandler(hConfig, nil, tgr.Metadata().Output, tgr.Metadata().Reply, runner)
			initCtx.handlers = append(initCtx.handlers, h)
			trigger.RegisterHandler(hConfig, h)
		}

		newTrg, ok := tgr.(trigger.Init)
		if ok {

			err := newTrg.Initialize(initCtx)
			if err != nil {
				return nil, err
			}
		} else {

			tgr.Init(legacyRunner)
		}
	}

	return triggers, nil
}

type initContext struct {
	handlers []*trigger.Handler
}

func (ctx *initContext) GetHandlers() ([]*trigger.Handler) {
	return ctx.handlers
}


