import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Bot, Cpu, FileText, ArrowUpRight, Activity, Users, Clock } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/Card';
import { StatusBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { cn, formatRelativeTime } from '@/utils/helpers';
import api from '@/api/client';
import type { AuditLog } from '@/types/api';

interface DashboardStats {
  rooms: { total: number; connected: number; disconnected: number };
  bots: { total: number; running: number; stopped: number; error: number };
  plugins: { total: number; enabled: number };
  recentActivity: AuditLog[];
}

export function DashboardPage() {
  const { currentOrganization } = useOrganization();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!currentOrganization) return;

    const fetchDashboardData = async () => {
      setIsLoading(true);
      setError(null);

      try {
        const [roomsRes, botsRes, pluginsRes, auditRes] = await Promise.all([
          api.getRooms(currentOrganization.id, { limit: 100 }),
          api.getBots(currentOrganization.id, { limit: 100 }),
          api.getPlugins({ limit: 100 }),
          api.getAuditLogs(currentOrganization.id, { limit: 10 }),
        ]);

        const rooms = roomsRes.items as import('@/types/api').Room[];
        const bots = botsRes.items as import('@/types/api').Bot[];
        const plugins = pluginsRes.items as import('@/types/api').Plugin[];
        const recentActivity = auditRes.items as import('@/types/api').AuditLog[];

        setStats({
          rooms: {
            total: rooms.length,
            connected: rooms.filter(r => r.status === 'connected').length,
            disconnected: rooms.filter(r => r.status === 'disconnected').length,
          },
          bots: {
            total: bots.length,
            running: bots.filter(b => b.status === 'running').length,
            stopped: bots.filter(b => b.status === 'stopped').length,
            error: bots.filter(b => b.status === 'error').length,
          },
          plugins: {
            total: plugins.length,
            enabled: plugins.filter(p => p.is_enabled).length,
          },
          recentActivity,
        });
      } catch (err) {
        console.error('Failed to fetch dashboard data:', err);
        setError('Failed to load dashboard data');
      } finally {
        setIsLoading(false);
      }
    };

    fetchDashboardData();
  }, [currentOrganization]);

  if (isLoading && !stats) {
    return (
      <div className="space-y-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map(i => (
            <Card key={i} className="animate-pulse">
              <CardContent className="space-y-4">
                <div className="h-4 w-1/4 bg-dark-200 dark:bg-dark-700 rounded" />
                <div className="h-8 w-1/2 bg-dark-200 dark:bg-dark-700 rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error && !stats) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600 dark:text-red-400">{error}</p>
        <Button variant="secondary" className="mt-4" onClick={() => window.location.reload()}>
          Retry
        </Button>
      </div>
    );
  }

  const statCards = [
    {
      title: 'Rooms',
      value: stats?.rooms.total || 0,
      change: stats ? `${stats.rooms.connected} connected` : '—',
      icon: Bot,
      color: 'text-primary-600 bg-primary-100 dark:bg-primary-900/30 dark:text-primary-400',
      href: '/rooms',
    },
    {
      title: 'Bots',
      value: stats?.bots.total || 0,
      change: stats ? `${stats.bots.running} running` : '—',
      icon: Cpu,
      color: 'text-green-600 bg-green-100 dark:bg-green-900/30 dark:text-green-400',
      href: '/bots',
    },
    {
      title: 'Plugins',
      value: stats?.plugins.total || 0,
      change: stats ? `${stats.plugins.enabled} enabled` : '—',
      icon: FileText,
      color: 'text-purple-600 bg-purple-100 dark:bg-purple-900/30 dark:text-purple-400',
      href: '/plugins',
    },
    {
      title: 'Activity (24h)',
      value: stats?.recentActivity.length || 0,
      change: 'Recent actions',
      icon: Activity,
      color: 'text-orange-600 bg-orange-100 dark:bg-orange-900/30 dark:text-orange-400',
      href: '/audit-logs',
    },
  ];

  return (
    <LoadingOverlay isLoading={isLoading && stats !== null}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Dashboard</h1>
            <p className="text-dark-600 dark:text-dark-400 mt-1">
              Overview of your {currentOrganization?.name} organization
            </p>
          </div>
          <div className="flex gap-2">
            <Link to="/rooms">
              <Button variant="secondary">
                <Bot className="h-4 w-4 mr-2" /> Manage Rooms
              </Button>
            </Link>
            <Link to="/bots">
              <Button>
                <Cpu className="h-4 w-4 mr-2" /> Manage Bots
              </Button>
            </Link>
          </div>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {statCards.map((stat) => (
            <Link key={stat.title} to={stat.href} className="block">
              <Card>
                <CardContent className="flex items-center gap-4">
                  <div className={cn('p-3 rounded-xl', stat.color)}>
                    <stat.icon className="h-6 w-6" aria-hidden="true" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-dark-600 dark:text-dark-400 truncate">{stat.title}</p>
                    <p className="text-2xl font-bold text-dark-900 dark:text-dark-100">{stat.value}</p>
                    <p className="text-xs text-dark-500 dark:text-dark-400">{stat.change}</p>
                  </div>
                  <ArrowUpRight className="h-5 w-5 text-dark-400" aria-hidden="true" />
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>

        {/* Detailed Status */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Rooms Status */}
          <Card>
            <CardHeader>
              <CardTitle>Room Status</CardTitle>
              <CardDescription>Current connection status of all rooms</CardDescription>
            </CardHeader>
            <CardContent>
              {stats && stats.rooms.total > 0 ? (
                <div className="space-y-4">
                  <div className="flex items-center gap-4">
                    <StatusBadge status="connected" />
                    <span className="text-sm text-dark-600 dark:text-dark-400">
                      {stats.rooms.connected} of {stats.rooms.total} connected
                    </span>
                  </div>
                  <div className="h-2 bg-dark-200 dark:bg-dark-700 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-green-500 transition-all duration-300"
                      style={{ width: `${stats.rooms.total > 0 ? (stats.rooms.connected / stats.rooms.total) * 100 : 0}%` }}
                    />
                  </div>
                  
                  <div className="flex items-center gap-4 mt-4">
                    <StatusBadge status="disconnected" />
                    <span className="text-sm text-dark-600 dark:text-dark-400">
                      {stats.rooms.disconnected} disconnected
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-center py-8">
                  <Bot className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                  <p className="text-dark-500 dark:text-dark-400">No rooms configured</p>
                  <Link to="/rooms">
                    <Button variant="secondary" className="mt-4">Create your first room</Button>
                  </Link>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Bots Status */}
          <Card>
            <CardHeader>
              <CardTitle>Bot Status</CardTitle>
              <CardDescription>Current state of all bots</CardDescription>
            </CardHeader>
            <CardContent>
              {stats && stats.bots.total > 0 ? (
                <div className="space-y-3">
                  {[
                    { status: 'running', count: stats.bots.running, label: 'Running' },
                    { status: 'stopped', count: stats.bots.stopped, label: 'Stopped' },
                    { status: 'error', count: stats.bots.error, label: 'Error' },
                    { status: 'starting', count: 0, label: 'Starting' },
                  ].map(({ status, count, label }) => (
                    count > 0 && (
                      <div key={status} className="flex items-center gap-3">
                        <StatusBadge status={status} />
                        <div className="flex-1">
                          <p className="text-sm font-medium text-dark-900 dark:text-dark-100">{label}</p>
                          <div className="h-1.5 bg-dark-200 dark:bg-dark-700 rounded-full overflow-hidden">
                            <div
                              className="h-full transition-all duration-300"
                              style={{ 
                                width: `${(count / stats.bots.total) * 100}%`,
                                backgroundColor: status === 'running' ? '#22c55e' : 
                                  status === 'error' ? '#ef4444' : '#64748b'
                              }}
                            />
                          </div>
                        </div>
                        <span className="text-sm text-dark-500 dark:text-dark-400 w-10 text-right">{count}</span>
                      </div>
                    )
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Cpu className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                  <p className="text-dark-500 dark:text-dark-400">No bots configured</p>
                  <Link to="/bots">
                    <Button variant="secondary" className="mt-4">Create your first bot</Button>
                  </Link>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Recent Activity</CardTitle>
                <CardDescription>Latest actions across your organization</CardDescription>
              </div>
              <Link to="/audit-logs">
                <Button variant="secondary">View all</Button>
              </Link>
            </div>
          </CardHeader>
          <CardContent>
            {stats && stats.recentActivity.length > 0 ? (
              <div className="space-y-4">
                {stats.recentActivity.slice(0, 10).map((activity) => (
                  <div key={activity.id} className="flex items-center gap-4 py-3 border-b border-dark-200 dark:border-dark-700 last:border-0">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-dark-100 dark:bg-dark-800">
                      <Activity className="h-5 w-5 text-dark-600 dark:text-dark-400" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-dark-900 dark:text-dark-100 truncate">
                        {activity.action.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                      </p>
                      <p className="text-xs text-dark-500 dark:text-dark-400">
                        {activity.resource_type}: {activity.resource_id}
                      </p>
                    </div>
                    <div className="flex items-center gap-3 text-sm text-dark-500 dark:text-dark-400">
                      {activity.actor_type === 'user' && (
                        <>
                          <Users className="h-4 w-4" />
                          <span>User</span>
                        </>
                      )}
                      <Clock className="h-4 w-4" />
                      <span>{formatRelativeTime(activity.created_at)}</span>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <Activity className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                <p className="text-dark-500 dark:text-dark-400">No recent activity</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </LoadingOverlay>
  );
}