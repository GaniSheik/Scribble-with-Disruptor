package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newDisruptorTestLobby() *Lobby {
	lobby := &Lobby{
		State:            Ongoing,
		CurrentWord:      "a",
		ScoreCalculation: ChillScoring,
	}
	lobby.WriteObject = noOpWriteObject
	lobby.WritePreparedMessage = noOpWritePreparedMessage
	return lobby
}

func Test_assignTurnRoles(t *testing.T) {
	t.Parallel()

	t.Run("three players get exactly one disruptor that is not the drawer", func(t *testing.T) {
		t.Parallel()

		lobby := newDisruptorTestLobby()
		drawer := &Player{Connected: true, State: Drawing}
		lobby.players = []*Player{
			drawer,
			{Connected: true, State: Guessing},
			{Connected: true, State: Guessing},
		}

		disruptor := assignTurnRoles(lobby, drawer)
		require.NotNil(t, disruptor)
		require.NotSame(t, drawer, disruptor)
		require.Equal(t, RoleDrawer, drawer.Role)
		require.Equal(t, RoleDisruptor, disruptor.Role)

		var disruptorCount int
		for _, player := range lobby.players {
			if player.Role == RoleDisruptor {
				disruptorCount++
			}
		}
		require.Equal(t, 1, disruptorCount)
	})

	t.Run("two players get no disruptor, as nobody would be left to guess", func(t *testing.T) {
		t.Parallel()

		lobby := newDisruptorTestLobby()
		drawer := &Player{Connected: true, State: Drawing}
		lobby.players = []*Player{
			drawer,
			{Connected: true, State: Guessing},
		}

		require.Nil(t, assignTurnRoles(lobby, drawer))
		for _, player := range lobby.players {
			require.NotEqual(t, RoleDisruptor, player.Role)
		}
	})

	t.Run("spectators and disconnected players are never picked", func(t *testing.T) {
		t.Parallel()

		lobby := newDisruptorTestLobby()
		drawer := &Player{Connected: true, State: Drawing}
		lobby.players = []*Player{
			drawer,
			{Connected: true, State: Guessing},
			{Connected: true, State: Spectating},
			{Connected: false, State: Guessing},
		}

		require.Nil(t, assignTurnRoles(lobby, drawer))
	})
}

func Test_awardDisruptorAction(t *testing.T) {
	t.Parallel()

	lobby := newDisruptorTestLobby()
	drawer := &Player{Connected: true, State: Drawing, Role: RoleDrawer}
	disruptor := &Player{Connected: true, State: Guessing, Role: RoleDisruptor}
	guesser := &Player{Connected: true, State: Guessing}
	lobby.players = []*Player{drawer, disruptor, guesser}

	// Only the disruptor can earn points.
	awardDisruptorAction(lobby, drawer)
	awardDisruptorAction(lobby, guesser)
	require.Zero(t, drawer.Score)
	require.Zero(t, guesser.Score)

	awardDisruptorAction(lobby, disruptor)
	require.Equal(t, disruptorPointsPerAction, disruptor.Score)

	// An immediate second action is inside the cooldown and earns nothing.
	awardDisruptorAction(lobby, disruptor)
	require.Equal(t, disruptorPointsPerAction, disruptor.Score)

	// Even with the cooldown out of the way, the per turn cap applies.
	for range 100 {
		disruptor.lastDisruptorAward = time.Time{}
		awardDisruptorAction(lobby, disruptor)
	}
	require.Equal(t, disruptorMaxPointsPerTurn, disruptor.Score)
	require.Equal(t, disruptorMaxPointsPerTurn, disruptor.LastScore)

	// Nothing is earned once the turn is over.
	lobby.CurrentWord = ""
	disruptor.lastDisruptorAward = time.Time{}
	disruptor.disruptorPoints = 0
	awardDisruptorAction(lobby, disruptor)
	require.Equal(t, disruptorMaxPointsPerTurn, disruptor.Score)
}

