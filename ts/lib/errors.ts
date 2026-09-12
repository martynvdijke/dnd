import { toast } from './dom';

export function renderError(e: unknown): void {
  const msg = e instanceof Error ? (e as Error).message : String((e as any)?.message ?? e);
  toast(msg, true);
}
