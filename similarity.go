package gosbert

import (
	"fmt"
	"path/filepath"
	"runtime"

	python3 "github.com/cel-edward/cpy3"
)

// Sbert is the primary handler for module functionality
type Sbert struct {
	module *python3.PyObject
}

// NewSbert returns a new Sbert struct with initialised Python references,
// returning an error if it fails to initalise correctly.
//
// IMPORTANT: you must call Finalize after NewSbert (defer is idiomatic) to avoid Python/C-API memory leaks
func NewSbert() (s Sbert, err error) {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		return s, fmt.Errorf("NewSbert: failed initialising python interpreter")
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return s, fmt.Errorf("NewSbert: failed finding runtime filepath")
	}
	dir, err := filepath.Abs(filepath.Dir(file)) // os.Args[0]
	if err != nil {
		return s, fmt.Errorf("NewSbert: %w", err)
	}

	ret := python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + dir + "\")")
	if ret != 0 {
		return s, fmt.Errorf("NewSbert: error appending '%s' to python sys.path", dir)
	}

	oImport := python3.PyImport_ImportModule("similarity") //ret val: new ref
	if !(oImport != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return s, fmt.Errorf("NewSbert: failed to import module 'similarity'")
	}

	defer oImport.DecRef()

	oModule := python3.PyImport_AddModule("similarity") //ret val: borrowed ref (from oImport)

	if !(oModule != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return s, fmt.Errorf("NewSbert: failed to add module 'similarity'")
	}

	return Sbert{
		module: oModule,
	}, nil
}

// Finalize should be called once no Python functionality is required.
// Idiomatically this would be with a defer call after NewSbert.
func (s Sbert) Finalize() {
	python3.Py_Finalize()
}

// GetSimilarity returns a similarity metric for each item in others relative to target,
// ranging (approx) from 0-1 where higher means the strings are semantically more similar.
//
// It is based on SBERT embedding of the strings following by a similarity comparison.
func (s Sbert) GetSimilarity(target string, others []string) ([]float64, error) {

	pyTarget := python3.PyUnicode_FromString(target) //retval: New reference, gets stolen later

	pyOthers := python3.PyList_New(len(others)) //retval: New reference, gets stolen later
	for i := 0; i < len(others); i++ {
		item := python3.PyUnicode_FromString(others[i]) //retval: New reference, gets stolen later
		ret := python3.PyList_SetItem(pyOthers, i, item)
		if ret != 0 {
			if python3.PyErr_Occurred() != nil {
				python3.PyErr_Print()
			}
			item.DecRef()
			pyOthers.DecRef()
			return nil, fmt.Errorf("similarity: failed setting list item")
		}
	}

	args := python3.PyTuple_New(2) //retval: New reference
	if args == nil {
		pyTarget.DecRef()
		pyOthers.DecRef()
		return nil, fmt.Errorf("similarity: failed creating args tuple")
	}
	defer args.DecRef()

	ret := python3.PyTuple_SetItem(args, 0, pyTarget) //steals ref to pyTarget
	if ret != 0 {
		if python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
		}
		pyTarget.DecRef()
		pyTarget = nil
		return nil, fmt.Errorf("similarity: failed setting args tuple item 0")
	}
	pyTarget = nil

	ret = python3.PyTuple_SetItem(args, 1, pyOthers) //steals ref to pyOthers
	if ret != 0 {
		if python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
		}
		pyOthers.DecRef()
		pyOthers = nil
		return nil, fmt.Errorf("similarity: failed setting args tuple item 1")
	}
	pyOthers = nil

	oDict := python3.PyModule_GetDict(s.module) //retval: Borrowed
	if !(oDict != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("similarity: could not get dict for module")
	}

	similarity := python3.PyDict_GetItemString(oDict, "similarity")
	if similarity == nil {
		return nil, fmt.Errorf("similarity: could not get function 'similarity', PyDict returned nil")
	}
	if !(similarity != nil && python3.PyCallable_Check(similarity)) { //retval: Borrowed
		return nil, fmt.Errorf("similarity: could not get function 'similarity', non-nil but callable is %t", python3.PyCallable_Check(similarity))
	}
	similarityDataPy := similarity.CallObject(args)
	if !(similarityDataPy != nil && python3.PyErr_Occurred() == nil) { //retval: New reference
		python3.PyErr_Print()
		return nil, fmt.Errorf("similarity: failed calling function 'similarity'")
	}
	defer similarityDataPy.DecRef()
	outliers, err := floatSliceFromPyList(similarityDataPy)
	if err != nil {
		return nil, fmt.Errorf("similarity: %s", err)
	}

	return outliers, nil

}

// floatSliceFromPyList converts a Python list held in pylist to a Go slice of float64
func floatSliceFromPyList(pylist *python3.PyObject) ([]float64, error) {
	seq := pylist.GetIter() //ret val: New reference
	if !(seq != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("floatSliceFromPyList: failed creating iterator for list")
	}
	defer seq.DecRef()
	tNext := seq.GetAttrString("__next__") //ret val: new ref
	if !(tNext != nil && python3.PyCallable_Check(tNext)) {
		return nil, fmt.Errorf("floatSliceFromPyList: iterator has no __next__ function")
	}
	defer tNext.DecRef()

	compare := python3.PyFloat_FromDouble(0)
	if compare == nil {
		return nil, fmt.Errorf("floatSliceFromPyList: failed creating compare var")
	}
	defer compare.DecRef()

	pyType := compare.Type() //ret val: new ref
	if pyType == nil && python3.PyErr_Occurred() != nil {
		python3.PyErr_Print()
		return nil, fmt.Errorf("floatSliceFromPyList: failed getting type of compare var")
	}
	defer pyType.DecRef()

	pyListLen := pylist.Length()
	if pyListLen == -1 {
		return nil, fmt.Errorf("floatSliceFromPyList: failed getting list length")
	}

	goList := make([]float64, pyListLen)
	for i := 0; i < pyListLen; i++ {
		item := tNext.CallObject(nil) //ret val: new ref
		if item == nil && python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
			return nil, fmt.Errorf("floatSliceFromPyList: failed getting next item in sequence")
		}
		itemType := item.Type()
		if itemType == nil && python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
			return nil, fmt.Errorf("floatSliceFromPyList: failed getting item type")
		}

		defer itemType.DecRef()

		if itemType != pyType {
			if item != nil {
				item.DecRef()
			}
			return nil, fmt.Errorf("floatSliceFromPyList: wrong python item type")
		}

		itemGo := python3.PyFloat_AsDouble(item)
		if itemGo != -1 && python3.PyErr_Occurred() == nil {
			goList[i] = itemGo
		} else {
			if item != nil {
				item.DecRef()
			}
			return nil, fmt.Errorf("floatSliceFromPyList: failed casting python value")
		}

		if item != nil {
			item.DecRef()
			item = nil
		}
	}
	return goList, nil
}
