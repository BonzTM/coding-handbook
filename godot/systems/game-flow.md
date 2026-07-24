# Game Flow

Game-flow defaults for Godot 4.x projects: engine-managed scene transitions unless a fade or persistent overlay demands a manual swap, cross-scene survivors as autoloads only, pause via `process_mode`, loading screens on the threaded loader with polling, and one quit path through `NOTIFICATION_WM_CLOSE_REQUEST`.

## Default Approach

The scene tree has exactly one current scene plus a small, fixed set of persistent autoloads. Everything that must outlive a transition — music, the transition overlay, the save service — is an autoload registered via [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md); everything else dies with its scene. Transitions go through `get_tree().change_scene_to_packed()` by default; the manual swap-under-root pattern is the documented exception, not a parallel system. Pause state, loading, and quitting each have one owner, so no scene ever asks "who unpauses the game?"

### Scene Transitions

- **Default: `change_scene_to_packed(packed_scene)`; `change_scene_to_file(path)` only for cold paths.** Both return an `Error` — `change_scene_to_packed` "Changes the running scene to a new instance of the given PackedScene (which must be valid)" and returns `ERR_CANT_CREATE` / `ERR_INVALID_PARAMETER` on failure (`docs.godotengine.org/en/stable/classes/class_scenetree.html`). Check the return value; a failed transition that scrolls past as a warning is a black screen in production. Prefer the packed variant because a loading screen has usually already produced the `PackedScene` via the threaded loader below — `change_scene_to_file` re-loads from disk on the spot.
- **Know the teardown order before writing code after the call.** The docs specify: "The current scene node is immediately removed from the tree. From that point, Node.get_tree() called on the current (outgoing) scene will return null. current_scene will be null too" — the old scene is freed and the new one added only "at the end of the frame," which "ensures that both scenes aren't running at the same time, while still freeing the previous scene in a safe way similar to Node.queue_free()" (`docs.godotengine.org/en/stable/classes/class_scenetree.html`). Consequence: in the outgoing scene, nothing after the `change_scene_*` call may touch `get_tree()`, groups, or tree-dependent state.
- **Manual swap under a persistent root is for transitions that must render across the change** — a fade, an animated wipe, a HUD that stays up. The pattern is the official scene-switcher example (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`): the transition autoload defers the swap (`_deferred_goto_scene.call_deferred(path)`) because "Deleting the current scene at this point is a bad idea, because it may still be executing code"; the deferred function frees the current scene, instantiates the new `PackedScene`, adds it under `get_tree().root`, and reassigns `get_tree().current_scene` so the built-in `change_scene_*` API and `reload_current_scene()` keep working.
- **One owner for manual swaps.** If the project adopts the manual pattern, the transition autoload is the only code that frees or replaces `current_scene`; gameplay scenes request a transition by calling it (or signaling an event bus per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md)) and never swap scenes themselves.
- **Removing is not the only option.** The manual-change tutorial's trade-off table applies: freeing the scene releases memory and CPU but "returning requires reloading"; hiding keeps data live at a memory cost; `remove_child()` detaches without deleting but the node gets no delta, input, or group updates while outside the tree (`docs.godotengine.org/en/stable/tutorials/scripting/change_scenes_manually.html`). Pick per scene and write the choice down; a detached scene someone forgot to re-add or free is a leak.

### The Persistent Set: What Survives A Transition

- **Autoloads are the survival mechanism.** They are children of `root` created before the main scene, they "are always loaded, no matter which scene is currently running," and "the last child of root is always the loaded scene" (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`). Music keeps playing through a transition because the music controller from [audio.md](audio.md) is an autoload, not because any scene carries it.
- **The persistent set is small and named**: the music/ambience controller ([audio.md](audio.md)), the transition overlay (a `CanvasLayer` with the fade rect, layered above every scene), and the save service ([save-and-load.md](save-and-load.md)). Each one passed the autoload-restraint bar in [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md); "it must survive scene changes" is a valid justification, "it's convenient globally" is not.
- **Never free an autoload.** "Autoloads must not be removed using free() or queue_free() at runtime, or the engine will crash" (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`).
- **Never park cross-scene state on the current scene** or hand a node reference from the old scene to the new one — the old scene is freed at end of frame and the reference dangles. State that crosses a transition goes through an autoload or a `Resource`.

### Pause

- **Pause is `get_tree().paused = true`, controlled from exactly one place** (a pause autoload or the pause-menu scene). When paused, physics stops and, for a non-processing node, "The `_process`, `_physics_process`, `_input`, and `_input_event` functions will not be called," while signals keep working (`docs.godotengine.org/en/stable/tutorials/scripting/pausing_games.html`).
- **Behavior under pause is declared per node via `process_mode`** — `PROCESS_MODE_INHERIT`, `PAUSABLE`, `WHEN_PAUSED`, `ALWAYS`, `DISABLED` (`docs.godotengine.org/en/stable/classes/class_node.html`). Gameplay stays on the default inherit-from-pausable chain; the pause menu is `PROCESS_MODE_WHEN_PAUSED` so it can process input while everything else is frozen — the tutorial's pattern (`docs.godotengine.org/en/stable/tutorials/scripting/pausing_games.html`). A menu left on the default mode pauses itself and the game is stuck.
- **`PROCESS_MODE_ALWAYS` is for the persistent set only** — music and the transition overlay keep running while paused. Marking a gameplay node `ALWAYS` to dodge a pause bug hides the bug; note the tutorial's warning that even a processing node gets no physics while the tree is paused.
- **Unpause before or during every transition.** `paused` is `SceneTree` state and survives scene changes; a menu scene loaded with `paused == true` is frozen. The transition owner resets it.

### Loading Screens And Threaded Loading

The blocking pattern — `load()` on the main thread — freezes the frame; a loading screen requires the threaded API (`docs.godotengine.org/en/stable/tutorials/io/background_loading.html`):

```gdscript
const NEXT := "res://levels/level_02.tscn"

func start_load() -> void:
	var err := ResourceLoader.load_threaded_request(NEXT)
	assert(err == OK)

func _process(_delta: float) -> void:
	var progress: Array = []
	match ResourceLoader.load_threaded_get_status(NEXT, progress):
		ResourceLoader.THREAD_LOAD_IN_PROGRESS:
			_bar.value = progress[0]  # 0.0 .. 1.0
		ResourceLoader.THREAD_LOAD_LOADED:
			var scene: PackedScene = ResourceLoader.load_threaded_get(NEXT)
			get_tree().change_scene_to_packed(scene)
		_:
			_fail_to_menu()  # THREAD_LOAD_FAILED / THREAD_LOAD_INVALID_RESOURCE
```

- **Poll status; never call `load_threaded_get()` early.** The docs are explicit that if it is called before loading finishes, "the calling thread will be blocked until the resource has finished loading" (`docs.godotengine.org/en/stable/classes/class_resourceloader.html`) — exactly the freeze the loading screen exists to avoid. `load_threaded_get_status()` fills an optional one-element progress array with the completion ratio for the progress bar.
- **Handle both failure states.** `THREAD_LOAD_FAILED` and `THREAD_LOAD_INVALID_RESOURCE` are real outcomes (`docs.godotengine.org/en/stable/classes/class_resourceloader.html`); a loading screen that only handles success spins forever on a broken path.
- **Leave `use_sub_threads` at its default `false`.** The docs sell it as faster but warn it "may affect the main thread (and thus cause game slowdowns)" (`docs.godotengine.org/en/stable/classes/class_resourceloader.html`), and it has a multi-version track record of hangs and deadlocks: silent hangs with no error (`github.com/godotengine/godot/issues/85255`), a `load_threaded_get` deadlock with C# scripts (`github.com/godotengine/godot/issues/103674`), and crashes on scenes with many dependencies (`github.com/godotengine/godot/issues/118185`). Enabling it is a measured, ADR-recorded exception per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md). C# projects loading scripts with generic base classes from worker threads have their own deadlock history (`github.com/godotengine/godot/issues/99839`).
- **The threaded loader is for `res://` content only.** It is still `ResourceLoader`; the untrusted-file rule in [save-and-load.md](save-and-load.md) applies unchanged — no `user://`, mod, or downloaded paths.

### Quitting Cleanly

- **One handler owns quitting.** A single autoload (typically the save service) implements `_notification()` for `NOTIFICATION_WM_CLOSE_REQUEST`, flushes state through [save-and-load.md](save-and-load.md), then calls `get_tree().quit()` (`docs.godotengine.org/en/stable/tutorials/inputs/handling_quit_requests.html`). Set `get_tree().auto_accept_quit = false` only if the handler must be able to cancel the quit (confirmation dialog); otherwise the OS request also quits directly and the handler merely gets notified.
- **In-game Quit buttons do not call `quit()` directly.** The tutorial warns that calling quit directly "will not allow custom actions to complete (such as saving, confirming the quit, or debugging)"; instead send `get_tree().root.propagate_notification(NOTIFICATION_WM_CLOSE_REQUEST)` so every node — including the one handler — runs its close path first (`docs.godotengine.org/en/stable/tutorials/inputs/handling_quit_requests.html`).
- **Mobile has no close button.** iOS and Android deliver `NOTIFICATION_APPLICATION_PAUSED` on suspend instead, and "On iOS, you only have approximately 5 seconds to finish a task started by this signal" — save immediately and synchronously there. Android's Back button raises `NOTIFICATION_WM_GO_BACK_REQUEST` while `quit_on_go_back` (default on) quits (`docs.godotengine.org/en/stable/tutorials/inputs/handling_quit_requests.html`); a shipping Android title decides deliberately whether Back means quit.
- **`quit()` exits "at the end of the current iteration"** with an exit code that "should be between 0 and 125" for portability (`docs.godotengine.org/en/stable/classes/class_scenetree.html`) — nonzero codes are how a headless test run reports failure to CI.

## Common Mistakes And Forbidden Patterns

- **Ignoring the `Error` from `change_scene_to_*`**, shipping a transition that can silently leave a dead tree.
- **Touching `get_tree()` or the tree after calling `change_scene_*` from the outgoing scene** — the scene is already removed and `get_tree()` is null per the documented teardown order.
- **Freeing the current scene without deferring** — it "may still be executing code"; the swap is `call_deferred` or it is a use-after-free.
- **A manual swap that forgets `get_tree().current_scene`**, silently breaking `change_scene_to_file()` and `reload_current_scene()` for every other caller.
- **Carrying node references or parking shared state on the dying scene** instead of an autoload or `Resource`.
- **`free()`/`queue_free()` on an autoload** — documented engine crash.
- **A pause menu on the default `process_mode`**, frozen by the pause it triggered; or gameplay nodes marked `PROCESS_MODE_ALWAYS` to paper over pause bugs.
- **Leaving `paused = true` across a transition**, delivering a frozen next scene.
- **Calling `load_threaded_get()` without polling status** — a "loading screen" that blocks exactly like `load()`.
- **`use_sub_threads = true` by reflex**, buying documented hangs and deadlocks for an unmeasured speedup, with no ADR.
- **Quit buttons calling `get_tree().quit()` directly**, skipping every node's close-request path including the save flush.
- **Multiple systems flipping `paused`, swapping scenes, or handling quit** — each of the three has exactly one owner.

## Verification And Proof

Tests run under the framework in [../quality/testing.md](../quality/testing.md), headless in CI per [../operations/ci-and-release.md](../operations/ci-and-release.md).

- **Transition lifecycle test**: hold a `weakref()` to the current scene, transition, await a frame, and assert the old scene is freed, the new scene is `current_scene`, and every autoload in the persistent set is still present.
- **Return-value test**: transition to an invalid `PackedScene` and assert the `Error` propagates to the caller instead of vanishing.
- **Pause matrix test**: set `paused = true` and assert a `PAUSABLE` gameplay node stops receiving `_process` while the `WHEN_PAUSED` menu receives input; assert `paused == false` after a transition completes.
- **Loading-screen test**: `load_threaded_request` a real scene, poll status in a bounded loop (fail the test past N frames), assert `THREAD_LOAD_LOADED` then successful instantiation; feed a bogus path and assert the failure branch runs instead of spinning.
- **Quit-path test**: call `get_tree().root.propagate_notification(NOTIFICATION_WM_CLOSE_REQUEST)` and assert the save service flushed (fresh file mtime or content) before exit.
- **Static sweep**: `grep -rn --include="*.gd" "use_sub_threads\s*=\s*true\|load_threaded_request(.*,\s*true" .` returns nothing, or every hit cites an ADR; `grep -rn --include="*.gd" "get_tree().quit()" .` hits only the quit handler and test code.

## Related

- [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) — scene self-containment, the reason a scene never swaps its own successor's internals.
- [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) — autoload restraint bar and the event-bus route for requesting transitions.
- [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — registering the transition owner, music controller, and save service.
- [audio.md](audio.md) — the music controller that makes music survive transitions.
- [save-and-load.md](save-and-load.md) — the save flush behind the quit path and the untrusted-file rule the threaded loader inherits.
- [ui-and-theming.md](ui-and-theming.md) — building the loading screen and pause menu as Control scenes.
