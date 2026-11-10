import { BotStateDOImpl } from "./bot-state-do";
import { RoomStateDOImpl } from "./room-state-do";
import { handleBotRoutes } from "./routes/bots";
import { handleRoomRoutes } from "./routes/rooms";
import { handlePluginRoutes } from "./routes/plugins";

export interface Env {
  BOT_STATE: DurableObjectNamespace;
  ROOM_STATE: DurableObjectNamespace;
  BOT_CONFIG: KVNamespace;
  DB: D1Database;
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    
    // CORS headers
    const corsHeaders = {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400',
    };

    // Handle CORS preflight
    if (request.method === 'OPTIONS') {
      return new Response(null, { headers: corsHeaders });
    }

    // Health check
    if (url.pathname === '/healthz') {
      return new Response(JSON.stringify({ status: 'ok' }), {
        headers: { 'Content-Type': 'application/json', ...corsHeaders },
      });
    }

    // Ready check
    if (url.pathname === '/readyz') {
      return new Response(JSON.stringify({ status: 'ready' }), {
        headers: { 'Content-Type': 'application/json', ...corsHeaders },
      });
    }

    // API routes
    if (url.pathname.startsWith('/api/v1/')) {
      const path = url.pathname.slice('/api/v1'.length);
      
      if (path.startsWith('/bots')) {
        return handleBotRoutes(request, env, path);
      }
      
      if (path.startsWith('/rooms')) {
        return handleRoomRoutes(request, env, path);
      }
      
      if (path.startsWith('/plugins')) {
        return handlePluginRoutes(request, env, path);
      }
      
      // WebSocket endpoint
      if (path === '/ws') {
        return handleWebSocket(request, env);
      }
    }

    // Durable Object endpoints
    if (url.pathname.startsWith('/do/')) {
      return handleDurableObject(request, env);
    }

    return new Response('Not Found', { status: 404, headers: corsHeaders });
  },
} satisfies ExportedHandler<Env>;

async function handleWebSocket(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const botId = url.searchParams.get('bot_id');
  const orgId = url.searchParams.get('org_id');

  if (!botId || !orgId) {
    return new Response('bot_id and org_id required', { status: 400 });
  }

  // Get the Durable Object for this bot
  const id = env.BOT_STATE.idFromName(`${orgId}:${botId}`);
  const stub = env.BOT_STATE.get(id);

  // Forward the WebSocket request to the Durable Object
  return stub.fetch(request);
}

async function handleDurableObject(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const parts = url.pathname.split('/').filter(Boolean);
  
  if (parts.length < 3) {
    return new Response('Invalid DO path', { status: 400 });
  }

  const doType = parts[1]; // 'bot' or 'room'
  const doName = parts[2]; // orgId:botId or orgId:roomId
  const action = parts[3] || '';

  let stub: DurableObjectStub;
  
  if (doType === 'bot') {
    const id = env.BOT_STATE.idFromName(doName);
    stub = env.BOT_STATE.get(id);
  } else if (doType === 'room') {
    const id = env.ROOM_STATE.idFromName(doName);
    stub = env.ROOM_STATE.get(id);
  } else {
    return new Response('Invalid DO type', { status: 400 });
  }

  return stub.fetch(request);
}

// Helper to create bot state
async function createBotState(
  env: Env,
  orgId: string,
  botId: string,
  initialState: unknown
): Promise<void> {
  const id = env.BOT_STATE.idFromName(`${orgId}:${botId}`);
  const stub = env.BOT_STATE.get(id);
  
  await stub.fetch('http://internal/init', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(initialState),
  });
}

// Helper to get bot status
async function getBotStatus(env: Env, orgId: string, botId: string): Promise<unknown> {
  const id = env.BOT_STATE.idFromName(`${orgId}:${botId}`);
  const stub = env.BOT_STATE.get(id);
  
  const response = await stub.fetch('http://internal/status');
  return response.json();
}

// Helper to update bot status
async function updateBotStatus(
  env: Env,
  orgId: string,
  botId: string,
  status: string,
  details?: Record<string, unknown>
): Promise<void> {
  const id = env.BOT_STATE.idFromName(`${orgId}:${botId}`);
  const stub = env.BOT_STATE.get(id);
  
  await stub.fetch('http://internal/status', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status, details }),
  });
}

interface Env {
  BOT_STATE: DurableObjectNamespace;
  ROOM_STATE: DurableObjectNamespace;
  BOT_CONFIG: KVNamespace;
  DB: D1Database;
}