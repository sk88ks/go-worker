package worker

import "reflect"

type (
	// Manager is manager for a workers
	Manager struct {
		count         int
		processNum    int
		workerNum     int
		forceStop     bool
		start         chan int
		stop          chan int
		In            chan *Process
		Out           chan *Process
		FailFilter    []FilterFunc
		SuccessFilter []FilterFunc
		NotExec       []string
	}

	// Process has an implementation
	// for worker and information about process
	Process struct {
		ID       string
		Function ProcessFunc
		Result   interface{}
		Error    error
	}

	// ProcessFunc is function executed by individual worker
	ProcessFunc func() (interface{}, error)

	// FilterFunc is function executed at middle of process as filter
	FilterFunc func(*Process)
)

// Wrap func to be used as WorkerFunc
// Can use only function as ex
// ex
// func(.....) {} // return nothing
// func(.....) error // return only error
// func(.....) res // return only result
// func(.....) (res, error) // return one result and  error
//
func wrap(function interface{}, args ...interface{}) ProcessFunc {
	rv := reflect.ValueOf(function)
	if rv.Kind() != reflect.Func {
		return nil
	}
	// Can not execute the function when argsNum is different with args len
	if argsNum := rv.Type().NumIn(); argsNum != len(args) {
		return nil
	}

	outNum := rv.Type().NumOut()
	resultIndex := -1
	errorIndex := -1
	switch outNum {
	case 0:
	case 1:
		outName := rv.Type().Out(0).Name()
		if outName == "error" {
			errorIndex = 0
		} else {
			resultIndex = 0
		}
	case 2:
		outName := rv.Type().Out(0).Name()
		if outName == "error" {
			resultIndex = 1
			errorIndex = 0
		} else if outName = rv.Type().Out(1).Name(); outName == "error" {
			resultIndex = 0
			errorIndex = 1
		}
	default:
		return nil
	}

	rArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		rArgs[i] = reflect.ValueOf(arg)
	}

	return func() (interface{}, error) {
		ret := rv.Call(rArgs)
		if outNum == 0 {
			return nil, nil
		}

		var e interface{}
		if errorIndex >= 0 {
			e = ret[errorIndex].Interface()
		}

		err, ok := e.(error)
		if !ok {
			err = nil
		}

		if resultIndex < 0 {
			return nil, err
		}

		return ret[resultIndex].Interface(), err
	}
}

// NewManager is creating a new manager for workers
func NewManager(workerNum int) *Manager {
	m := &Manager{
		count:         0,
		processNum:    0,
		workerNum:     workerNum,
		forceStop:     false,
		start:         make(chan int),
		stop:          make(chan int),
		In:            make(chan *Process, 1),
		Out:           make(chan *Process, 1),
		FailFilter:    []FilterFunc{},
		SuccessFilter: []FilterFunc{},
		NotExec:       []string{},
	}

	for i := 0; i < workerNum; i++ {
		go m.startWorker()
	}

	return m
}

// start worker
func (m *Manager) startWorker() {
	for {
		select {
		case <-m.stop:
			break
		case p := <-m.In:
			p.Result, p.Error = p.Function()
			m.Out <- p
		}
	}
}

// Stop forces workers to stop their process
func (m *Manager) stopProcesses() {
	num := m.workerNum + m.processNum
	for i := 0; i < num; i++ {
		m.stop <- 1
	}
}

// Add is adding a new worker process into worker queue
func (m *Manager) Add(id string, function interface{}, args ...interface{}) *Manager {
	f := wrap(function, args...)
	if f == nil {
		m.NotExec = append(m.NotExec, id)
		return m
	}

	p := &Process{
		ID:       id,
		Function: f,
		Result:   nil,
		Error:    nil,
	}

	m.count++
	m.processNum++
	go func() {
		<-m.start
		for {
			select {
			case <-m.stop:
				return
			case m.In <- p:
				return
			}
		}
	}()

	return m
}

// Fail adds a fail filter
func (m *Manager) Fail(f ...FilterFunc) *Manager {
	m.FailFilter = append(m.FailFilter, f...)
	return m
}

// Success adds a success filter
func (m *Manager) Success(f ...FilterFunc) *Manager {
	m.SuccessFilter = append(m.SuccessFilter, f...)
	return m
}

// Run retrieves result by worker
func (m *Manager) Run(result interface{}) []*Process {
	processes := []*Process{}
	if m.count == 0 {
		return processes
	}

	setFlg := false
	rv := reflect.ValueOf(result)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.IsValid() {
		setFlg = true
	}

	// give processes workers
	for i := 0; i < m.count; i++ {
		m.start <- 1
	}

	for {
		if m.count <= 0 {
			break
		}

		if m.forceStop {
			break
		}

		select {
		case p := <-m.Out:
			if p.Error != nil {
				// execute fail filter
				for _, f := range m.FailFilter {
					f(p)
				}
			} else if p.Result != nil {
				// execute success filter
				for _, f := range m.SuccessFilter {
					f(p)
				}
			}

			processes = append(processes, p)
			m.count--

			if !setFlg {
				continue
			}

			field := rv.FieldByName(p.ID)
			if !field.CanSet() {
				continue
			}
			fieldType := field.Type()

			resultValue := reflect.ValueOf(p.Result)
			resultType := resultValue.Type()

			if fieldType != resultType {
				continue
			}
			field.Set(resultValue)
		}
	}

	go m.stopProcesses()
	return processes
}

// Stop forces workers to stop
func (m *Manager) Stop() {
	m.forceStop = true
}

// GetNotExecute returns ids represents process
func (m *Manager) GetNotExecute() []string {
	return m.NotExec
}
