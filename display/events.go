package display

import (
	"log"

	"github.com/bluepeppers/allegro"
)

func (d *DisplayEngine) eventHandler() {
	src := d.Display.GetEventSource()
	defer src.StopGetEvents()
	es := []*allegro.EventSource{src}
	queue := allegro.GetEvents(es)
	stopped := false
	for !stopped {
		ev := <-queue
		switch tev := ev.(type) {
		case allegro.DisplayCloseEvent:
			d.statusLock.Lock()
			d.running = false
			d.statusLock.Unlock()
		case allegro.DisplayResizeEvent:
			d.handleResize(tev)
		}
		d.statusLock.RLock()
		stopped = !d.running
		d.statusLock.RUnlock()
	}
}


func (d *DisplayEngine) handleResize(ev allegro.DisplayResizeEvent) {
	d.drawLock.Lock()
	d.viewport.W = ev.W
	d.viewport.H = ev.H
	log.Printf("Acknowledging resize to %v, %v", ev.W, ev.H)
	d.Display.AcknowledgeResize()
	d.drawLock.Unlock()
}

