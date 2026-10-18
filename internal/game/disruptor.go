package game

import "time"

// Names of the effects a disruptor can trigger. Clients use the very same
// strings, see disruptor.js.
const (
	DisruptorEffectShaky       = "shaky"
	DisruptorEffectLag         = "lag"
	DisruptorEffectFlip        = "flip"
	DisruptorEffectRandomSize  = "random-size"
	DisruptorEffectRandomColor = "random-color"
	DisruptorEffectEraser      = "eraser"
	DisruptorEffectUndo        = "undo"
)

// Severity describes how much an effect gets in the way of the drawer. The
// more severe an effect, the longer its cooldown.
const (
	SeverityMild   = 1
	SeverityMedium = 2
	SeveritySevere = 3
)

// DisruptorEffect describes a single effect, that the disruptor can trigger
// by pressing a button.
type DisruptorEffect struct {
	Name     string
	Severity int
	// Duration is how long the drawer is affected. Zero means the effect is
	// instant (like undo) and doesn't need to be removed again.
	Duration time.Duration
	// Cooldown is the time until the disruptor may use this very effect again.
	// It starts when the effect is triggered, so it always includes Duration.
	Cooldown time.Duration
}

// disruptorEffects is the single source of truth for what effects exist and
// how long they have to cool down. Tune the balance here.
var disruptorEffects = []DisruptorEffect{
	// Mild: annoying, but the drawer can still work around it.
	{Name: DisruptorEffectShaky, Severity: SeverityMild, Duration: 6 * time.Second, Cooldown: 15 * time.Second},
	{Name: DisruptorEffectRandomSize, Severity: SeverityMild, Duration: 8 * time.Second, Cooldown: 15 * time.Second},
	{Name: DisruptorEffectRandomColor, Severity: SeverityMild, Duration: 8 * time.Second, Cooldown: 15 * time.Second},
	// Medium: the drawer has to fight their own pen.
	{Name: DisruptorEffectLag, Severity: SeverityMedium, Duration: 6 * time.Second, Cooldown: 25 * time.Second},
	{Name: DisruptorEffectEraser, Severity: SeverityMedium, Duration: 4 * time.Second, Cooldown: 30 * time.Second},
	// Severe: the drawer basically can't draw (flip) or loses progress (undo).
	{Name: DisruptorEffectFlip, Severity: SeveritySevere, Duration: 5 * time.Second, Cooldown: 40 * time.Second},
	{Name: DisruptorEffectUndo, Severity: SeveritySevere, Duration: 0, Cooldown: 45 * time.Second},
}

func findDisruptorEffect(name string) (DisruptorEffect, bool) {
	for _, effect := range disruptorEffects {
		if effect.Name == name {
			return effect, true
		}
	}

	return DisruptorEffect{}, false
}

// handleDisruptorEffect validates and applies an effect requested by the
// disruptor. Cooldowns are enforced here, so a modified client can't skip
// them. Most effects only change how the drawer's pen behaves, which happens
// on the drawer's client. For those, the server just announces the effect to
// everybody. Undo is different, as the drawing lives on the server.
func (lobby *Lobby) handleDisruptorEffect(player *Player, effectName string) {
	if !lobby.isActiveDisruptor(player) {
		return
	}

	effect, known := findDisruptorEffect(effectName)
	if !known {
		return
	}

	now := time.Now()
	if readyAt, used := player.disruptorEffectReadyAt[effect.Name]; used && now.Before(readyAt) {
		return
	}

	if effect.Name == DisruptorEffectUndo && !lobby.undoLastStroke() {
		// Nothing to undo. We don't burn the cooldown for that.
		return
	}

	if player.disruptorEffectReadyAt == nil {
		player.disruptorEffectReadyAt = make(map[string]time.Time, len(disruptorEffects))
	}
	player.disruptorEffectReadyAt[effect.Name] = now.Add(effect.Cooldown)

	lobby.Broadcast(&Event{
		Type: EventTypeDisruptorEffect,
		Data: &DisruptorEffectEvent{
			Effect:     effect.Name,
			PlayerID:   player.ID,
			PlayerName: player.Name,
			DurationMs: int(effect.Duration.Milliseconds()),
			CooldownMs: int(effect.Cooldown.Milliseconds()),
		},
	})

	awardDisruptorAction(lobby, player)
}

// undoLastStroke removes the last connected stroke (or fill) from the drawing
// and tells everybody about the new drawing. It returns false if there was
// nothing to undo.
func (lobby *Lobby) undoLastStroke() bool {
	if len(lobby.currentDrawing) == 0 || len(lobby.connectedDrawEventsIndexStack) == 0 {
		return false
	}

	lastIndex := len(lobby.connectedDrawEventsIndexStack) - 1
	undoFrom := lobby.connectedDrawEventsIndexStack[lastIndex]
	lobby.connectedDrawEventsIndexStack = lobby.connectedDrawEventsIndexStack[:lastIndex]
	if undoFrom >= len(lobby.currentDrawing) {
		return false
	}

	lobby.currentDrawing = lobby.currentDrawing[:undoFrom]
	lobby.Broadcast(&Event{Type: EventTypeDrawing, Data: lobby.currentDrawing})
	return true
}
