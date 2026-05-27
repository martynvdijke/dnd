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
