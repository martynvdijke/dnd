/**
 * Single registration point for global (window) function exposure.
 *
 * All inline `(window as any).xxx = ...` assignments in app modules are
 * replaced by `expose('xxx', fn)`. This module is the ONLY place besides
 * ts/lib/bridge.ts that performs the raw window cast for registration,
 * keeping global namespace pollution explicit and auditable.
 *
 * This module intentionally has no imports to avoid circular dependencies.
 */
export function expose(name: string, value: unknown): void {
  (window as unknown as Record<string, unknown>)[name] = value;
}

export function exposeTyped<K extends keyof Window>(name: K, value: Window[K]): void {
  (window as unknown as Record<string, unknown>)[name as string] = value;
}
