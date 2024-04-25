package gosbert

import (
	"fmt"
	"os"
	"path/filepath"

	python3 "github.com/cel-edward/cpy3"
)

// func main() {

// 	defer python3.Py_Finalize()
// 	python3.Py_Initialize()
// 	if !python3.Py_IsInitialized() {
// 		log.Fatal("failed initialising python interpreter")
// 	}

// 	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// we could also use PySys_GetObject("path") + PySys_SetPath,
// 	//but this is easier (at the cost of less flexible error handling)
// 	ret := python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + dir + "\")")
// 	if ret != 0 {
// 		log.Fatalf("error appending '%s' to python sys.path", dir)
// 	}

// 	oImport := python3.PyImport_ImportModule("similarity") //ret val: new ref
// 	if !(oImport != nil && python3.PyErr_Occurred() == nil) {
// 		python3.PyErr_Print()
// 		log.Fatal("failed to import module 'similarity'")
// 	}

// 	defer oImport.DecRef()

// 	oModule := python3.PyImport_AddModule("similarity") //ret val: borrowed ref (from oImport)

// 	if !(oModule != nil && python3.PyErr_Occurred() == nil) {
// 		python3.PyErr_Print()
// 		log.Fatal("failed to add module 'similarity'")
// 	}

// 	err = runPython(oModule)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

type Sbert struct {
	module *python3.PyObject
}

func NewSbert() (s Sbert, err error) {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		return s, fmt.Errorf("NewSbert: failed initialising python interpreter")
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return s, fmt.Errorf("NewSbert: %w", err)
	}

	fmt.Println(dir)

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

func (s Sbert) Finalize() {
	python3.Py_Finalize()
}

// func runPython(module *python3.PyObject) error {
// 	s := "I am a test string"
// 	ss := []string{
// 		"I am also a testing string",
// 		"there is nothing in common with what I say",
// 		"equally nothing of equivalence here",
// 	}
// 	similarities, err := similarity(module, s, ss)
// 	if err != nil {
// 		return fmt.Errorf("runPython: %w", err)
// 	}
// 	fmt.Println(similarities)
// 	return nil
// }

func (s Sbert) Similarity(target string, others []string) ([]float64, error) {

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

// func intSliceFromPyList(pylist *python3.PyObject) ([]int, error) {

// 	seq := pylist.GetIter() //ret val: New reference
// 	if !(seq != nil && python3.PyErr_Occurred() == nil) {
// 		python3.PyErr_Print()
// 		return nil, fmt.Errorf("intSliceFromPyList: failed creating iterator for list")
// 	}
// 	defer seq.DecRef()
// 	tNext := seq.GetAttrString("__next__") //ret val: new ref
// 	if !(tNext != nil && python3.PyCallable_Check(tNext)) {
// 		return nil, fmt.Errorf("intSliceFromPyList: iterator has no __next__ function")
// 	}
// 	defer tNext.DecRef()

// 	compare := python3.PyLong_FromGoInt(0)
// 	if compare == nil {
// 		return nil, fmt.Errorf("intSliceFromPyList: failed creating compare var")
// 	}
// 	defer compare.DecRef()

// 	pyType := compare.Type() //ret val: new ref
// 	if pyType == nil && python3.PyErr_Occurred() != nil {
// 		python3.PyErr_Print()
// 		return nil, fmt.Errorf("intSliceFromPyList: failed getting type of compare var")
// 	}
// 	defer pyType.DecRef()

// 	pyListLen := pylist.Length()
// 	if pyListLen == -1 {
// 		return nil, fmt.Errorf("intSliceFromPyList: failed getting list length")
// 	}

// 	goList := make([]int, pyListLen)
// 	for i := 0; i < pyListLen; i++ {
// 		item := tNext.CallObject(nil) //ret val: new ref
// 		if item == nil && python3.PyErr_Occurred() != nil {
// 			python3.PyErr_Print()
// 			return nil, fmt.Errorf("intSliceFromPyList: failed getting next item in sequence")
// 		}
// 		itemType := item.Type()
// 		if itemType == nil && python3.PyErr_Occurred() != nil {
// 			python3.PyErr_Print()
// 			return nil, fmt.Errorf("intSliceFromPyList: failed getting item type")
// 		}

// 		defer itemType.DecRef()

// 		if itemType != pyType {
// 			if item != nil {
// 				item.DecRef()
// 			}
// 			return nil, fmt.Errorf("intSliceFromPyList: wrong python item type")
// 		}

// 		itemGo := python3.PyLong_AsLong(item)
// 		if itemGo != -1 && python3.PyErr_Occurred() == nil {
// 			goList[i] = itemGo
// 		} else {
// 			if item != nil {
// 				item.DecRef()
// 			}
// 			return nil, fmt.Errorf("intSliceFromPyList: failed casting python value")
// 		}

// 		if item != nil {
// 			item.DecRef()
// 			item = nil
// 		}
// 	}
// 	return goList, nil
// }
