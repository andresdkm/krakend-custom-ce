package main

import (
	"context"
	"kraken-builder-plugins/pkg/hotels/common/core"
	"kraken-builder-plugins/pkg/hotels/common/models"
	"net/http"
)

const PluginName = "tbo-provider"

var pluginCore *core.HotelPluginCore

func main() {}

func init() {
	pluginCore = core.NewHotelPluginCore()

	pluginCore.RegisterTransformer(NewTBOTransformer())
}

type registerer string
type handlerRegisterer string

var ClientRegisterer = registerer(PluginName)

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil
}

var HandlerRegisterer = handlerRegisterer(PluginName)

func (hr handlerRegisterer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(hr), hr.registerHandlers)
}

func (hr handlerRegisterer) registerHandlers(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), nil
}

var ModifierRegisterer = registerer(PluginName)

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), pluginCore.CreateModifierFactory(models.ProviderTBO), false, true)
}
