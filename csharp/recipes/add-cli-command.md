# Recipe: Add CLI Command

Use this when a repo-owned binary gains a new subcommand or user-visible option set.

## Files To Touch

- `src/Orders.Cli/Commands/<Name>Command.cs` and the root-command wiring in `src/Orders.Cli/Program.cs`
- config/options classes if the command introduces new config or env integration (follow [add-config-key.md](add-config-key.md))
- the owning `Orders.Core` class if the command triggers domain behavior
- CLI tests and README/help text when user-visible behavior changes

## Steps

1. Build the command with System.CommandLine's GA API: `Command`, `Option<T>`, `Argument<T>`, and `SetAction` — never the pre-GA `SetHandler`. Keep the command class thin: declare symbols, parse, delegate.

   ```csharp
   var idOption = new Option<Guid>("--order-id") { Description = "Order to cancel.", Required = true };
   var cancel = new Command("cancel", "Cancel an order.") { idOption };
   cancel.SetAction(async (parseResult, cancellationToken) =>
   {
       var id = parseResult.GetValue(idOption);
       return await canceller.CancelAsync(id, cancellationToken) ? 0 : 1;
   });
   ```

2. Read values with `parseResult.GetValue(symbol)` and pass validated, typed values into `Orders.Core` — no parsing or domain decisions inside the action.
3. Return a non-zero exit code from the action on failure and write the reason to stderr; do not hide failures behind logs only. Async actions return `Task<int>` and honor the injected `CancellationToken` (Ctrl+C).
4. Wire the command onto the root in `Program.cs` (`rootCommand.Subcommands.Add(cancel)`; entry point stays `return rootCommand.Parse(args).Invoke();` or `InvokeAsync`).
5. Update descriptions and usage examples if the command is user-facing — `--help` output is part of the contract.

## Invariants To Preserve

- business logic stays out of `Orders.Cli` — the action parses, delegates to Core, and maps the result to an exit code
- option and env precedence remain documented and predictable
- `--version` and `--help` output still work from the built artifact
- every action honors cancellation; no `SetHandler`, no blocking `.Result`/`.Wait()`

## Proof

- parsing tests: `command.Parse("cancel --order-id …")` asserting `GetValue` results, and a negative case asserting `parseResult.Errors` is non-empty for missing/malformed input
- `dotnet build src/Orders.Cli` (clean under warnings-as-errors), then run `pwsh ./verify.ps1`
- local `--help` smoke test against the built artifact: `dotnet run --project src/Orders.Cli -- cancel --help`
- one success and one failure-path execution against the built binary, asserting exit codes `0` and non-zero
