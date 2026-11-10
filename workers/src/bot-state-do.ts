import { DurableObject } from "cloudflare:workers";

export interface BotState {
  id: string;
  organizationId: string;
  roomId?: string;
  name: string;
  username: string;
  status: BotStatus;
  config: BotConfig;
  lastStatusUpdate: number;
  connectionHistory: ConnectionEvent[];
  pluginConfigs: Map<string, PluginConfig>;
  metadata: Record<string, unknown>;
}

export type BotStatus = 'stopped' | 'starting' | 'running' | 'stopping' | 'error' | 'reconnecting';

export interface BotConfig {
  autoReconnect: boolean;
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  proxyUrl?: string;
  camfrogBaseUrl: string;
  retryAttempts: number;
  retryDelay: number;
}

export interface ConnectionEvent {
  timestamp: number;
  type: 'connected' | 'disconnected' | 'error' | 'reconnecting' | 'status_change';
  details: Record<string, unknown>;
  previousStatus?: BotStatus;
  newStatus?: BotStatus;
}

export interface PluginConfig {
  pluginId: string;
  roomId?: string;
  settings: Record<string, unknown>;
  isEnabled: boolean;
  version: number;
}

export interface BotStateDO extends DurableObject<Env> {
  // State management
  getState(): Promise<BotState | null>;
  setState(state: BotState): Promise<void>;
  
  // Status management
  updateStatus(status: BotStatus, details?: Record<string, unknown>): Promise<BotState>;
  getStatus(): Promise<BotStatus>;
  
  // Configuration
  updateConfig(config: Partial<BotConfig>): Promise<BotConfig>;
  getConfig(): Promise<BotConfig>;
  
  // Room assignment
  assignRoom(roomId: string): Promise<void>;
  unassignRoom(): Promise<void>;
  
  // Plugin configuration
  setPluginConfig(pluginId: string, roomId: string | undefined, config: PluginConfig): Promise<void>;
  getPluginConfig(pluginId: string, roomId?: string): Promise<PluginConfig | null>;
  removePluginConfig(pluginId: string, roomId?: string): Promise<void>;
  
  // Connection events
  recordConnectionEvent(event: ConnectionEvent): Promise<void>;
  getConnectionHistory(limit?: number): Promise<ConnectionEvent[]>;
  
  // WebSocket management
  addWebSocketClient(clientId: string): Promise<void>;
  removeWebSocketClient(clientId: string): Promise<void>;
  broadcastToClients(message: unknown): Promise<void>;
  
  // Health and monitoring
  healthCheck(): Promise<{ healthy: boolean; details: Record<string, unknown> }>;
}

export class BotStateDOImpl implements BotStateDO {
  private state: BotState;
  private websocketClients: Map<string, WebSocket> = new Map();

  constructor(ctx: DurableObjectState, env: Env) {
    this.state = {
      id: '',
      organizationId: '',
      name: '',
      username: '',
      status: 'stopped',
      config: {
        autoReconnect: true,
        logLevel: 'info',
        camfrogBaseUrl: '',
        retryAttempts: 3,
        retryDelay: 1000,
      },
      lastStatusUpdate: Date.now(),
      connectionHistory: [],
      pluginConfigs: new Map(),
      metadata: {},
    };
  }

  async getState(): Promise<BotState | null> {
    const stored = await this.ctx.storage.get<BotState>('bot-state');
    return stored || this.state;
  }

  async setState(state: BotState): Promise<void> {
    this.state = state;
    await this.ctx.storage.put('bot-state', state);
  }

  async updateStatus(status: BotStatus, details: Record<string, unknown> = {}): Promise<BotState> {
    const currentState = await this.getState();
    if (!currentState) throw new Error('Bot state not initialized');

    const previousStatus = currentState.status;
    const event: ConnectionEvent = {
      timestamp: Date.now(),
      type: 'status_change',
      details,
      previousStatus,
      newStatus: status,
    };

    currentState.status = status;
    currentState.lastStatusUpdate = Date.now();
    currentState.connectionHistory.push(event);
    
    // Keep only last 100 events
    if (currentState.connectionHistory.length > 100) {
      currentState.connectionHistory = currentState.connectionHistory.slice(-100);
    }

    await this.setState(currentState);
    await this.broadcastToClients({ type: 'status_update', payload: { status, ...details } });
    
    return currentState;
  }

  async getStatus(): Promise<BotStatus> {
    const state = await this.getState();
    return state?.status || 'stopped';
  }

