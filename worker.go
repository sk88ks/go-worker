package worker

import "reflect"

type (
	// Manager is manager for a workers
	Manager struct {
		count         int
		workerNum     int
		forceStop     bool
		start         chan int
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
		Error    error
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

	errorIndex := -1
	if outNum := rv.Type().NumOut(); outNum != 0 {
		for i := 0; i < outNum; i++ {
			outName := rv.Type().Out(i).Name()
			if outName == "error" {
				errorIndex = i
				break
			}
		}
	}

	rArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		rArgs[i] = reflect.ValueOf(arg)
	}

	return func() ([]interface{}, error) {
		ret := rv.Call(rArgs)
		result := []interface{}{}
		err, ok := ret[errorIndex].Interface().(error)
		if !ok {
			err = nil
		}

		for i, v := range ret {
			if i == errorIndex {
				continue
			}
			result = append(result, v.Interface())
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
		start:         make(chan int),
		Processes:     make(chan *Process, 1),
		Result:        make(chan *Process, 1),
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
	for p := range m.Processes {
		if m.forceStop {
			return
		}

		p.Result, p.Error = p.Function()
		if p.Error != nil {
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
func (m *Manager) Add(id string, function interface{}, args ...interface{}) {
	f := wrap(function, args...)
	if f == nil {
		m.NotExec = append(m.NotExec, id)
		return
	}

	p := &Process{
		index:    m.count,
		ID:       id,
		Function: f,
		Result:   nil,
		Error:    nil,
	}

	m.count++
	go func() {
		<-m.start
		for {
			select {
			case m.Processes <- p:
				return
			}
		}
	}()
}

// Start give procces worker
func (m *Manager) Start() {
	for i := 0; i < m.count; i++ {
		m.start <- 1
	}
}

// AddFail adds a fail filter
func (m *Manager) AddFail(f ...FilterFunc) {
	m.FailFilter = append(m.FailFilter, f...)
}

// AddSuccess adds a success filter
func (m *Manager) AddSuccess(f ...FilterFunc) {
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
			if res.Error != nil {
				m.forceStop = true
				return result, res.Error
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

// GetNotExecute returns ids represents process
func (m *Manager) GetNotExecute() []string {
	return m.NotExec
}
