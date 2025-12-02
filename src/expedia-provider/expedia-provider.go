package main

import (
	"context"
	"hotels-common/core"
	"hotels-common/models"
	"net/http"
)

const PluginName = "expedia-provider"

var pluginCore *core.HotelPluginCore

func main() {}

func init() {
	pluginCore = core.NewHotelPluginCore()
	pluginCore.RegisterTransformer(ExpediaTransformerImpl())
}

type clientRegisterer string
type handlerRegisterer string
type modifierRegisterer string

var ClientRegisterer = clientRegisterer(PluginName)

func (r clientRegisterer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r clientRegisterer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
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

var ModifierRegisterer = modifierRegisterer(PluginName)

func (r modifierRegisterer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), pluginCore.CreateModifierFactory(models.ProviderExpedia), false, true)
}
