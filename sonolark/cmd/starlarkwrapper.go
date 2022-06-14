package cmd

import (
	"fmt"

	"github.com/k14s/starlark-go/starlark"
	nethttp "github.com/pcj/starlark-go-nethttp"
	starlarkcore "go.starlark.net/starlark"
)

type Module struct {
	netMod *nethttp.Module
}

// toYttStarlark converts the nethttp module to a ytt-starlark compatible one.
// If we get more of these sorts of conversions, a more reusable pattern will probably
// emerge.
func toStarlarkModule(input *nethttp.Module) *Module {
	return &Module{input}
}

func (mod *Module) String() string       { return "<module http>" }
func (mod *Module) Type() string         { return "module" }
func (mod *Module) Freeze()              { mod.netMod.Freeze() }
func (mod *Module) Truth() starlark.Bool { return starlark.True }
func (mod *Module) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: %s", mod.Type())
}

func (mod *Module) Attr(name string) (starlark.Value, error) {
	val, err := mod.netMod.Attr(name)
	if coreCallable, isCallable := val.(starlarkcore.Callable); isCallable {
		return &yttStarlarkCallableWrapper{coreCallable}, err
	}
	return &yttStarlarkValueWrapper{val}, err
}

func (mod *Module) AttrNames() []string {
	return mod.netMod.AttrNames()
}

// yttStarlarkValueWrapper wraps a type of the core starlark type but translates the Bool type so that
// it meets the necessary interfaces for the ytt-fork of starlark.
type yttStarlarkValueWrapper struct {
	v starlarkcore.Value
}

func (wrapper *yttStarlarkValueWrapper) String() string { return wrapper.v.String() }
func (wrapper *yttStarlarkValueWrapper) Type() string   { return wrapper.v.Type() }
func (wrapper *yttStarlarkValueWrapper) Freeze()        { wrapper.v.Freeze() }
func (wrapper *yttStarlarkValueWrapper) Truth() starlark.Bool {
	if wrapper.v.Truth() == starlarkcore.True {
		return starlark.True
	}
	return starlark.False
}
func (wrapper *yttStarlarkValueWrapper) Hash() (uint32, error) {
	return wrapper.v.Hash()
}

// yttStarlarkValueWrapper wraps a type of the core starlark type but translates the Bool type so that
// it meets the necessary interfaces for the ytt-fork of starlark.
type yttStarlarkValueWrapperToCore struct {
	v starlark.Value
}

func (wrapper *yttStarlarkValueWrapperToCore) String() string { return wrapper.v.String() }
func (wrapper *yttStarlarkValueWrapperToCore) Type() string   { return wrapper.v.Type() }
func (wrapper *yttStarlarkValueWrapperToCore) Freeze()        { wrapper.v.Freeze() }
func (wrapper *yttStarlarkValueWrapperToCore) Truth() starlarkcore.Bool {
	if wrapper.v.Truth() == starlark.True {
		return starlarkcore.True
	}
	return starlarkcore.False
}
func (wrapper *yttStarlarkValueWrapperToCore) Hash() (uint32, error) {
	return wrapper.v.Hash()
}

type yttStarlarkCallableWrapper struct {
	v starlarkcore.Callable
}

func (wrapper *yttStarlarkCallableWrapper) Value() string { return wrapper.v.String() }
func (wrapper *yttStarlarkCallableWrapper) Name() string  { return wrapper.v.Name() }
func (wrapper *yttStarlarkCallableWrapper) CallInternal(
	thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	coreThread := starlarkcore.Thread{
		Name: thread.Name,
		Print: func(thread *starlarkcore.Thread, msg string) {
			thread.Print(thread, msg)
		},
		Load: func(thread *starlarkcore.Thread, module string) (starlarkcore.StringDict, error) {
			return thread.Load(thread, module)
		},
	}

	v, err := wrapper.v.CallInternal(
		&coreThread,
		toCoreTuple(args),
		toCoreTupleSlice(kwargs),
	)
	return &yttStarlarkValueWrapper{v}, err
}

func (wrapper *yttStarlarkCallableWrapper) String() string { return wrapper.v.String() }
func (wrapper *yttStarlarkCallableWrapper) Type() string   { return wrapper.v.Type() }
func (wrapper *yttStarlarkCallableWrapper) Freeze()        { wrapper.v.Freeze() }
func (wrapper *yttStarlarkCallableWrapper) Truth() starlark.Bool {
	if wrapper.v.Truth() == starlarkcore.True {
		return starlark.True
	}
	return starlark.False
}
func (wrapper *yttStarlarkCallableWrapper) Hash() (uint32, error) {
	return wrapper.v.Hash()
}

func toCoreTuple(input starlark.Tuple) starlarkcore.Tuple {
	result := []starlarkcore.Value{}
	for _, v := range input {
		switch t := v.(type) {
		case starlark.String:
			result = append(result, starlarkcore.String(t))
		default:
			result = append(result, &yttStarlarkValueWrapperToCore{v})
		}
	}
	return result
}

func toCoreTupleSlice(input []starlark.Tuple) []starlarkcore.Tuple {
	result := []starlarkcore.Tuple{}
	for _, v := range input {
		result = append(result, toCoreTuple(v))
	}
	return result
}

type yttStarlarkThreadWrapper struct {
	v starlarkcore.Thread
}

type yttStarlarkTupleWrapper struct {
	v starlarkcore.Tuple
}
