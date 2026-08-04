import { showView } from './navigation';
import { updateFabForView } from './fab';
import { expose } from './lib/expose';

export function openCharacter(id: number): void {
  showView('sheet');
  updateFabForView('sheet');
}

export function renderQuickActions(el: HTMLElement, c: any): void {
  el.innerHTML = `
    <div class="quick-actions-bar">
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('short')">
        <i class="fa-solid fa-campground me-1"></i>Short Rest
      </button>
      <button class="btn btn-sm btn-outline-primary" onclick="doRest('long')">
        <i class="fa-solid fa-moon me-1"></i>Long Rest
      </button>
      <button class="btn btn-sm btn-gold" onclick="doLevelUp()">
        <i class="fa-solid fa-arrow-up me-1"></i>Level Up
      </button>
    </div>
  `;
}

export function renderSessionQuickActions(el: HTMLElement): void {
  el.innerHTML = `
    <div class="quick-actions-bar session-actions">
      <div class="input-group input-group-sm">
        <input type="number" class="form-control" id="dmgInput" value="0" placeholder="Damage">
        <button class="btn btn-danger" onclick="applyDamage()">Damage</button>
      </div>
      <div class="input-group input-group-sm">
        <input type="number" class="form-control" id="healInput" value="0" placeholder="Heal">
        <button class="btn btn-success" onclick="applyHeal()">Heal</button>
      </div>
      <button class="btn btn-sm btn-outline-primary" onclick="showAddCondition()">
        <i class="fa-solid fa-plus me-1"></i>Condition
      </button>
    </div>
  `;
}

expose('openCharacter', openCharacter);
