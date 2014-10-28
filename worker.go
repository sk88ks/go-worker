package worker

import "reflect"

type (
	// Manager is manager for a workers
	Manager struct {
		count         int
		workerNum     int
		forceStop     bool
		Processes     chan *Process
		Result        chan *Process
		FailFilter    []FilterFunc
		SuccessFilter []FilterFunc
		NotExec       []string
	}

	// Process has an implementation
	// for worker and information about process
	Process struct {
		index    int
		ID       string
		Function ProcessFunc
		Result   []interface{}
		Err      error
	}

	// ProcessFunc is function executed by individual worker
	ProcessFunc func() ([]interface{}, error)

	// FilterFunc is function executed at middle of process as filter
	FilterFunc func(*Process)
)

// Wrap func to be used as WorkerFunc
func wrap(function interface{}, args ...interface{}) ProcessFunc {
	rv := reflect.ValueOf(function)
	if rv.Kind() != reflect.Func {
		return nil
	}
	// Can not execute the function when argsNum is different with args len
	if argsNum := rv.Type().NumIn(); argsNum != len(args) {
		return nil
	}
	rArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		rArgs[i] = reflect.ValueOf(arg)
	}

	return func() ([]interface{}, error) {
		ret := rv.Call(rArgs)
		result := []interface{}{}
		var err error
		for i := len(ret) - 1; i >= 0; i-- {
			if e, ok := ret[i].Interface().(error); ok {
				err = e
				break
			}
			result = append(result, ret[i].Interface())
		}

		if err != nil {
			return result, err
		}

		return result, nil
	}
}

// NewManager is creating a new manager for workers
func NewManager(workerNum int) *Manager {
	m := &Manager{
		count:         0,
		workerNum:     workerNum,
		forceStop:     false,
		Processes:     make(chan *Process, 1),
		FailFilter:    []FilterFunc{},
		SuccessFilter: []FilterFunc{},
		NotExec:       []string{},
	}

	for i := 0; i < workerNum; i++ {
		go m.startWorker()
	}

	return m
}

func filter() {

}

// start worker
func (m *Manager) startWorker() {
	for p := range m.Processes {
		if m.forceStop {
			return
		}

		p.Result, p.Err = p.Function()
		if len(p.Result) != 0 {
			// execute fail filter
			for _, f := range m.FailFilter {
				f(p)
			}
		} else if len(p.Result) != 0 {
			// execute success filter
			for _, f := range m.SuccessFilter {
				f(p)
			}
		}

		m.Result <- p
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
		index:    m.count,
		ID:       id,
		Function: f,
		Result:   nil,
		Err:      nil,
	}

	go func() {
		for {
			select {
			case m.Processes <- p:
				m.count++
			}
		}
	}()

	return m
}

// Fail adds a fail filter
func (m *Manager) Fail(f ...FilterFunc) {
	m.FailFilter = append(m.FailFilter, f...)
}

// Success adds a success filter
func (m *Manager) Success(f ...FilterFunc) {
	m.SuccessFilter = append(m.SuccessFilter, f...)
}

// End retrieves result by worker
func (m *Manager) End() []*Process {
	result := make([]*Process, m.count)
	for {
		select {
		case res := <-m.Result:
			result[res.index] = res
			m.count--
		}

		if m.count <= 0 {
			break
		}
	}

	return result
}

// EndWithFailStop retrieves results by worker
// Force worker to stop a process when err occured
func (m *Manager) EndWithFailStop() ([]*Process, error) {
	result := make([]*Process, m.count)
	for {
		select {
		case res := <-m.Result:
			if res.Err != nil {
				m.forceStop = true
				return result, res.Err
			}
			result[res.index] = res
			m.count--
		}

		if m.count <= 0 {
			break
		}
	}

	return result, nil
}
