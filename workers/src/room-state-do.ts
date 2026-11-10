import { DurableObject } from "cloudflare:workers";

export interface RoomState {
  id: string;
  organizationId: string;
  name: string;
  accessKeyHash: string;
  language: string;
  status: RoomStatus;
  config: RoomConfig;
  assignedBotId?: string;
  botStatus?: BotStatus;
  lastBotStatusUpdate: number;
  connectedUsers: Map<string, ConnectedUser>;
  metadata: Record<string, unknown>;
}

export type RoomStatus = 'connected' | 'disconnected' | 'reconnecting' | 'error';
export type BotStatus = 'stopped' | 'starting' | 'running' | 'stopping' | 'error' | 'reconnecting';

export interface RoomConfig {
  autoReconnect: boolean;
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  proxyUrl?: string;
}

export interface ConnectedUser {
  userId: string;
  username: string;
  connectedAt: number;
  lastActivity: number;
  ipAddress?: string;
  userAgent?: string;
}

export interface RoomStateDO extends DurableObject<Env> {
  // State management
  getState(): Promise<RoomState | null>;
  setState(state: RoomState): Promise<void>;
  
  // Status management
  updateStatus(status: RoomStatus, details?: Record<string, unknown>): Promise<RoomState>;
  getStatus(): Promise<RoomStatus>;
  
  // Configuration
  updateConfig(config: Partial<RoomConfig>): Promise<RoomConfig>;
  getConfig(): Promise<RoomConfig>;
  
  // Bot assignment
  assignBot(botId: string): Promise<void>;
  unassignBot(): Promise<void>;
  updateBotStatus(botStatus: BotStatus): Promise<void>;
  
  // User presence
  addUser(user: ConnectedUser): Promise<void>;
  removeUser(userId: string): Promise<void>;
  updateUserActivity(userId: string): Promise<void>;
  getConnectedUsers(): Promise<ConnectedUser[]>;
  
  // Room operations
  pause(pause: boolean): Promise<void>;
  setLanguage(locale: string): Promise<void>;
  
  // History
  getHistory(limit?: number): Promise<RoomEvent[]>;
}

export interface RoomEvent {
  timestamp: number;
  type: 'status_change' | 'bot_assigned' | 'bot_unassigned' | 'user_joined' | 'user_left' | 'paused' | 'language_changed' | 'config_changed';
  details: Record<string, unknown>;
}

export class RoomStateDOImpl implements RoomStateDO {
  private state: RoomState;

  constructor(ctx: DurableObjectState, env: Env) {
    this.state = {
      id: '',
      organizationId: '',
      name: '',
      accessKeyHash: '',
      language: 'en_US.UTF-8',
      status: 'disconnected',
      config: {
        autoReconnect: true,
        logLevel: 'info',
      },
      connectedUsers: new Map(),
      metadata: {},
    };
  }

  async getState(): Promise<RoomState | null> {
    const stored = await this.ctx.storage.get<RoomState>('room-state');
    return stored || this.state;
  }

  async setState(state: RoomState): Promise<void> {
    this.state = state;
    await this.ctx.storage.put('room-state', state);
  }

  async updateStatus(status: RoomStatus, details: Record<string, unknown> = {}): Promise<RoomState> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.status = status;
    state.metadata.lastStatusUpdate = Date.now();
    state.metadata.lastStatusDetails = details;
    
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'status_change',
      details: { status, ...details },
    });
    
    await this.setState(state);
    return state;
  }

  async getStatus(): Promise<RoomStatus> {
    const state = await this.getState();
    return state?.status || 'disconnected';
  }

  async updateConfig(config: Partial<RoomConfig>): Promise<RoomConfig> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.config = { ...state.config, ...config };
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'config_changed',
      details: config,
    });
    await this.setState(state);
    return state.config;
  }

  async getConfig(): Promise<RoomConfig> {
    const state = await this.getState();
    return state?.config || this.state.config;
  }

  async assignBot(botId: string): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.assignedBotId = botId;
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'bot_assigned',
      details: { botId },
    });
    await this.setState(state);
  }

  async unassignBot(): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    const oldBotId = state.assignedBotId;
    state.assignedBotId = undefined;
    state.botStatus = undefined;
    
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'bot_unassigned',
      details: { previousBotId: oldBotId },
    });
    await this.setState(state);
  }

  async updateBotStatus(botStatus: BotStatus): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.botStatus = botStatus;
    state.lastBotStatusUpdate = Date.now();
    await this.setState(state);
  }

  async addUser(user: ConnectedUser): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.connectedUsers.set(user.userId, user);
    
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'user_joined',
      details: { userId: user.userId, username: user.username },
    });
    
    await this.setState(state);
  }

  async removeUser(userId: string): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    const user = state.connectedUsers.get(userId);
    state.connectedUsers.delete(userId);
    
    if (user) {
      await this.recordEvent({
        timestamp: Date.now(),
        type: 'user_left',
        details: { userId, username: user.username },
      });
      await this.setState(state);
    }
  }

  async updateUserActivity(userId: string): Promise<void> {
    const state = await this.getState();
    if (!state) return;

    const user = state.connectedUsers.get(userId);
    if (user) {
      user.lastActivity = Date.now();
      await this.setState(state);
    }
  }

  async getConnectedUsers(): Promise<ConnectedUser[]> {
    const state = await this.getState();
    if (!state) return [];
    return Array.from(state.connectedUsers.values());
  }

  async pause(pause: boolean): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    await this.recordEvent({
      timestamp: Date.now(),
      type: 'paused',
      details: { paused: pause },
    });
    
    // Store pause state in metadata
    state.metadata.paused = pause;
    await this.setState(state);
  }

  async setLanguage(locale: string): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Room state not initialized');

    state.language = locale;
    
    await this.recordEvent({
      timestamp: Date.now(),
      type: 'language_changed',
      details: { locale },
    });
    
    await this.setState(state);
  }

  async getHistory(limit = 50): Promise<RoomEvent[]> {
    const events = await this.ctx.storage.get<RoomEvent[]>('room-events') || [];
    return events.slice(-limit);
  }

  private async recordEvent(event: RoomEvent): Promise<void> {
    const events = await this.ctx.storage.get<RoomEvent[]>('room-events') || [];
    events.push(event);
    if (events.length > 200) {
      events.splice(0, events.length - 200);
    }
    await this.ctx.storage.put('room-events', events);
  }
}

interface Env {
  BOT_STATE: DurableObjectNamespace;
  ROOM_STATE: DurableObjectNamespace;
  BOT_CONFIG: KVNamespace;
  DB: D1Database;
}