package display

import (
	"log"

	"github.com/bluepeppers/allegro"
)

func (d *DisplayEngine) eventHandler() {
	src := d.Display.GetEventSource()
	defer src.StopGetEvents()
	es := []*allegro.EventSource{src, allegro.GetMouseEventSource()}
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
		case allegro.MouseButtonDown:
			tx, ty := d.viewport.ScreenCoordinatesToTile(tev.X, tev.Y, d.config)
			log.Printf("S: (%v, %v) T: (%v, %v)", tev.X, tev.Y, tx, ty)
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
	d.Display.AcknowledgeResize()
	d.drawLock.Unlock()
}
