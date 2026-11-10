/// <reference types="@cloudflare/workers-types" />

declare module "cloudflare:workers" {
  export interface DurableObjectState {
    storage: DurableObjectStorage;
    blockConcurrencyWhile<T>(callback: () => Promise<T>): Promise<T>;
    waitUntil(promise: Promise<unknown>): void;
    acceptWebSocket(webSocket: WebSocket): void;
    getWebSockets(): WebSocket[];
  }

  export interface DurableObjectStorage {
    get<T>(key: string): Promise<T | undefined>;
    put(key: string, value: unknown): Promise<void>;
    delete(key: string): Promise<boolean>;
    list<T>(): Promise<Map<string, T>>;
    list<T>(options: { start?: string; end?: string; prefix?: string; limit?: number; reverse?: boolean }): Promise<Map<string, T>>;
  }

  export class DurableObjectNamespace<DO extends DurableObject = DurableObject> {
    idFromName(name: string): DurableObjectId;
    idFromString(id: string): DurableObjectId;
    get(id: DurableObjectId): DurableObjectStub<DO>;
    jurisdiction(jurisdiction: string): DurableObjectNamespace<DO>;
  }

  export interface DurableObjectId {
    toString(): string;
    equals(other: DurableObjectId): boolean;
  }

  export interface DurableObjectStub<DO extends DurableObject = DurableObject> {
    fetch(request: Request | string, init?: RequestInit): Promise<Response>;
    fetch(request: string, init?: RequestInit): Promise<Response>;
  }

  export class DurableObject<Env = unknown> {
    constructor(ctx: DurableObjectState, env: Env);
  }
}