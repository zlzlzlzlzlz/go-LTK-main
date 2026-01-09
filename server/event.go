package server

// 事件类型
type eventType uint8

const (
// sendCardEvent eventType = iota //发卡事件
)

type eventI interface {
	trigger(*Games)          //触发事件
	isSkipAble() bool        //事件是否跳过
	setSkip(bool)            //设置是否跳过事件
	getEventType() eventType //获取事件类型
}

type event struct {
	skip      bool
	eventType eventType
}

func (e *event) trigger(g *Games) {}

func (e *event) isSkipAble() bool {
	return e.skip
}

func (e *event) setSkip(skip bool) {
	e.skip = skip
}

func (e *event) getEventType() eventType {
	return e.eventType
}

// 事件列表
type eventList struct {
	list []eventI
}

// 返回事件列表大小
func (l *eventList) size() int {
	return len(l.list)
}

// 在列表尾部添加事件
func (l *eventList) append(e ...eventI) {
	l.list = append(l.list, e...)
}

// 在index为i与i+1的事件之间插入事件
func (l *eventList) insert(i int, e ...eventI) {
	l.list = append(l.list[:i+1], append(e, l.list[i+1:]...)...)
}
