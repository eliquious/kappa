package utils

import (
    "os"

    log "github.com/mgutz/logxi/v1"
)

// Task represents a process running in a separate goroutine.
type Task interface {

    // Run starts the process and waits until the a message is sent over the channel to close.
    Run(logger log.Logger, closer <-chan bool)

    // Wait will wait until the process has exited after receiving a close message.
    Wait()
}

// TaskRunner executes tasks and stops then when a signal is sent
type TaskRunner struct {
    tasks []Task
}

// AddTasks adds tasks to the TaskRunner.
func (t *TaskRunner) AddTasks(tasks ...Task) {
    t.tasks = append(t.tasks, tasks...)
}

// Run starts all the tasks and wait until a signal is given.
func (t *TaskRunner) Run(logger log.Logger, joiner <-chan os.Signal) {
    closer := make(chan bool, len(t.tasks))

    // Start tasks
    for _, task := range t.tasks {
        task.Run(logger, closer)
    }

    // Joins the task runner until a signal is received
    <-joiner

    // send quit message to tasks and wait for exiting
    for _, task := range t.tasks {
        closer <- true
        task.Wait()
    }
}
