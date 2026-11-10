import { useEffect, useState, useCallback, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Search, Bot, Play, Pause, RotateCcw, Trash2, MoreVertical, Wifi } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardContent } from '@/components/ui/Card';
import { StatusBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { formatRelativeTime, truncate } from '@/utils/helpers';
import api from '@/api/client';
import type { Bot as BotType } from '@/types/api';
import { BotFormDialog } from '@/components/BotFormDialog';
import { useBotWebSockets } from '@/hooks/useBotWebSockets';

export function BotsPage() {
  const { currentOrganization } = useOrganization();
  const [bots, setBots] = useState<BotType[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [_error, setError] = useState<string | null>(null);
  
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [editingBot, setEditingBot] = useState<BotType | null>(null);
  const [pagination, setPagination] = useState({ limit: 20, offset: 0, total: 0 });

  // Extract bot IDs for WebSocket connections
  const botIds = useMemo(() => bots.map(b => b.id), [bots]);

  // Real-time WebSocket connections for bot status updates
  const { getConnectionState } = useBotWebSockets({
    botIds,
    onStatus: (botId, status) => {
      // Update bot status in real-time
      setBots(prev => prev.map(bot => 
        bot.id === botId ? { ...bot, status: status.status as import('@/types/api').BotStatus } : bot
      ));
    },
    onError: (error) => {
      console.error('WebSocket error:', error);
    },
    reconnect: true,
    reconnectInterval: 5000,
  });

  useEffect(() => {
    if (!currentOrganization) return;
    fetchBots();
  }, [currentOrganization, pagination.offset, searchQuery, statusFilter]);

  const fetchBots = async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.getBots(currentOrganization.id, {
        limit: pagination.limit,
        offset: pagination.offset,
        filter: {
          ...(searchQuery && { search: searchQuery }),
          ...(statusFilter !== 'all' && { status: statusFilter }),
        },
      });
      setBots(res.items);
      setPagination(prev => ({ ...prev, total: res.total_count }));
    } catch (err) {
      console.error('Failed to fetch bots:', err);
      setError('Failed to load bots');
    } finally {
      setIsLoading(false);
    }
  };

  const handleBotAction = async (botId: string, action: 'start' | 'stop' | 'restart') => {
    try {
      if (action === 'start') await api.startBot(currentOrganization!.id, botId);
      else if (action === 'stop') await api.stopBot(currentOrganization!.id, botId);
      else if (action === 'restart') await api.restartBot(currentOrganization!.id, botId);
      fetchBots();
    } catch (err) {
      console.error(`Failed to ${action} bot:`, err);
      alert(`Failed to ${action} bot`);
    }
  };

  const handleDeleteBot = async (botId: string) => {
    if (!confirm('Are you sure you want to delete this bot? This action cannot be undone.')) return;
    try {
      await api.deleteBot(currentOrganization!.id, botId);
      fetchBots();
    } catch (err) {
      console.error('Failed to delete bot:', err);
      alert('Failed to delete bot');
    }
  };

  const handleBotSaved = () => {
    setShowCreateDialog(false);
    setEditingBot(null);
    fetchBots();
  };

  // Helper to get real-time bot status
  const getBotDisplayStatus = useCallback((bot: BotType): string => {
    const connState = getConnectionState(bot.id);
    if (connState.lastStatus?.status) return connState.lastStatus.status;
    return bot.status;
  }, [getConnectionState]);

  const getBotConnectionStatus = useCallback((bot: BotType): boolean => {
    const connState = getConnectionState(bot.id);
    return connState.isConnected;
  }, [getConnectionState]);

  return (
    <LoadingOverlay isLoading={isLoading && bots.length === 0}>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Bots</h1>
            <p className="text-dark-600 dark:text-dark-400 mt-1">Manage your Camfrog bots</p>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" /> Add Bot
          </Button>
        </div>

        <Card className="overflow-hidden">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-dark-400" />
                <input
                  type="text"
                  placeholder="Search bots..."
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setPagination(prev => ({ ...prev, offset: 0 }));
                  }}
                  className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 bg-white text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                />
              </div>
              <Select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPagination(prev => ({ ...prev, offset: 0 }));
                }}
                options={[
                  { value: 'all', label: 'All Statuses' },
                  { value: 'running', label: 'Running' },
                  { value: 'stopped', label: 'Stopped' },
                  { value: 'starting', label: 'Starting' },
                  { value: 'stopping', label: 'Stopping' },
                  { value: 'error', label: 'Error' },
                ]}
                className="w-full sm:w-48"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-0">
            {bots.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-dark-200 dark:border-dark-700 bg-dark-50 dark:bg-dark-900/50">
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Bot</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Username</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Status</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Room</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Last Connected</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-dark-200 dark:divide-dark-700">
                    {bots.map((bot) => (
                      <tr key={bot.id} className="hover:bg-dark-50 dark:hover:bg-dark-800/50">
                        <td className="px-4 py-3">
                          <p className="font-medium text-dark-900 dark:text-dark-100">{bot.name}</p>
                          <p className="text-xs text-dark-500 dark:text-dark-400 font-mono">{truncate(bot.id, 8)}</p>
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400 font-mono">@{bot.username}</td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <StatusBadge status={getBotDisplayStatus(bot)} />
                            {getBotConnectionStatus(bot) && (
                              <span className="flex items-center gap-1 text-xs text-green-600 dark:text-green-400" title="Real-time connected">
                                <Wifi className="h-3 w-3" />
                                Live
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400">
                          {bot.room_id ? <span className="font-mono text-xs">{truncate(bot.room_id, 8)}</span> : <span className="text-dark-400 italic">Unassigned</span>}
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-500 dark:text-dark-400">
                          {bot.last_connected_at ? formatRelativeTime(bot.last_connected_at) : 'Never'}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            {getBotDisplayStatus(bot) === 'running' ? (
                              <>
                                <Button variant="ghost" size="sm" onClick={() => handleBotAction(bot.id, 'stop')} title="Stop">
                                  <Pause className="h-4 w-4" />
                                </Button>
                                <Button variant="ghost" size="sm" onClick={() => handleBotAction(bot.id, 'restart')} title="Restart">
                                  <RotateCcw className="h-4 w-4" />
                                </Button>
                              </>
                            ) : getBotDisplayStatus(bot) === 'stopped' ? (
                              <Button variant="ghost" size="sm" onClick={() => handleBotAction(bot.id, 'start')} title="Start">
                                <Play className="h-4 w-4" />
                              </Button>
                            ) : (
                              <Button variant="ghost" size="sm" onClick={() => handleBotAction(bot.id, 'restart')} title="Restart">
                                <RotateCcw className="h-4 w-4" />
                              </Button>
                            )}
                            <Link to={`/bots/${bot.id}`}>
                              <Button variant="ghost" size="sm">
                                <MoreVertical className="h-4 w-4" />
                              </Button>
                            </Link>
                            <Button variant="ghost" size="sm" onClick={() => handleDeleteBot(bot.id)}>
                              <Trash2 className="h-4 w-4 text-red-600" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-12">
                <Bot className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                <p className="text-dark-500 dark:text-dark-400">No bots found</p>
                <Button variant="secondary" className="mt-4" onClick={() => setShowCreateDialog(true)}>
                  Create your first bot
                </Button>
              </div>
            )}

            {pagination.total > pagination.limit && (
              <div className="px-4 py-3 border-t border-dark-200 dark:border-dark-700 flex items-center justify-between">
                <p className="text-sm text-dark-600 dark:text-dark-400">
                  Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, pagination.total)} of {pagination.total}
                </p>
                <div className="flex gap-2">
                  <Button variant="secondary" size="sm" disabled={pagination.offset === 0} onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}>Previous</Button>
                  <Button variant="secondary" size="sm" disabled={pagination.offset + pagination.limit >= pagination.total} onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}>Next</Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <BotFormDialog
        isOpen={showCreateDialog || !!editingBot}
        onClose={() => { setShowCreateDialog(false); setEditingBot(null); }}
        onSuccess={handleBotSaved}
        initialData={editingBot}
        organizationId={currentOrganization?.id}
      />
    </LoadingOverlay>
  );
}