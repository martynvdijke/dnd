// @ts-nocheck — legacy helper extracted from untyped monolith
import { api } from './api';
import { currentChar, setCurrentChar } from './state';

export async function refreshChar(id?: string): Promise<void> {
  const cid = id ?? currentChar?.id;
  if (!cid) return;
  setCurrentChar(await api('GET', `/api/characters/${cid}`));
}
