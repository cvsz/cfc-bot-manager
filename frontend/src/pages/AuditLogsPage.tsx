import { useEffect, useState } from 'react';
import { Search, FileText, Clock, User, Download, ChevronLeft, ChevronRight } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardContent } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { formatRelativeTime, truncate } from '@/utils/helpers';
import api from '@/api/client';
import type { AuditLog } from '@/types/api';

export function AuditLogsPage() {
  const { currentOrganization } = useOrganization();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  
  const [searchQuery, setSearchQuery] = useState('');
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [resourceFilter, setResourceFilter] = useState<string>('all');
  const [pagination, setPagination] = useState({ limit: 50, offset: 0, total: 0 });
  const [availableActions, setAvailableActions] = useState<string[]>([]);
  const [availableResources, setAvailableResources] = useState<string[]>([]);

  useEffect(() => {
    if (!currentOrganization) return;
    fetchLogs();
  }, [currentOrganization, pagination.offset, searchQuery, actionFilter, resourceFilter]);

  const fetchLogs = async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    console.warn(null);
    try {
      const res = await api.getAuditLogs(currentOrganization.id, {
        limit: pagination.limit,
        offset: pagination.offset,
        filter: {
          ...(searchQuery && { search: searchQuery }),
          ...(actionFilter !== 'all' && { action: actionFilter }),
          ...(resourceFilter !== 'all' && { resource_type: resourceFilter }),
        },
      });
      setLogs(res.items as AuditLog[]);
      setPagination(prev => ({ ...prev, total: res.total_count }));

      // Extract unique actions and resources for filters
      const allRes = await api.getAuditLogs(currentOrganization.id, { limit: 1000 });
      const actions = [...new Set((allRes.items as AuditLog[]).map(l => l.action))].sort();
      const resources = [...new Set((allRes.items as AuditLog[]).map(l => l.resource_type))].sort();
      setAvailableActions(actions);
      setAvailableResources(resources);
    } catch (err) {
      console.error('Failed to fetch audit logs:', err);
      console.warn('Failed to load audit logs');
    } finally {
      setIsLoading(false);
    }
  };

  const getActionColor = (action: string) => {
    if (action.includes('create')) return 'success';
    if (action.includes('delete') || action.includes('remove')) return 'error';
    if (action.includes('update') || action.includes('edit') || action.includes('modify')) return 'warning';
    if (action.includes('enable') || action.includes('start')) return 'success';
    if (action.includes('disable') || action.includes('stop')) return 'error';
    return 'default';
  };

  return (
    <LoadingOverlay isLoading={isLoading && logs.length === 0}>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Audit Logs</h1>
            <p className="text-dark-600 dark:text-dark-400 mt-1">Track all privileged operations across your organization</p>
          </div>
          <Button variant="secondary">
            <Download className="h-4 w-4 mr-2" /> Export
          </Button>
        </div>

        <Card className="overflow-hidden">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-dark-400" />
                <input
                  type="text"
                  placeholder="Search logs..."
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setPagination(prev => ({ ...prev, offset: 0 }));
                  }}
                  className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 bg-white text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                />
              </div>
              <Select
                value={actionFilter}
                onChange={(e) => { setActionFilter(e.target.value); setPagination(prev => ({ ...prev, offset: 0 })); }}
                options={[{ value: 'all', label: 'All Actions' }, ...availableActions.map(a => ({ value: a, label: a.replace(/_/g, ' ') }))]}
                className="w-full sm:w-48"
              />
              <Select
                value={resourceFilter}
                onChange={(e) => { setResourceFilter(e.target.value); setPagination(prev => ({ ...prev, offset: 0 })); }}
                options={[{ value: 'all', label: 'All Resources' }, ...availableResources.map(r => ({ value: r, label: r }))]}
                className="w-full sm:w-48"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-0">
            {logs.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-dark-200 dark:border-dark-700 bg-dark-50 dark:bg-dark-900/50">
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Time</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Action</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Resource</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Actor</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Details</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-dark-200 dark:divide-dark-700">
                    {logs.map((log) => (
                      <tr key={log.id} className="hover:bg-dark-50 dark:hover:bg-dark-800/50">
                        <td className="px-4 py-3 whitespace-nowrap">
                          <div className="flex items-center gap-2">
                            <Clock className="h-4 w-4 text-dark-400" />
                            <span className="text-sm text-dark-600 dark:text-dark-400">{formatRelativeTime(log.created_at)}</span>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <Badge variant={getActionColor(log.action) as any}>
                            {log.action.replace(/_/g, ' ')}
                          </Badge>
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400">
                          {log.resource_type}: <span className="font-mono text-xs">{truncate(log.resource_id, 12)}</span>
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <User className="h-4 w-4 text-dark-400" />
                            <span className="text-sm text-dark-600 dark:text-dark-400">
                              {log.actor_type === 'user' ? 'User' : log.actor_type === 'system' ? 'System' : 'API Key'}
                              {log.actor_user_id && <span className="font-mono text-xs text-dark-400">({truncate(log.actor_user_id, 8)})</span>}
                            </span>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-500 dark:text-dark-400 max-w-md truncate">
                          {log.new_value ? `New: ${truncate(log.new_value, 60)}` : log.old_value ? `Old: ${truncate(log.old_value, 60)}` : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-12">
                <FileText className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                <p className="text-dark-500 dark:text-dark-400">No audit logs found</p>
              </div>
            )}

            {pagination.total > pagination.limit && (
              <div className="px-4 py-3 border-t border-dark-200 dark:border-dark-700 flex items-center justify-between">
                <p className="text-sm text-dark-600 dark:text-dark-400">
                  Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, pagination.total)} of {pagination.total}
                </p>
                <div className="flex gap-2">
                  <Button variant="secondary" size="sm" disabled={pagination.offset === 0} onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}>
                    <ChevronLeft className="h-4 w-4 mr-1" /> Previous
                  </Button>
                  <Button variant="secondary" size="sm" disabled={pagination.offset + pagination.limit >= pagination.total} onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}>
                    Next <ChevronRight className="h-4 w-4 ml-1" />
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </LoadingOverlay>
  );
}