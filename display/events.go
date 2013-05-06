package display

import (
	"log"

	"github.com/bluepeppers/allegro"
)

func (d *DisplayEngine) eventHandler() {
	es := []*allegro.EventSource{d.display.GetEventSource(),
		allegro.GetKeyboardEventSource(),
		allegro.GetMouseEventSource()}
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
		case allegro.KeyCharEvent:
			d.handleKeyChar(tev)
		case allegro.MouseButtonDown:
			d.handleMouseDown(tev)
		}
		d.statusLock.RLock()
		stopped = !d.running
		d.statusLock.RUnlock()
	}
}

func (d *DisplayEngine) handleKeyChar(ev allegro.KeyCharEvent) {
	var x, y int
	switch ev.Keycode {
	case allegro.KEY_LEFT:
		x = -SCROLL_SPEED
	case allegro.KEY_RIGHT:
		x = SCROLL_SPEED
	case allegro.KEY_UP:
		y = -SCROLL_SPEED
	case allegro.KEY_DOWN:
		y = SCROLL_SPEED
	}
	d.drawLock.Lock()
	d.viewport.X += x
	d.viewport.Y += y
	d.drawLock.Unlock()
}

func (d *DisplayEngine) handleResize(ev allegro.DisplayResizeEvent) {
	d.drawLock.Lock()
	d.viewport.W = ev.W
	d.viewport.H = ev.H
	log.Printf("Acknowledging resize to %v, %v", ev.W, ev.H)
	d.display.AcknowledgeResize()
	d.drawLock.Unlock()
}

func (d *DisplayEngine) handleMouseDown(event allegro.MouseButtonDown) {
	if event.Button == 1 {
		go d.startScrolling(event)
	}
}

func (d *DisplayEngine) startScrolling(start allegro.MouseButtonDown) {
	timer := allegro.CreateTimer(float64(1) / 20)
	es := []*allegro.EventSource{allegro.GetMouseEventSource(),
		timer.GetEventSource()}
	timer.Start()
	defer timer.Destroy()

	x, y := start.X, start.Y
	running := true
	for ev := range allegro.GetEvents(es) {
		switch tev := ev.(type) {
		case allegro.MouseButtonUp:
			if tev.Button == start.Button {
				running = false
			}
		case allegro.TimerEvent:
			d.drawLock.Lock()
			d.viewport.X += (x - start.X) / 20
			d.viewport.Y += (y - start.Y) / 20
			d.drawLock.Unlock()
		case allegro.MouseAxesEvent:
			x, y = tev.X, tev.Y
		}
		d.statusLock.RLock()
		running = d.running && running
		d.statusLock.RUnlock()
		if !running {
			break
		}
	}
}
