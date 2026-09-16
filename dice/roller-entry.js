/**
 * Dice roller bundle entry point — the Go <-> JS contract.
 *
 * The Go side (dice/engine.go) loads this bundle in goja and relies on the
 * following globals. Changing any of them breaks the Go caller; the contract is
 * asserted by `TestDiceContract` in dice/contract_test.go.
 *
 *   globalThis.__diceRoll(expression: string): string
 *     Rolls `expression` (e.g. "2d6+3") and returns a JSON string of
 *     DiceRoll.toJSON(), e.g.
 *       {"notation":"2d6","total":6,"rolls":[...],"output":"2d6: [1, 5] = 6",
 *        "averageTotal":7,"maxTotal":12,"minTotal":2,"type":"dice-roll"}
 *     On a parse/roll error it returns {"error":"<message>"} instead of throwing.
 *
 *   globalThis.__diceRoller: DiceRoller
 *     The rpg-dice-roller instance backing __diceRoll. Exposed for direct use;
 *     Go calls __diceRoll, not this.
 *
 *   globalThis.__diceBuildStamp: string
 *     The pinned @dice-roller/rpg-dice-roller version, injected at build time by
 *     the `build:dice` esbuild `--banner:js` flag. dice/engine.go verifies it at
 *     engine init and fails fast if the embedded bundle is stale.
 */
import { DiceRoller } from '@dice-roller/rpg-dice-roller';

// Global registry so Go can call roll()
globalThis.__diceRoller = new DiceRoller();

globalThis.__diceRoll = function(expression) {
  try {
    const roll = globalThis.__diceRoller.roll(expression);
    const json = roll.toJSON();
    return JSON.stringify(json);
  } catch (e) {
    return JSON.stringify({ error: e.message });
  }
};
