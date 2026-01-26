package videowf

type TaskStatus string

const (
	TaskStatusNil     TaskStatus = ""
	TaskStatusPending TaskStatus = "pending"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusFailed  TaskStatus = "failed"
)

func (ts TaskStatus) IsValid() bool {
	switch ts {
	case TaskStatusNil, TaskStatusPending, TaskStatusDone, TaskStatusFailed:
		return true
	}
	return false
}

type Task struct {
	Status TaskStatus `json:"status,omitempty"`
	Error  string     `json:"error,omitempty"`
}

func (t *Task) IsZero() bool {
	return t.Status == TaskStatusNil && t.Error == ""
}

func (t *Task) MarkPending() {
	t.Status = TaskStatusPending
	t.Error = ""
}

func (t *Task) MarkDone() {
	t.Status = TaskStatusDone
	t.Error = ""
}

func (t *Task) MarkFailed(err error) {
	t.Status = TaskStatusFailed
	if err != nil {
		t.Error = err.Error()
	} else {
		t.Error = ""
	}
}

func (t *Task) HasBeenStarted() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusDone || t.Status == TaskStatusFailed
}
