package trafficlight

import (
	"sync/atomic"
	"time"
)

// Important : Le feu doit etre démarré (Start) avant d'appeler
// toute autre méthode.

// State représente l'état d'un feu de circulation.
type State int

const (
	Red State = iota
	Green
)

func (s State) String() string {
	switch s {
	case Red:
		return "Red"
	case Green:
		return "Green"
	default:
		return "Unknown"
	}
}

// ZoneActivator est une abstraction pour les zones de force contrôlées par
// le feu. Le package trafficlight appelle uniquement Activate/Deactivate.
type ZoneActivator interface {
	Activate()
	Deactivate()
}

// trafficCmd est le type commun pour tous les messages/commandes envoyés
// à l'acteur (la goroutine propriétaire de l'état). Les valeurs concrètes
// envoyées sur `cmdCh` sont des structs ci-dessous.
type trafficCmd interface{}

// cmdSubscribe : demande un abonnement. Le champ `resp` est un canal par
// lequel l'acteur renverra au demandeur un `chan State` sur lequel il
// recevra ensuite les changements d'état.
type cmdSubscribe struct{ resp chan chan State }

// cmdUnsubscribe : demande la suppression d'un abonnement ; `ch` est le
// canal précédemment renvoyé par l'acteur.
type cmdUnsubscribe struct{ ch chan State }

// cmdAddZone : ajoute une zone (`ZoneActivator`) au contrôle du feu.
type cmdAddZone struct{ z ZoneActivator }

// cmdRemoveZone : supprime une zone précédemment ajoutée.
type cmdRemoveZone struct{ z ZoneActivator }

// cmdSetState : demande de forcer l'état du feu (`Red` ou `Green`).
type cmdSetState struct{ s State }

// cmdGetState : requête de lecture synchrone de l'état ; l'acteur
// écrira l'état courant sur `resp`.
type cmdGetState struct{ resp chan State }

// cmdStop : demande un arrêt propre de l'acteur ; l'acteur répondra
// sur `resp` quand l'arrêt sera terminé.
type cmdStop struct{ resp chan struct{} }

// TrafficLight acteur. L'état interne (state, listes de
// zones, abonnés, timer) est détenu par la goroutine actrice.
type TrafficLight struct {
	id            string
	greenDuration time.Duration
	redDuration   time.Duration

	// canal sur lequel les autres goroutines envoient des commandes
	cmdCh chan trafficCmd

	// flag atomique pour rendre Start/Stop idempotents et éviter les
	// démarrages concurrents (utilise `sync/atomic`).
	started int32
}

// NewTrafficLight crée le trafic light (ne démarre pas l'acteur).
func NewTrafficLight(id string, green, red time.Duration) *TrafficLight {
	if green <= 0 {
		green = 10 * time.Second
	}
	if red <= 0 {
		red = 10 * time.Second
	}
	return &TrafficLight{
		id:            id,
		greenDuration: green,
		redDuration:   red,
	}
}

// Start lance la goroutine-acteur. Appels multiples sont ignorés.
func (t *TrafficLight) Start() {

	if !atomic.CompareAndSwapInt32(&t.started, 0, 1) {
		return
	}
	t.cmdCh = make(chan trafficCmd)

	go func() {
		state := Red
		var timer *time.Timer
		zones := make([]ZoneActivator, 0)
		subs := make(map[chan State]struct{})

		resetTimer := func() {
			if timer != nil {
				timer.Stop()
			}
			if state == Green {
				timer = time.NewTimer(t.greenDuration)
			} else {
				timer = time.NewTimer(t.redDuration)
			}
		}

		notify := func(s State) {
			for ch := range subs {
				select {
				case ch <- s:
				default:
				}
			}
		}

		// apply initial state to zones (none initially)
		resetTimer()

		for {
			select {
			case cmd := <-t.cmdCh:
				switch c := cmd.(type) {
				case cmdSubscribe:
					ch := make(chan State, 1)
					subs[ch] = struct{}{}
					ch <- state
					c.resp <- ch
				case cmdUnsubscribe:
					if _, ok := subs[c.ch]; ok {
						delete(subs, c.ch)
						close(c.ch)
					}
				case cmdAddZone:
					zones = append(zones, c.z)
					if state == Red {
						c.z.Activate()
					} else {
						c.z.Deactivate()
					}
				case cmdRemoveZone:
					for i, z := range zones {
						if z == c.z {
							zones = append(zones[:i], zones[i+1:]...)
							break
						}
					}
				case cmdSetState:
					if state != c.s {
						state = c.s
						if state == Red {
							for _, z := range zones {
								z.Activate()
							}
						} else {
							for _, z := range zones {
								z.Deactivate()
							}
						}
						notify(state)
						resetTimer()
					}
				case cmdGetState:
					c.resp <- state
				case cmdStop:
					// cleanup
					for ch := range subs {
						close(ch)
						delete(subs, ch)
					}
					c.resp <- struct{}{}
					return
				}

			case <-timer.C:
				if state == Green {
					state = Red
					for _, z := range zones {
						z.Activate()
					}
				} else {
					state = Green
					for _, z := range zones {
						z.Deactivate()
					}
				}
				notify(state)
				resetTimer()
			}
		}
	}()
}

// Stop arrête l'acteur proprement.
func (t *TrafficLight) Stop() {
	if atomic.LoadInt32(&t.started) == 0 {
		return
	}
	resp := make(chan struct{})
	t.cmdCh <- cmdStop{resp: resp}
	<-resp
	atomic.StoreInt32(&t.started, 0)
	close(t.cmdCh)
}

// State retourne l'état courant du feu.
func (t *TrafficLight) State() State {
	if atomic.LoadInt32(&t.started) == 0 {
		return Red
	}
	resp := make(chan State)
	t.cmdCh <- cmdGetState{resp: resp}
	return <-resp
}

// SetState force un état via l'acteur.
func (t *TrafficLight) SetState(s State) {
	if atomic.LoadInt32(&t.started) == 0 {
		return
	}
	t.cmdCh <- cmdSetState{s: s}
}

// AddZone ajoute une zone au contrôle du feu.
func (t *TrafficLight) AddZone(z ZoneActivator) {
	if atomic.LoadInt32(&t.started) == 0 {
		return
	}
	t.cmdCh <- cmdAddZone{z: z}
}

// RemoveZone retire une zone.
func (t *TrafficLight) RemoveZone(z ZoneActivator) {
	if atomic.LoadInt32(&t.started) == 0 {
		return
	}
	t.cmdCh <- cmdRemoveZone{z: z}
}

// Subscribe s'abonne aux changements d'état (canal bufferisé 1). Retourne nil si pas démarré.
func (t *TrafficLight) Subscribe() chan State {
	if atomic.LoadInt32(&t.started) == 0 {
		return nil
	}
	resp := make(chan chan State)
	t.cmdCh <- cmdSubscribe{resp: resp}
	return <-resp
}

// Unsubscribe arrête l'abonnement et ferme le canal.
func (t *TrafficLight) Unsubscribe(ch chan State) {
	if atomic.LoadInt32(&t.started) == 0 || ch == nil {
		return
	}
	t.cmdCh <- cmdUnsubscribe{ch: ch}
}
