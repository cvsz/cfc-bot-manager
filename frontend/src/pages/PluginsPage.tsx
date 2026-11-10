import { useEffect, useState } from 'react';
import { Search, Puzzle, CheckCircle, AlertCircle, HelpCircle, XCircle, ExternalLink } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardContent } from '@/components/ui/Card';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { truncate } from '@/utils/helpers';
import api from '@/api/client';
import type { Plugin, PluginConfiguration } from '@/types/api';
import { PluginConfigDialog } from '@/components/PluginConfigDialog';

export function PluginsPage() {
  const { currentOrganization } = useOrganization();
  const [plugins, setPlugins] = useState<Plugin[]>([]);
  const [configs, setConfigs] = useState<PluginConfiguration[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [_error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [compatFilter, setCompatFilter] = useState<string>('all');
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);
  const [showConfigDialog, setShowConfigDialog] = useState(false);

  useEffect(() => {
    fetchPlugins();
    if (currentOrganization) {
      fetchPluginConfigs();
    }
  }, [currentOrganization, searchQuery, compatFilter]);

  const fetchPlugins = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.getPlugins({ limit: 100, filter: { ...(searchQuery && { search: searchQuery }), ...(compatFilter !== 'all' && { compatibility: compatFilter }) } });
      setPlugins(res.items as import('@/types/api').Plugin[]);
    } catch (err) {
      console.error('Failed to fetch plugins:', err);
      setError('Failed to load plugins');
    } finally {
      setIsLoading(false);
    }
  };

  const fetchPluginConfigs = async () => {
    if (!currentOrganization) return;
    try {
      const res = await api.getPluginConfigs(currentOrganization.id, undefined, undefined, { limit: 100 });
      setConfigs(res.items as import('@/types/api').PluginConfiguration[]);
    } catch (err) {
      console.error('Failed to fetch plugin configs:', err);
    }
  };

  const handleConfigSaved = () => {
    setShowConfigDialog(false);
    setSelectedPlugin(null);
    fetchPluginConfigs();
  };

  const compatIcons: Record<string, React.ReactNode> = {
    VERIFIED: <CheckCircle className="h-4 w-4 text-green-600" />,
    PARTIALLY_VERIFIED: <AlertCircle className="h-4 w-4 text-yellow-600" />,
    UNVERIFIED: <HelpCircle className="h-4 w-4 text-gray-500" />,
    UNSUPPORTED: <XCircle className="h-4 w-4 text-red-600" />,
    DISABLED: <XCircle className="h-4 w-4 text-gray-500" />,
  };

  return (
    <LoadingOverlay isLoading={isLoading && plugins.length === 0}>
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Plugins</h1>
            <p className="text-dark-600 dark:text-dark-400 mt-1">Manage Camfrog plugins and their configurations</p>
          </div>
        </div>

        <Card className="overflow-hidden">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-dark-400" />
                <input
                  type="text"
                  placeholder="Search plugins..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 bg-white text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                />
              </div>
              <Select
                value={compatFilter}
                onChange={(e) => setCompatFilter(e.target.value)}
                options={[
                  { value: 'all', label: 'All' },
                  { value: 'VERIFIED', label: 'Verified' },
                  { value: 'PARTIALLY_VERIFIED', label: 'Partially Verified' },
                  { value: 'UNVERIFIED', label: 'Unverified' },
                  { value: 'UNSUPPORTED', label: 'Unsupported' },
                  { value: 'DISABLED', label: 'Disabled' },
                ]}
                className="w-full sm:w-48"
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-0">
            {plugins.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-dark-200 dark:border-dark-700 bg-dark-50 dark:bg-dark-900/50">
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Plugin</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Version</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Compatibility</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Capabilities</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Status</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Configuration</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-dark-200 dark:divide-dark-700">
                    {plugins.map((plugin) => {
                      const pluginConfigs = configs.filter(c => c.plugin_id === plugin.id);
                      const enabledConfigs = pluginConfigs.filter(c => c.is_enabled);
                      return (
                        <tr key={plugin.id} className="hover:bg-dark-50 dark:hover:bg-dark-800/50">
                          <td className="px-4 py-3">
                            <p className="font-medium text-dark-900 dark:text-dark-100">{plugin.name}</p>
                            <p className="text-xs text-dark-500 dark:text-dark-400 font-mono">#{plugin.plugin_number} • {truncate(plugin.id, 8)}</p>
                          </td>
                          <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400">{plugin.version}</td>
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              {compatIcons[plugin.compatibility] || <HelpCircle className="h-4 w-4 text-gray-500" />}
                              <StatusBadge status={plugin.compatibility} />
                            </div>
                          </td>
                          <td className="px-4 py-3">
                            <div className="flex flex-wrap gap-1">
                              {plugin.capability_flags?.map((cap, i) => (
                                <Badge key={i} variant="default" className="text-xs">{cap}</Badge>
                              ))}
                            </div>
                          </td>
                          <td className="px-4 py-3">
                            <Badge variant={plugin.is_enabled ? 'success' : 'default'}>
                              {plugin.is_enabled ? 'Enabled' : 'Disabled'}
                            </Badge>
                          </td>
                          <td className="px-4 py-3">
                            <div className="flex items-center gap-2">
                              {enabledConfigs.length > 0 && (
                                <Badge variant="info">{enabledConfigs.length} active config(s)</Badge>
                              )}
                              {pluginConfigs.length > 0 && (
                                <Badge variant="default">{pluginConfigs.length} total config(s)</Badge>
                              )}
                              {pluginConfigs.length === 0 && (
                                <span className="text-sm text-dark-400 dark:text-dark-500 italic">Not configured</span>
                              )}
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => {
                                  setSelectedPlugin(plugin);
                                  setShowConfigDialog(true);
                                }}
                              >
                                <ExternalLink className="h-4 w-4" />
                              </Button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-12">
                <Puzzle className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                <p className="text-dark-500 dark:text-dark-400">No plugins found</p>
              </div>
            )}
          </CardContent>
        </Card>

        <PluginConfigDialog
          isOpen={showConfigDialog}
          onClose={() => { setShowConfigDialog(false); setSelectedPlugin(null); }}
          onSuccess={handleConfigSaved}
          plugin={selectedPlugin}
          organizationId={currentOrganization?.id}
          existingConfigs={selectedPlugin ? configs.filter(c => c.plugin_id === selectedPlugin.id) : []}
        />
      </div>
    </LoadingOverlay>
  );
}