  async updateConfig(config: Partial<BotConfig>): Promise<BotConfig> {
    const state = await this.getState();
    if (!state) throw new Error('Bot state not initialized');

    state.config = { ...state.config, ...config };
    await this.setState(state);
    return state.config;
  }

  async getConfig(): Promise<BotConfig> {
    const state = await this.getState();
    return state?.config || this.state.config;
  }

  async assignRoom(roomId: string): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Bot state not initialized');
    
    state.roomId = roomId;
    await this.setState(state);
  }

  async unassignRoom(): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Bot state not initialized');
    
    state.roomId = undefined;
    await this.setState(state);
  }

  async setPluginConfig(pluginId: string, roomId: string | undefined, config: PluginConfig): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Bot state not initialized');

    const key = roomId ? `${pluginId}:${roomId}` : pluginId;
    state.pluginConfigs.set(key, config);
    await this.setState(state);
  }

  async getPluginConfig(pluginId: string, roomId?: string): Promise<PluginConfig | null> {
    const state = await this.getState();
    if (!state) return null;

    const key = roomId ? `${pluginId}:${roomId}` : pluginId;
    return state.pluginConfigs.get(key) || null;
  }

  async removePluginConfig(pluginId: string, roomId?: string): Promise<void> {
    const state = await this.getState();
    if (!state) return;

    const key = roomId ? `${pluginId}:${roomId}` : pluginId;
    state.pluginConfigs.delete(key);
    await this.setState(state);
  }

  async recordConnectionEvent(event: ConnectionEvent): Promise<void> {
    const state = await this.getState();
    if (!state) throw new Error('Bot state not initialized');

    state.connectionHistory.push(event);
    if (state.connectionHistory.length > 100) {
      state.connectionHistory = state.connectionHistory.slice(-100);
    }
    await this.setState(state);
  }

  async getConnectionHistory(limit = 50): Promise<ConnectionEvent[]> {
    const state = await this.getState();
    if (!state) return [];
    return state.connectionHistory.slice(-limit);
  }

  async addWebSocketClient(clientId: string): Promise<void> {
    // In a real implementation, you'd store the WebSocket reference
    // For now, we just track the client ID
  }

  async removeWebSocketClient(clientId: string): Promise<void> {
    // Remove client tracking
  }

  async broadcastToClients(message: unknown): Promise<void> {
    const messageStr = JSON.stringify(message);
    for (const [clientId, ws] of this.websocketClients) {
      if (ws.readyState === WebSocket.OPEN) {
        try {
          ws.send(messageStr);
        } catch {
          this.websocketClients.delete(clientId);
        }
      }
    }
  }

  async healthCheck(): Promise<{ healthy: boolean; details: Record<string, unknown> }> {
    const state = await this.getState();
    if (!state) {
      return { healthy: false, details: { error: 'State not initialized' } };
    }

    return {
      healthy: true,
      details: {
        status: state.status,
        lastUpdate: state.lastStatusUpdate,
        uptime: Date.now() - (state.metadata.startTime as number || Date.now()),
        eventsCount: state.connectionHistory.length,
        pluginConfigsCount: state.pluginConfigs.size,
      },
    };
  }
}

// WebSocket handler for real-time updates
export class BotWebSocketHandler {
  private do: BotStateDOImpl;

  constructor(do: BotStateDOImpl) {
    this.do = do;
  }

  async handleWebSocket(ws: WebSocket, request: Request): Promise<void> {
    ws.accept();
    
    const clientId = crypto.randomUUID();
    
    ws.addEventListener('message', async (event) => {
      try {
        const message = JSON.parse(event.data as string);
        await this.handleMessage(message, clientId);
      } catch (err) {
        console.error('WebSocket message error:', err);
      }
    });

    ws.addEventListener('close', () => {
      this.do.removeWebSocketClient(clientId);
    });

    ws.addEventListener('error', (err) => {
      console.error('WebSocket error:', err);
      this.do.removeWebSocketClient(clientId);
    });
  }

  private async handleMessage(message: unknown, clientId: string): Promise<void> {
    const msg = message as { type: string; payload?: unknown };
    
    switch (msg.type) {
      case 'subscribe':
        // Client wants to subscribe to updates
        break;
      case 'ping':
        // Respond with pong
        break;
      case 'get_status':
        const status = await this.do.getStatus();
        // Send status back
        break;
    }
  }
}

interface Env {
  BOT_STATE: DurableObjectNamespace<BotStateDO>;
  ROOM_STATE: DurableObjectNamespace;
  BOT_CONFIG: KVNamespace;
  DB: D1Database;
}