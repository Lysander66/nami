package nami

type ReadyNotifier interface {
	WorkerReady(chan Task)
}

type Scheduler interface {
	ReadyNotifier
	Run()
	Submit(Task)
	Worker() chan Task
}

type TaskScheduler struct {
	taskQueue  chan Task
	workerPool chan chan Task
}

func NewTaskScheduler() Scheduler {
	scheduler := TaskScheduler{
		taskQueue:  make(chan Task),
		workerPool: make(chan chan Task),
	}
	return &scheduler
}

func (s TaskScheduler) Run() {
	go func() {
		var tasks []Task
		var workers []chan Task
		for {
			var activeTask Task
			var activeWorker chan Task
			if len(tasks) > 0 && len(workers) > 0 {
				activeTask = tasks[0]
				activeWorker = workers[0]
			}
			select {
			case t := <-s.taskQueue:
				tasks = append(tasks, t)
			case w := <-s.workerPool:
				workers = append(workers, w)
			case activeWorker <- activeTask:
				tasks = tasks[1:]
				workers = workers[1:]
			}
		}
	}()
}

func (s TaskScheduler) Submit(task Task) {
	s.taskQueue <- task
}

func (s TaskScheduler) Worker() chan Task {
	return make(chan Task)
}

func (s TaskScheduler) WorkerReady(w chan Task) {
	s.workerPool <- w
}

// -------------------------------- SimpleScheduler -------------------------------------
type SimpleScheduler struct {
	worker chan Task
}

func (s *SimpleScheduler) Run() {
	s.worker = make(chan Task)
}

func (s *SimpleScheduler) Submit(task Task) {
	go func() {
		s.worker <- task
	}()
}

func (s *SimpleScheduler) Worker() chan Task {
	return s.worker
}

func (s *SimpleScheduler) WorkerReady(chan Task) {
	//panic("implement me")
}