func Test_disruptorDoesNotInflateDrawerScore(t *testing.T) {
	t.Parallel()

	lobby := newDisruptorTestLobby()
	lobby.players = []*Player{
		{Connected: true, State: Drawing, Role: RoleDrawer},
		{Connected: true, State: Standby, LastScore: 100},
		{Connected: true, State: Guessing, Role: RoleDisruptor, LastScore: 50},
	}

	require.Equal(t, 100, ChillScoring.CalculateDrawerScore(lobby))
}

func Test_disruptorDoesNotKeepTheTurnAlive(t *testing.T) {
	t.Parallel()

	lobby := newDisruptorTestLobby()
	lobby.players = []*Player{
		{Connected: true, State: Drawing, Role: RoleDrawer},
		{Connected: true, State: Standby},
		{Connected: true, State: Guessing, Role: RoleDisruptor},
	}
	require.False(t, lobby.isAnyoneStillGuessing())

	lobby.players = append(lobby.players, &Player{Connected: true, State: Guessing})
	require.True(t, lobby.isAnyoneStillGuessing())
}

func Test_disruptorCanNotDraw(t *testing.T) {
	t.Parallel()

	lobby := newDisruptorTestLobby()
	drawer := &Player{Connected: true, State: Drawing, Role: RoleDrawer}
	disruptor := &Player{Connected: true, State: Guessing, Role: RoleDisruptor}
	guesser := &Player{Connected: true, State: Guessing}
	guessedDisruptor := &Player{Connected: true, State: Standby, Role: RoleDisruptor}

	require.True(t, lobby.canDraw(drawer))
	// The disruptor interferes via effects only and must never draw.
	require.False(t, lobby.canDraw(disruptor))
	require.True(t, lobby.isActiveDisruptor(disruptor))
	require.False(t, lobby.canDraw(guesser))
	require.False(t, lobby.canDraw(guessedDisruptor))

	// Stale roles must not grant permissions once the game is over.
	lobby.State = GameOver
	require.False(t, lobby.canDraw(disruptor))
	require.False(t, lobby.isActiveDisruptor(disruptor))
}

func Test_readyDataForDisruptor(t *testing.T) {
	t.Parallel()

	lobby := newDisruptorTestLobby()
	drawer := &Player{Connected: true, State: Drawing, Role: RoleDrawer}
	disruptor := &Player{Connected: true, State: Guessing, Role: RoleDisruptor}
	guesser := &Player{Connected: true, State: Guessing}
	lobby.players = []*Player{drawer, disruptor, guesser}

	drawerReady := generateReadyData(lobby, drawer)
	require.True(t, drawerReady.AllowDrawing)
	require.False(t, drawerReady.AllowDisruption)

	disruptorReady := generateReadyData(lobby, disruptor)
	require.False(t, disruptorReady.AllowDrawing)
	require.True(t, disruptorReady.AllowDisruption)

	guesserReady := generateReadyData(lobby, guesser)
	require.False(t, guesserReady.AllowDrawing)
	require.False(t, guesserReady.AllowDisruption)
}

func Test_disruptorEffectCooldownsScaleWithSeverity(t *testing.T) {
	t.Parallel()

	seen := make(map[string]bool)
	maxCooldown := make(map[int]time.Duration)
	minCooldown := make(map[int]time.Duration)
	for _, effect := range disruptorEffects {
		require.False(t, seen[effect.Name], "duplicate effect %s", effect.Name)
		seen[effect.Name] = true

		// The cooldown starts with the effect, so it has to outlast it.
		require.Greater(t, effect.Cooldown, effect.Duration, effect.Name)

		if effect.Cooldown > maxCooldown[effect.Severity] {
			maxCooldown[effect.Severity] = effect.Cooldown
		}
		if current, ok := minCooldown[effect.Severity]; !ok || effect.Cooldown < current {
			minCooldown[effect.Severity] = effect.Cooldown
		}
	}

	// Every effect of a more severe tier cools down longer than any effect
	// of the tier below.
	require.Less(t, maxCooldown[SeverityMild], minCooldown[SeverityMedium])
	require.Less(t, maxCooldown[SeverityMedium], minCooldown[SeveritySevere])
}

