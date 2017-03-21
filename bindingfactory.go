package hermes

import "github.com/apourchet/hermes/binding"

type BindingFactory func(params, queries []string) binding.Binding

var DefaultBindingFactory BindingFactory = JSONBindingFactory

func JSONBindingFactory(_, _ []string) binding.Binding {
	return &binding.JSONBinding{}
}

func URLBindingFactory(params, queries []string) binding.Binding {
	return &binding.URLBinding{params, queries}
}

func URLThenJSONBindingFactory(params, queries []string) binding.Binding {
	json := &binding.JSONBinding{}
	url := &binding.URLBinding{params, queries}
	return binding.NewSequentialBinding(url, json)
}
