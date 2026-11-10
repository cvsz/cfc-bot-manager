import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';
import { useOrganization } from '@/context/OrganizationContext';

interface WebSocketMessage {
  type: string;
  bot_id?: string;
  room_id?: string;
  payload?: unknown;
  error?: string;
  timestamp: string;
}

interface BotStatusPayload {
  uptime: string;
  reason: string;
  status: string;
  raw?: unknown;
}

interface UseBotWebSocketsOptions {
  botIds: string[];
  onStatus?: (botId: string, status: BotStatusPayload) => void;
  onError?: (error: string) => void;
  reconnect?: boolean;
  reconnectInterval?: number;
}

interface BotConnectionState {
  isConnected: boolean;
  connectionError: string | null;
  lastStatus: BotStatusPayload | null;
  connect: () => void;
  disconnect: () => void;
}

export function useBotWebSockets({
  botIds,
  onStatus,
  onError,
  reconnect = true,
  reconnectInterval = 5000,
}: UseBotWebSocketsOptions) {
  const { isAuthenticated } = useAuth();
  const { currentOrganization } = useOrganization();
  
  const [connections, setConnections] = useState<Map<string, BotConnectionState>>(new Map());
  const wsRefs = useRef<Map<string, WebSocket>>(new Map());
  const reconnectTimeoutRefs = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());
  const reconnectAttemptsRefs = useRef<Map<string, number>>(new Map());

  const getWsUrl = useCallback((botId: string) => {
    const orgId = currentOrganization?.id;
    if (!orgId) return null;
    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/api/v1/ws?bot_id=${botId}&org_id=${orgId}`;
    return wsUrl;
  }, [currentOrganization]);

  const getAccessToken = useCallback(() => {
    return localStorage.getItem('access_token');
  }, []);

  const connect = useCallback((botId: string) => {
    if (!isAuthenticated) return;

    const token = getAccessToken();
    if (!token) return;

    const wsUrl = getWsUrl(botId);
    if (!wsUrl) return;

    const ws = new WebSocket(wsUrl);
    wsRefs.current.set(botId, ws);

    ws.onopen = () => {
      console.log('WebSocket connected');
      setConnections(prev => {
        const newMap = new Map(prev);
        const existing = newMap.get(botId) || { isConnected: false, connectionError: null, lastStatus: null, connect: () => {}, disconnect: () => {} };
        newMap.set(botId, { ...existing, isConnected: true, connectionError: null });
        return newMap;
      });
      reconnectAttemptsRefs.current.set(botId, 0);
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        
        if (message.type === 'status' && message.payload) {
          const status = message.payload as BotStatusPayload;
          
          setConnections(prev => {
            const newMap = new Map(prev);
            const existing = newMap.get(botId) || { isConnected: false, connectionError: null, lastStatus: null, connect: () => {}, disconnect: () => {} };
            newMap.set(botId, { ...existing, lastStatus: status });
            return newMap;
          });
          
          if (onStatus) {
            onStatus(botId, status);
          }
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err);
      }
    };

    ws.onclose = (event) => {
      console.log('WebSocket closed:', event.code, event.reason);
      setConnections(prev => {
        const newMap = new Map(prev);
        const existing = newMap.get(botId);
        if (existing) {
          newMap.set(botId, { ...existing, isConnected: false });
        }
        return newMap;
      });
      
      // Attempt reconnect
      if (reconnect) {
        const attempts = (reconnectAttemptsRefs.current.get(botId) || 0) + 1;
        reconnectAttemptsRefs.current.set(botId, attempts);
        
        if (attempts <= 10) {
          setTimeout(() => connect(botId), reconnectInterval * Math.min(attempts, 5));
        } else if (onError) {
          onError(`Max reconnection attempts reached for bot ${botId}`);
        }
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      if (onError) onError(`WebSocket connection error for bot ${botId}`);
    };
  }, [isAuthenticated, getAccessToken, getWsUrl, onStatus, onError, reconnect, reconnectInterval]);

  const disconnect = useCallback((botId: string) => {
    const timeout = reconnectTimeoutRefs.current.get(botId);
    if (timeout) {
      clearTimeout(timeout);
      reconnectTimeoutRefs.current.delete(botId);
    }
    
    const ws = wsRefs.current.get(botId);
    if (ws) {
      ws.close();
      wsRefs.current.delete(botId);
    }
    
    setConnections(prev => {
      const newMap = new Map(prev);
      const existing = newMap.get(botId);
      if (existing) {
        newMap.set(botId, { ...existing, isConnected: false });
      }
      return newMap;
    });
  }, []);

  // Connect on mount and when botIds change
  useEffect(() => {
    if (!isAuthenticated) return;
    
    botIds.forEach(botId => connect(botId));
    
    return () => {
      botIds.forEach(botId => disconnect(botId));
    };
  }, [botIds, isAuthenticated, connect, disconnect]);

  // Send ping periodically to keep connections alive
  useEffect(() => {
    const interval = setInterval(() => {
      wsRefs.current.forEach((ws) => {
        if (ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({ type: 'ping' }));
        }
      });
    }, 30000);

    return () => clearInterval(interval);
  }, []);

  // Get connection state for a specific bot
  const getConnectionState = useCallback((botId: string): BotConnectionState => {
    return connections.get(botId) || {
      isConnected: false,
      connectionError: null,
      lastStatus: null,
      connect: () => connect(botId),
      disconnect: () => disconnect(botId),
    };
  }, [connections, connect, disconnect]);

  return {
    connections,
    getConnectionState,
    connect,
    disconnect,
  };
}