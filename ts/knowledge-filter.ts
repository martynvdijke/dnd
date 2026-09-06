export type KnowledgeStatus = "rumor" | "confirmed" | "revealed" | "false" | "";

export function filterKnowledgeByStatus<T extends { status: string }>(items: T[], status: KnowledgeStatus): T[] {
  if (!status) return items;
  return items.filter((i) => i.status === status);
}

export function buildKnowledgeFilterParams(status: KnowledgeStatus): string {
  return status ? `?status=${status}` : "";
}
