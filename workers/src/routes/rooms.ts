import { Env } from "../index";

export async function handleRoomRoutes(request: Request, env: Env, path: string): Promise<Response> {
  const corsHeaders = getCorsHeaders();
  const url = new URL(request.url);
  const parts = path.split('/').filter(Boolean);
  
  if (parts.length < 2) {
    return jsonResponse({ error: 'Room ID required' }, 400);
  }

  const roomId = parts[1];
  const orgId = url.searchParams.get('org_id');
  
  if (!orgId) {
    return jsonResponse({ error: 'org_id query parameter required' }, 400);
  }

  const doId = env.ROOM_STATE.idFromName(`${orgId}:${roomId}`);
  const stub = env.ROOM_STATE.get(doId);

  const subPath = parts.slice(2).join('/');
  const internalUrl = `http://internal/${subPath}`;

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