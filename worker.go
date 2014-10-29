package worker

import "reflect"

type (
	// Manager is manager for a workers
	Manager struct {
		count         int
		workerNum     int
		forceStop     bool
		start         chan int
		In            chan *Process
		Out           chan *Process
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

		var e interface{}
		for i, v := range ret {
			if i == errorIndex {
				e = ret[i].Interface()
				continue
			}
			result = append(result, v.Interface())
		}

		err, ok := e.(error)
		if !ok {
			err = nil
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
	for p := range m.In {
		if m.forceStop {
			return
		}

		p.Result, p.Error = p.Function()
		m.Out <- p
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
			case m.In <- p:
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
		case p := <-m.Out:
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

			result[p.index] = p
			m.count--
		}

		if m.count <= 0 {
			break
		}
	}

	return result
}

// GetNotExecute returns ids represents process
func (m *Manager) GetNotExecute() []string {
	return m.NotExec
}
