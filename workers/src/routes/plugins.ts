import { Env } from "../index";

export async function handlePluginRoutes(request: Request, env: Env, path: string): Promise<Response> {
  const corsHeaders = getCorsHeaders();
  const url = new URL(request.url);
  const parts = path.split('/').filter(Boolean);
  
  // For plugins, we use the bot's DO since plugin configs are per-bot/room
  const pluginId = url.searchParams.get('plugin_id');
  const roomId = url.searchParams.get('room_id');
  const botId = url.searchParams.get('bot_id');
  const orgId = url.searchParams.get('org_id');
  
  if (!orgId) {
    return jsonResponse({ error: 'org_id query parameter required' }, 400);
  }

  // If we have a bot_id, use the bot's DO; otherwise use room DO
  let stub;
  if (botId) {
    const doId = env.BOT_STATE.idFromName(`${orgId}:${botId}`);
    stub = env.BOT_STATE.get(`${orgId}:${botId}`);
  } else if (roomId) {
    const doId = env.ROOM_STATE.idFromName(`${orgId}:${roomId}`);
    stub = env.ROOM_STATE.get(doId);
  } else {
    return jsonResponse({ error: 'bot_id or room_id required' }, 400);
  }

  const internalUrl = `http://internal/plugins${url.pathname.replace('/api/v1/plugins', '')}${url.search}`;
  
  const internalRequest = new Request(internalUrl, {
    method: request.method,
    headers: request.headers,
    body: request.body,
  });

  return stub.fetch(internalRequest);
}

function getCorsHeaders(): Record<string, string> {
  return {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
    'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    'Access-Control-Max-Age': '86400',
  };
}

function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      'Content-Type': 'application/json',
      ...getCorsHeaders(),
    },
  });
}