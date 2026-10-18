// Disruptor effects.
//
// The disruptor can't draw. Instead, they press buttons, which temporarily
// mess with the DRAWER's pen. The server validates the button press (role,
// cooldown) and announces it to everyone via a "disruptor-effect" event. This
// file is the part that runs on the drawer's client and makes the effect
// actually happen for a limited time. Strokes are sent to the server like any
// other line, so everybody simply sees the messy result.
//
// Effects (the names match internal/game/disruptor.go):
//  - shaky:        random jitter on the stroke coordinates.
//  - lag:          the pen trails behind the cursor and catches up afterwards.
//  - flip:         each stroke is mirrored or rotated by 90°, so moving the
//                  mouse right might draw left, up or down.
//  - random-size:  the brush size keeps changing on its own.
//  - random-color: the brush color keeps changing on its own.
//  - eraser:       the pen is forced to be the eraser.
//  - undo:         handled by the server entirely, nothing to do on the client
//                  except for showing a notification.
//
// Effects that touch the toolbar (size, color, eraser) need access to the
// state that lives in lobby.js. That's what init(hooks) is for.

const disruptorEffects = (() => {
    const BRUSH_SIZES = [8, 16, 24, 32];
    const FLIP_MODES = ["mirrorX", "mirrorY", "rot90", "rot270"];
    // Color indices 1-25 are real colors, 0 is white and would look like an
    // eraser.
    const FIRST_RANDOM_COLOR_INDEX = 1;
    const COLOR_COUNT = 26;

    // The brush size and color change every couple of hundred milliseconds,
    // which is often enough to ruin a line, but not so often that it flickers.
    const RANDOM_CHANGE_MIN_MS = 450;
    const RANDOM_CHANGE_MAX_MS = 1100;

    // How much of the remaining distance the lagging pen closes per step.
    const LAG_FACTOR = 0.18;
    const LAG_STEP_MS = 16;
    const LAG_MAX_CATCHUP_STEPS = 90;
    // A stroke is considered continued, if the next segment starts close to
    // where the previous one ended and happens shortly after it.
    const STROKE_CONTINUE_MAX_GAP_MS = 250;
    const STROKE_CONTINUE_MAX_DISTANCE = 4;

    // Human readable names for buttons and notifications.
    const LABELS = {
        shaky: "Shaky hand",
        lag: "Laggy pen",
        flip: "Flipped canvas",
        "random-size": "Random size",
        "random-color": "Random color",
        eraser: "Forced eraser",
        undo: "Undo",
    };

    // Effects that run for a while and can be started and stopped here.
    // "undo" is missing on purpose, as it is instant and server side.
    const TIMED_EFFECTS = new Set([
        "shaky",
        "lag",
        "flip",
        "random-size",
        "random-color",
        "eraser",
    ]);

    // hooks are provided by lobby.js, see init.
    let hooks = null;

    // Effect name -> true, while the effect is running.
    const active = {};
    // Effect name -> timeout, that ends the effect.
    const expiryTimers = {};
    // Effect name -> timeout of the ticker, that keeps changing size / color.
    const tickerTimers = {};
    // What the drawer had selected before random-size / random-color kicked
    // in, so we can give it back afterwards.
    const saved = { size: null, colorIndex: null };

    let stroke = null;
    let catchupTimer = null;

    function randomInt(maxExclusive) {
        return Math.floor(Math.random() * maxExclusive);
    }

    function randomBetween(min, max) {
        return min + Math.random() * (max - min);
    }

    // hooks:
    //   getSize(), setSize(px)
    //   getColorIndex(), setColorIndex(index)
    //   forceEraser(bool)
    //   onChange(activeEffectNames)  optional, e.g. for updating the UI
    function init(newHooks) {
        hooks = newHooks;
    }

    function labelFor(effect) {
        return LABELS[effect] ?? effect;
    }

    function isActive(effect) {
        return active[effect] === true;
    }

    function needsStrokeProcessing() {
        return (
            active.shaky === true || active.lag === true || active.flip === true
        );
    }

    // ---- stroke processing (shaky, lag, flip) ----

    function cancelCatchup() {
        if (catchupTimer !== null) {
            clearTimeout(catchupTimer);
            catchupTimer = null;
        }
    }

    // Finishes the pending lag movement right away, so the line ends where
    // the cursor is and not somewhere behind it.
    function flushLag() {
        cancelCatchup();
        if (stroke === null || stroke.emit === null) {
            return;
        }
        if (
            Math.hypot(
                stroke.targetX - stroke.lagX,
                stroke.targetY - stroke.lagY,
            ) < 0.5
        ) {
            return;
        }
        stroke.lagX = stroke.targetX;
        stroke.lagY = stroke.targetY;
        emitStep();
    }

    // Starts the next segment from scratch. Needed whenever the way input is
    // transformed changes, as the next segment would otherwise connect the
    // old, transformed pen position with the new one via a long line.
    function endStroke() {
        cancelCatchup();
        stroke = null;
    }

    function transformPoint(mode, x, y, width, height) {
        if (!mode) {
            return { x, y };
        }
        // We work with normalized coordinates, as the canvas is not square
        // and a plain rotation would otherwise leave the drawing area.
        const u = x / width;
        const v = y / height;
        let nu = u;
        let nv = v;
        switch (mode) {
            case "mirrorX":
                nu = 1 - u;
                break;
            case "mirrorY":
                nv = 1 - v;
                break;
            case "rot90":
                nu = 1 - v;
                nv = u;
                break;
            case "rot270":
                nu = v;
                nv = 1 - u;
                break;
        }
        return { x: nu * width, y: nv * height };
    }

    function jitterPoint(point, brushWidth) {
        if (!active.shaky) {
            return point;
        }
        const amplitude = 2 + brushWidth * 0.2;
        return {
            x: point.x + (Math.random() * 2 - 1) * amplitude,
            y: point.y + (Math.random() * 2 - 1) * amplitude,
        };
    }

    function clampPoint(point, width, height) {
        return {
            x: Math.min(width - 1, Math.max(0, point.x)),
            y: Math.min(height - 1, Math.max(0, point.y)),
        };
    }

    function outputPoint(x, y) {
        const transformed = transformPoint(
            active.flip ? stroke.flipMode : null,
            x,
            y,
            stroke.width,
            stroke.height,
        );
        const jittered = jitterPoint(transformed, stroke.brushWidth);
        return clampPoint(jittered, stroke.width, stroke.height);
    }

    function startStroke(x, y, brushWidth, width, height, emit) {
        stroke = {
            flipMode: FLIP_MODES[randomInt(FLIP_MODES.length)],
            brushWidth: brushWidth,
            width: width,
            height: height,
            emit: emit,
            lagX: x,
            lagY: y,
            targetX: x,
            targetY: y,
            lastInputX: x,
            lastInputY: y,
            lastInputTime: performance.now(),
            outX: 0,
            outY: 0,
        };
        const start = outputPoint(x, y);
        stroke.outX = start.x;
        stroke.outY = start.y;
    }

    function emitStep() {
        const end = outputPoint(stroke.lagX, stroke.lagY);
        stroke.emit(stroke.outX, stroke.outY, end.x, end.y, stroke.brushWidth);
        stroke.outX = end.x;
        stroke.outY = end.y;
    }

    function step() {
        if (active.lag) {
            stroke.lagX += (stroke.targetX - stroke.lagX) * LAG_FACTOR;
            stroke.lagY += (stroke.targetY - stroke.lagY) * LAG_FACTOR;
        } else {
            stroke.lagX = stroke.targetX;
            stroke.lagY = stroke.targetY;
        }
        emitStep();
    }

    function scheduleCatchup(stepsLeft) {
        cancelCatchup();
        if (!active.lag || stepsLeft <= 0 || stroke === null) {
            return;
        }
        const remaining = Math.hypot(
            stroke.targetX - stroke.lagX,
            stroke.targetY - stroke.lagY,
        );
        if (remaining < 1.5) {
            return;
        }
        catchupTimer = setTimeout(() => {
            catchupTimer = null;
            if (stroke === null) {
                return;
            }
            step();
            scheduleCatchup(stepsLeft - 1);
        }, LAG_STEP_MS);
    }

    // Takes a segment as the drawer moved the pen and emits the disrupted
    // segment(s) via emit(x1, y1, x2, y2, brushWidth). The coordinates are in
    // client space of the drawing board. Only call this while
    // needsStrokeProcessing() is true.
    function process(x1, y1, x2, y2, brushWidth, width, height, emit) {
        const now = performance.now();
        const continuing =
            stroke !== null &&
            now - stroke.lastInputTime < STROKE_CONTINUE_MAX_GAP_MS &&
            Math.hypot(x1 - stroke.lastInputX, y1 - stroke.lastInputY) <=
                STROKE_CONTINUE_MAX_DISTANCE;

        if (!continuing) {
            cancelCatchup();
            startStroke(x1, y1, brushWidth, width, height, emit);
        }

        stroke.brushWidth = brushWidth;
        stroke.emit = emit;
        stroke.targetX = x2;
        stroke.targetY = y2;
        stroke.lastInputX = x2;
        stroke.lastInputY = y2;
        stroke.lastInputTime = now;

        step();
        scheduleCatchup(LAG_MAX_CATCHUP_STEPS);
    }

    // ---- random size / color ----

    function stopTicker(effect) {
        if (tickerTimers[effect] !== undefined) {
            clearTimeout(tickerTimers[effect]);
            delete tickerTimers[effect];
        }
    }

    function startTicker(effect, tick) {
        stopTicker(effect);
        const run = () => {
            if (!active[effect]) {
                return;
            }
            tick();
            tickerTimers[effect] = setTimeout(
                run,
                randomBetween(RANDOM_CHANGE_MIN_MS, RANDOM_CHANGE_MAX_MS),
            );
        };
        // The first change happens immediately, so the drawer notices that
        // something is going on.
        run();
    }

    function tickRandomSize() {
        const current = hooks.getSize();
        const options = BRUSH_SIZES.filter((size) => size !== current);
        hooks.setSize(options[randomInt(options.length)]);
    }

    function tickRandomColor() {
        const current = hooks.getColorIndex();
        let next = current;
        // Make sure that the change is visible.
        while (next === current) {
            next =
                FIRST_RANDOM_COLOR_INDEX +
                randomInt(COLOR_COUNT - FIRST_RANDOM_COLOR_INDEX);
        }
        hooks.setColorIndex(next);
    }

    // ---- starting and stopping effects ----

    function notifyChange() {
        if (hooks?.onChange) {
            hooks.onChange(Object.keys(active).filter((name) => active[name]));
        }
    }

    function start(effect) {
        switch (effect) {
            case "shaky":
            case "lag":
            case "flip":
                // The way input is mapped to output changes from now on.
                endStroke();
                break;
            case "random-size":
                saved.size = hooks.getSize();
                break;
            case "random-color":
                saved.colorIndex = hooks.getColorIndex();
                break;
            case "eraser":
                hooks.forceEraser(true);
                break;
        }
        active[effect] = true;

        if (effect === "random-size") {
            startTicker(effect, tickRandomSize);
        } else if (effect === "random-color") {
            startTicker(effect, tickRandomColor);
        }
    }

    function stop(effect) {
        if (!active[effect]) {
            return;
        }
        // Flush while the effect is still marked as active, so that the
        // pending lag movement is emitted as it would have been.
        if (effect === "lag") {
            flushLag();
        }
        delete active[effect];

        switch (effect) {
            case "shaky":
            case "lag":
            case "flip":
                endStroke();
                break;
            case "random-size":
                stopTicker(effect);
                if (saved.size !== null) {
                    hooks.setSize(saved.size);
                    saved.size = null;
                }
                break;
            case "random-color":
                stopTicker(effect);
                if (saved.colorIndex !== null) {
                    hooks.setColorIndex(saved.colorIndex);
                    saved.colorIndex = null;
                }
                break;
            case "eraser":
                hooks.forceEraser(false);
                break;
        }
    }

    // Applies an effect for durationMs. Using an effect that is still running
    // restarts its timer, instead of stacking it.
    function apply(effect, durationMs) {
        if (hooks === null || !TIMED_EFFECTS.has(effect)) {
            return;
        }

        if (expiryTimers[effect] !== undefined) {
            clearTimeout(expiryTimers[effect]);
            delete expiryTimers[effect];
        }
        if (!active[effect]) {
            start(effect);
        }

        expiryTimers[effect] = setTimeout(() => {
            delete expiryTimers[effect];
            stop(effect);
            notifyChange();
        }, durationMs);

        notifyChange();
    }

    // Ends every running effect right away and gives the drawer back
    // their pen. Called when the turn ends or the drawer changes.
    function clearAll() {
        for (const effect of Object.keys(expiryTimers)) {
            clearTimeout(expiryTimers[effect]);
            delete expiryTimers[effect];
        }
        for (const effect of Object.keys(active)) {
            stop(effect);
        }
        endStroke();
        notifyChange();
    }

    return {
        init,
        apply,
        clearAll,
        isActive,
        labelFor,
        needsStrokeProcessing,
        process,
    };
})();
