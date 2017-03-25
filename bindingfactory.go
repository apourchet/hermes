package hermes

import "github.com/apourchet/hermes/binding"

type BindingFactory func(params, queries []string, headers map[string]string) binding.Binding

var DefaultBindingFactory BindingFactory = URLThenJSONBindingFactory

func JSONBindingFactory(_, _ []string, _ map[string]string) binding.Binding {
	return &binding.JSONBinding{}
}

func URLBindingFactory(params, queries []string, _ map[string]string) binding.Binding {
	return &binding.URLBinding{params, queries}
}

func URLThenJSONBindingFactory(params, queries []string, headers map[string]string) binding.Binding {
	header := &binding.HeaderBinding{headers}
	url := &binding.URLBinding{params, queries}
	json := &binding.JSONBinding{}
	return binding.NewSequentialBinding(header, url, json)
}
