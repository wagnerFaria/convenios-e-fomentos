export const auditActions = { read: 'admin::settings.audit-logs.read' } as const;

export interface AuditLog {
  id: string;
  actorType: string;
  actorId: string | null;
  /** Nome atual do administrador; nulo para outros atores, administrador excluído ou sem nome. */
  actorName: string | null;
  action: string;
  entity: string;
  entityId: string | null;
  payload: { before?: unknown; after?: unknown } & Record<string, unknown>;
  ip: string | null;
  createdAt: string;
}

export interface AuditListParams {
  action?: string;
  entity?: string;
  entityId?: string;
  actorId?: string;
  /** Datas em RFC 3339. */
  from?: string;
  to?: string;
  page?: number;
  pageSize?: number;
}