func Test_handleDisruptorEffect(t *testing.T) {
	t.Parallel()

	newLobby := func() (*Lobby, *Player, *Player, *Player) {
		lobby := newDisruptorTestLobby()
		drawer := &Player{Connected: true, State: Drawing, Role: RoleDrawer}
		disruptor := &Player{Connected: true, State: Guessing, Role: RoleDisruptor}
		guesser := &Player{Connected: true, State: Guessing}
		lobby.players = []*Player{drawer, disruptor, guesser}
		return lobby, drawer, disruptor, guesser
	}

	t.Run("only the disruptor can trigger effects", func(t *testing.T) {
		t.Parallel()

		lobby, drawer, _, guesser := newLobby()
		lobby.handleDisruptorEffect(drawer, DisruptorEffectShaky)
		lobby.handleDisruptorEffect(guesser, DisruptorEffectShaky)
		require.Empty(t, drawer.disruptorEffectReadyAt)
		require.Empty(t, guesser.disruptorEffectReadyAt)
		require.Zero(t, drawer.Score)
		require.Zero(t, guesser.Score)
	})

	t.Run("unknown effects and finished turns are ignored", func(t *testing.T) {
		t.Parallel()

		lobby, _, disruptor, _ := newLobby()
		lobby.handleDisruptorEffect(disruptor, "explode")
		require.Empty(t, disruptor.disruptorEffectReadyAt)

		lobby.CurrentWord = ""
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectShaky)
		require.Empty(t, disruptor.disruptorEffectReadyAt)
	})

	t.Run("cooldown is enforced per effect", func(t *testing.T) {
		t.Parallel()

		lobby, _, disruptor, _ := newLobby()
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectShaky)
		shakyReadyAt, used := disruptor.disruptorEffectReadyAt[DisruptorEffectShaky]
		require.True(t, used)
		require.True(t, shakyReadyAt.After(time.Now()))
		require.Equal(t, disruptorPointsPerAction, disruptor.Score)

		// Still cooling down: nothing changes.
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectShaky)
		require.Equal(t, shakyReadyAt, disruptor.disruptorEffectReadyAt[DisruptorEffectShaky])

		// Another effect has its own cooldown.
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectLag)
		_, used = disruptor.disruptorEffectReadyAt[DisruptorEffectLag]
		require.True(t, used)

		// Once the cooldown is over, the effect can be used again.
		disruptor.disruptorEffectReadyAt[DisruptorEffectShaky] = time.Now().Add(-time.Second)
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectShaky)
		require.True(t, disruptor.disruptorEffectReadyAt[DisruptorEffectShaky].After(time.Now()))
	})

	t.Run("undo removes the last stroke", func(t *testing.T) {
		t.Parallel()

		lobby, _, disruptor, _ := newLobby()

		// Nothing to undo: no effect and no cooldown.
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectUndo)
		require.Empty(t, disruptor.disruptorEffectReadyAt)

		lobby.currentDrawing = []any{&LineEvent{}, &LineEvent{}, &LineEvent{}}
		lobby.connectedDrawEventsIndexStack = []int{0, 2}
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectUndo)
		require.Len(t, lobby.currentDrawing, 2)
		require.Equal(t, []int{0}, lobby.connectedDrawEventsIndexStack)
		_, used := disruptor.disruptorEffectReadyAt[DisruptorEffectUndo]
		require.True(t, used)

		// The cooldown prevents undoing the next stroke right away.
		lobby.handleDisruptorEffect(disruptor, DisruptorEffectUndo)
		require.Len(t, lobby.currentDrawing, 2)
	})
}
