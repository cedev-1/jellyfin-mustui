package player

import (
	"net/http"
	"sync"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type State int

const (
	StateStopped State = iota
	StatePlaying
	StatePaused
)

type Track struct {
	ID       string
	Name     string
	Artist   string
	Album    string
	Duration time.Duration
	URL      string
}

type Player struct {
	mu           sync.Mutex
	state        State
	currentTrack *Track
	queue        []Track
	queueIndex   int

	streamer beep.StreamSeekCloser
	ctrl     *beep.Ctrl
	format   beep.Format
	done     chan bool

	position time.Duration
	volume   float64

	OnStateChange func(State)
	OnTrackChange func(*Track)
	OnProgress    func(time.Duration, time.Duration)
}

func New() *Player {
	return &Player{
		state:      StateStopped,
		queue:      make([]Track, 0),
		queueIndex: -1,
		done:       make(chan bool),
		volume:     1.0,
	}
}

func (p *Player) Init() error {
	return speaker.Init(44100, 44100/10)
}

func (p *Player) LoadTrack(track Track) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer != nil {
		speaker.Clear()
		p.streamer.Close()
	}
	resp, err := http.Get(track.URL)
	if err != nil {
		return err
	}

	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		resp.Body.Close()
		return err
	}

	p.streamer = streamer
	p.format = format
	p.currentTrack = &track
	p.state = StateStopped
	p.position = 0

	if p.OnTrackChange != nil {
		p.OnTrackChange(&track)
	}

	return nil
}

func (p *Player) Play() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.streamer == nil {
		return
	}

	if p.state == StatePlaying {
		return
	}

	p.ctrl = &beep.Ctrl{Streamer: p.streamer, Paused: false}
	p.state = StatePlaying

	speaker.Play(beep.Seq(p.ctrl, beep.Callback(func() {
		p.done <- true
	})))

	if p.OnStateChange != nil {
		p.OnStateChange(p.state)
	}

	go p.trackProgress()
}

func (p *Player) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl == nil || p.state != StatePlaying {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = true
	speaker.Unlock()

	p.state = StatePaused

	if p.OnStateChange != nil {
		p.OnStateChange(p.state)
	}
}

func (p *Player) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctrl == nil || p.state != StatePaused {
		return
	}

	speaker.Lock()
	p.ctrl.Paused = false
	speaker.Unlock()

	p.state = StatePlaying

	if p.OnStateChange != nil {
		p.OnStateChange(p.state)
	}
}

func (p *Player) TogglePause() {
	if p.state == StatePlaying {
		p.Pause()
	} else if p.state == StatePaused {
		p.Resume()
	} else if p.currentTrack != nil {
		p.Play()
	}
}

func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	speaker.Clear()
	if p.streamer != nil {
		p.streamer.Close()
		p.streamer = nil
	}
	p.ctrl = nil
	p.state = StateStopped
	p.position = 0

	if p.OnStateChange != nil {
		p.OnStateChange(p.state)
	}
}

func (p *Player) SetQueue(tracks []Track) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.queue = tracks
	p.queueIndex = -1
}

func (p *Player) PlayFromQueue(index int) error {
	p.mu.Lock()
	if index < 0 || index >= len(p.queue) {
		p.mu.Unlock()
		return nil
	}
	p.queueIndex = index
	track := p.queue[index]
	p.mu.Unlock()

	if err := p.LoadTrack(track); err != nil {
		return err
	}
	p.Play()
	return nil
}

func (p *Player) Next() error {
	p.mu.Lock()
	nextIndex := p.queueIndex + 1
	if nextIndex >= len(p.queue) {
		nextIndex = 0
	}
	p.mu.Unlock()
	return p.PlayFromQueue(nextIndex)
}

func (p *Player) Previous() error {
	p.mu.Lock()
	prevIndex := p.queueIndex - 1
	if prevIndex < 0 {
		prevIndex = len(p.queue) - 1
	}
	p.mu.Unlock()
	return p.PlayFromQueue(prevIndex)
}

func (p *Player) GetState() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

func (p *Player) GetCurrentTrack() *Track {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentTrack
}

func (p *Player) GetPosition() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil {
		return 0
	}
	return p.format.SampleRate.D(p.streamer.Position())
}

func (p *Player) GetDuration() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.streamer == nil {
		return 0
	}
	return p.format.SampleRate.D(p.streamer.Len())
}

func (p *Player) trackProgress() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if p.GetState() == StatePlaying && p.OnProgress != nil {
				p.OnProgress(p.GetPosition(), p.GetDuration())
			}
		case <-p.done:
			p.Next()
			return
		}
	}
}

func (p *Player) Close() {
	p.Stop()
	speaker.Close()
}
