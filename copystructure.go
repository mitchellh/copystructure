package copystructure

import (
	"fmt"
	"reflect"

	"github.com/mitchellh/reflectwalk"
)

// Copy returns a deep copy of v.
func Copy(v interface{}) (interface{}, error) {
	w := new(walker)
	err := reflectwalk.Walk(v, w)
	if err != nil {
		return nil, err
	}

	return w.Result, nil
}

type walker struct {
	Result interface{}

	current reflect.Value
	mapkey  reflect.Value
	locs    []reflectwalk.Location
	vals    []reflect.Value
	maps    []reflect.Value
}

func (w *walker) Enter(l reflectwalk.Location) error {
	// Push the last location so we know the latest of where we are
	w.locs = append(w.locs, l)

	return nil
}

func (w *walker) Exit(l reflectwalk.Location) error {
	// Pop off the last location so we're accurate
	w.locs = w.locs[:len(w.locs)-1]

	switch l {
	case reflectwalk.Map:
		// Pop map off our map list
		w.maps = w.maps[:len(w.maps)-1]
	case reflectwalk.MapValue:
		// Pop off the key and value
		mv := w.valPop()
		mk := w.valPop()
		m := w.maps[len(w.maps)-1]
		m.SetMapIndex(mk, mv)
	case reflectwalk.WalkLoc:
		// If we exited the walk location, then we're done walking, and
		// we need to make sure we set a result.
		if w.Result == nil && len(w.vals) > 0 {
			w.Result = w.vals[0].Interface()
		}

		// Clear out the slices for GC
		w.locs = nil
		w.vals = nil
	}

	return nil
}

func (w *walker) Map(m reflect.Value) error {
	t := m.Type()
	newMap := reflect.MakeMap(reflect.MapOf(t.Key(), t.Elem()))
	w.maps = append(w.maps, newMap)
	w.valPush(newMap)
	return nil
}

func (w *walker) MapElem(m, k, v reflect.Value) error {
	return nil
}

func (w *walker) Primitive(v reflect.Value) error {
	w.valPush(v)
	return nil
}

func (w *walker) valPop() reflect.Value {
	result := w.vals[len(w.vals)-1]
	println(fmt.Sprintf("POP: %s", result))
	w.vals = w.vals[:len(w.vals)-1]
	return result
}

func (w *walker) valPush(v reflect.Value) {
	println(fmt.Sprintf("PUSH: %s", v))
	w.vals = append(w.vals, v)
}
