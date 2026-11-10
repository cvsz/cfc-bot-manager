import { useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from '@/context/AuthContext';

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

interface UseWebSocketOptions {
  botId?: string;
  onStatus?: (status: BotStatusPayload) => void;
  onEvent?: (eventType: string, payload: unknown) => void;
  onError?: (error: string) => void;
  reconnect?: boolean;
  reconnectInterval?: number;
}

export function useWebSocket({
  botId,
  onStatus,
  onEvent,
  onError,
  reconnect = true,
  reconnectInterval = 5000,
}: UseWebSocketOptions) {
  const { isAuthenticated, getAccessToken } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 10;

  const connect = useCallback(() => {
    if (!isAuthenticated || !botId) return;

    const token = getAccessToken();
    if (!token) return;

    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/api/v1/ws?bot_id=${botId}&org_id=current`;
    
    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setConnectionError(null);
        reconnectAttemptsRef.current = 0;
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          
          switch (message.type) {
            case 'status':
              if (onStatus && message.payload) {
                onStatus(message.payload as BotStatusPayload);
              }
              break;
            case 'ping':
              // Send pong
              ws.send(JSON.stringify({ type: 'pong' }));
              break;
            default:
              if (onEvent) {
                onEvent(message.type, message.payload);
              }
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        setIsConnected(false);
        
        if (reconnect && reconnectAttemptsRef.current < maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectInterval * reconnectAttemptsRef.current);
        } else if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
          setConnectionError('Max reconnection attempts reached');
          if (onError) onError('Max reconnection attempts reached');
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionError('WebSocket connection error');
        if (onError) onError('WebSocket connection error');
      };
    } catch (err) {
      console.error('Failed to create WebSocket:', err);
      setConnectionError('Failed to create WebSocket connection');
    }
  }, [isAuthenticated, botId, getAccessToken, onStatus, onEvent, onError, reconnect, reconnectInterval]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((message: object) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
    }
  }, []);

  // Send ping periodically to keep connection alive
  useEffect(() => {
    if (!isConnected) return;
    
    const pingInterval = setInterval(() => {
      sendMessage({ type: 'ping' });
    }, 30000);

    return () => clearInterval(pingInterval);
  }, [isConnected, sendMessage]);

  // Connect on mount and when botId changes
  useEffect(() => {
    if (botId) {
      connect();
    }
    
    return () => {
      disconnect();
    };
  }, [botId, connect, disconnect]);

  return {
    isConnected,
    connectionError,
    connect,
    disconnect,
    sendMessage,
  };
}

export function useBotStatus(botId: string | undefined) {
  const [status, setStatus] = useState<BotStatusPayload | null>(null);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);

  const { isConnected, connectionError, connect, disconnect } = useWebSocket({
    botId,
    onStatus: (status) => {
      setStatus(status);
      setLastUpdate(new Date());
    },
    onEvent: (type, payload) => {
      console.log('Bot event:', type, payload);
    },
    onError: (error) => {
      console.error('Bot WebSocket error:', error);
    },
  });

  return {
    status,
    lastUpdate,
    isConnected,
    connectionError,
    connect,
    disconnect,
  };
}