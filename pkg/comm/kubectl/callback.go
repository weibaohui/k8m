package kubectl

import (
	"context"
	"fmt"
	"sort"

	"k8s.io/klog/v2"
)

func initializeCallbacks(kubectl *Kubectl) *callbacks {
	return &callbacks{
		processors: map[string]*processor{
			"get":    {kubectl: kubectl},
			"patch":  {kubectl: kubectl},
			"create": {kubectl: kubectl},
			"update": {kubectl: kubectl},
			"delete": {kubectl: kubectl},
			"list":   {kubectl: kubectl},
		},
	}
}

// callbacks gorm callbacks manager
type callbacks struct {
	processors map[string]*processor
}

type processor struct {
	kubectl   *Kubectl
	fns       []func(context.Context, *Kubectl) error
	callbacks []*callback
}
type callback struct {
	name      string
	before    string
	after     string
	remove    bool
	replace   bool
	handler   func(context.Context, *Kubectl) error
	processor *processor
}

func (cs *callbacks) Create() *processor {
	return cs.processors["create"]
}
func (cs *callbacks) Patch() *processor {
	return cs.processors["patch"]
}
func (cs *callbacks) Get() *processor {
	return cs.processors["get"]
}

func (cs *callbacks) Update() *processor {
	return cs.processors["update"]
}

func (cs *callbacks) Delete() *processor {
	return cs.processors["delete"]
}
func (cs *callbacks) List() *processor {
	return cs.processors["list"]
}

func (c *callback) Remove(name string) error {
	klog.V(4).Infof("removing callback `%s` \n", name)
	c.name = name
	c.remove = true
	c.processor.callbacks = append(c.processor.callbacks, c)
	return c.processor.compile()
}

func (c *callback) Replace(name string, fn func(context.Context, *Kubectl) error) error {
	klog.V(4).Infof("replacing callback `%s` \n", name)
	c.name = name
	c.handler = fn
	c.replace = true
	c.processor.callbacks = append(c.processor.callbacks, c)
	return c.processor.compile()
}

func (c *callback) Before(name string) *callback {
	c.before = name
	return c
}

func (c *callback) After(name string) *callback {
	c.after = name
	return c
}

func (c *callback) Register(name string, fn func(context.Context, *Kubectl) error) error {
	c.name = name
	c.handler = fn
	c.processor.callbacks = append(c.processor.callbacks, c)
	return c.processor.compile()
}

func (p *processor) Get(name string) func(context.Context, *Kubectl) error {
	for i := len(p.callbacks) - 1; i >= 0; i-- {
		if v := p.callbacks[i]; v.name == name && !v.remove {
			return v.handler
		}
	}
	return nil
}

func (p *processor) Remove(name string) error {
	return (&callback{processor: p}).Remove(name)
}

func (p *processor) Replace(name string, fn func(context.Context, *Kubectl) error) error {
	return (&callback{processor: p}).Replace(name, fn)
}

func (p *processor) Execute(ctx context.Context, k8s *Kubectl) error {
	for _, f := range p.fns {
		err := f(ctx, k8s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *processor) Before(name string) *callback {
	return &callback{before: name, processor: p}
}

func (p *processor) After(name string) *callback {
	return &callback{after: name, processor: p}
}

func (p *processor) Register(name string, fn func(context.Context, *Kubectl) error) error {
	return (&callback{processor: p}).Register(name, fn)
}

func (p *processor) compile() (err error) {
	var callbacks []*callback
	removedMap := map[string]bool{}
	for _, cb := range p.callbacks {
		callbacks = append(callbacks, cb)
		if cb.remove {
			removedMap[cb.name] = true
		}
	}

	if len(removedMap) > 0 {
		callbacks = removeCallbacks(callbacks, removedMap)
	}
	p.callbacks = callbacks

	if p.fns, err = sortCallbacks(p.callbacks); err != nil {
		klog.V(4).Infof("Got error when compile callbacks, got %v", err)
	}
	return
}
func sortCallbacks(cs []*callback) (fns []func(context.Context, *Kubectl) error, err error) {
	var (
		names, sorted []string
		sortCallback  func(*callback) error
	)
	sort.SliceStable(cs, func(i, j int) bool {
		if cs[j].before == "*" && cs[i].before != "*" {
			return true
		}
		if cs[j].after == "*" && cs[i].after != "*" {
			return true
		}
		return false
	})

	for _, c := range cs {
		// show warning message the callback name already exists
		if idx := getRIndex(names, c.name); idx > -1 && !c.replace && !c.remove && !cs[idx].remove {
			klog.V(4).Infof("duplicated callback `%s` \n", c.name)
		}
		names = append(names, c.name)
	}

	sortCallback = func(c *callback) error {
		if c.before != "" { // if defined before callback
			if c.before == "*" && len(sorted) > 0 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					sorted = append([]string{c.name}, sorted...)
				}
			} else if sortedIdx := getRIndex(sorted, c.before); sortedIdx != -1 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					// if before callback already sorted, append current callback just after it
					sorted = append(sorted[:sortedIdx], append([]string{c.name}, sorted[sortedIdx:]...)...)
				} else if curIdx > sortedIdx {
					return fmt.Errorf("conflicting callback %s with before %s", c.name, c.before)
				}
			} else if idx := getRIndex(names, c.before); idx != -1 {
				// if before callback exists
				cs[idx].after = c.name
			}
		}

		if c.after != "" { // if defined after callback
			if c.after == "*" && len(sorted) > 0 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					sorted = append(sorted, c.name)
				}
			} else if sortedIdx := getRIndex(sorted, c.after); sortedIdx != -1 {
				if curIdx := getRIndex(sorted, c.name); curIdx == -1 {
					// if after callback sorted, append current callback to last
					sorted = append(sorted, c.name)
				} else if curIdx < sortedIdx {
					return fmt.Errorf("conflicting callback %s with before %s", c.name, c.after)
				}
			} else if idx := getRIndex(names, c.after); idx != -1 {
				// if after callback exists but haven't sorted
				// set after callback's before callback to current callback
				after := cs[idx]

				if after.before == "" {
					after.before = c.name
				}

				if err := sortCallback(after); err != nil {
					return err
				}

				if err := sortCallback(c); err != nil {
					return err
				}
			}
		}

		// if current callback haven't been sorted, append it to last
		if getRIndex(sorted, c.name) == -1 {
			sorted = append(sorted, c.name)
		}

		return nil
	}

	for _, c := range cs {
		if err = sortCallback(c); err != nil {
			return
		}
	}

	for _, name := range sorted {
		if idx := getRIndex(names, name); !cs[idx].remove {
			fns = append(fns, cs[idx].handler)
		}
	}

	return
}

func removeCallbacks(cs []*callback, nameMap map[string]bool) []*callback {
	callbacks := make([]*callback, 0, len(cs))
	for _, callback := range cs {
		if nameMap[callback.name] {
			continue
		}
		callbacks = append(callbacks, callback)
	}
	return callbacks
}

// getRIndex get right index from string slice
func getRIndex(strs []string, str string) int {
	for i := len(strs) - 1; i >= 0; i-- {
		if strs[i] == str {
			return i
		}
	}
	return -1
}
