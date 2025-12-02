package main

import (
	"context"
	"fmt"
	"hotels-common/models"
	"hotels-common/transformers"
	"net/http"
)

const PluginName = "tbo-request"

func main() {}

func init() {
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
	requestFactory, _ := transformers.GetRequestFactory(models.ProviderTBO)
	f(string(r), requestFactory.BuildRequest(), true, false)
	fmt.Printf("Plugin registered with %s\n", string(r))
	fmt.Println(string(r), "registered!!!")
}
