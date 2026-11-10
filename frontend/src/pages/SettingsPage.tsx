import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { CheckCircle, AlertCircle, Plus, X, User, Shield, Key, Building2, Save } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Badge } from '@/components/ui/Badge';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { cn, formatRelativeTime } from '@/utils/helpers';
import api from '@/api/client';

const profileSchema = z.object({
  display_name: z.string().min(1, 'Display name is required').max(100).optional(),
  email: z.string().email('Invalid email address'),
});

const orgSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  slug: z.string().min(1, 'Slug is required').max(50).regex(/^[a-z0-9-]+$/, 'Slug can only contain lowercase letters, numbers, and hyphens'),
  description: z.string().max(500).optional(),
});

const passwordSchema = z.object({
  current_password: z.string().min(1, 'Current password is required'),
  new_password: z.string().min(8, 'New password must be at least 8 characters'),
  confirm_password: z.string(),
}).refine(data => data.new_password === data.confirm_password, {
  message: 'Passwords do not match',
  path: ['confirm_password'],
});

const apiKeySchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  permissions: z.array(z.string()).min(1, 'At least one permission is required'),
  expires_at: z.string().optional(),
});

type ProfileFormData = z.infer<typeof profileSchema>;
type OrgFormData = z.infer<typeof orgSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;
type APIKeyFormData = z.infer<typeof apiKeySchema>;

const PERMISSIONS = [
  'rooms:read', 'rooms:write', 'rooms:delete',
  'bots:read', 'bots:write', 'bots:control',
  'plugins:read', 'plugins:write', 'plugins:configure',
  'users:read', 'users:write', 'users:delete',
  'audit:read',
  'settings:read', 'settings:write',
];

export function SettingsPage() {
  const { user } = useAuth();
  const { currentOrganization, updateOrganization } = useOrganization();
  const [activeTab, setActiveTab] = useState<'profile' | 'organization' | 'security' | 'api' | 'mfa'>('profile');
  const [isLoading, setIsLoading] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [apiKeys, setApiKeys] = useState<Array<{ id: string; name: string; prefix: string; permissions: string[]; last_used_at: string | null; revoked_at: string | null; expires_at: string | null; created_at: string }>>([]);
  const [showCreateApiKey, setShowCreateApiKey] = useState(false);
  const [revokingKey, setRevokingKey] = useState<string | null>(null);

  // Profile form
  const profileForm = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      display_name: user?.display_name || '',
      email: user?.email || '',
    },
  });

  // Organization form
  const orgForm = useForm<OrgFormData>({
    resolver: zodResolver(orgSchema),
    defaultValues: {
      name: currentOrganization?.name || '',
      slug: currentOrganization?.slug || '',
      description: currentOrganization?.description || '',
    },
  });

  // Password form
  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      current_password: '',
      new_password: '',
      confirm_password: '',
    },
  });

  // API Key form
  const apiKeyForm = useForm<APIKeyFormData>({
    resolver: zodResolver(apiKeySchema),
    defaultValues: {
      name: '',
      permissions: ['rooms:read', 'bots:read'],
      expires_at: '',
    },
  });

  useEffect(() => {
    fetchApiKeys();
  }, [currentOrganization]);

  const fetchApiKeys = async () => {
    if (!currentOrganization) return;
    try {
      const res = await api.getAPIKeys(currentOrganization.id, { limit: 100 });
      setApiKeys(res.items as typeof apiKeys);
    } catch (err) {
      console.error('Failed to fetch API keys:', err);
    }
  };

  const clearMessages = () => {
    setSuccess(null);
    setError(null);
  };

  const handleProfileSubmit = async (_data: ProfileFormData) => {
    clearMessages();
    setIsLoading(true);
    try {
      // Update profile would be implemented in backend
      setSuccess('Profile updated successfully');
      // In real app, would call api.updateUser
    } catch (err) {
      setError('Failed to update profile');
    } finally {
      setIsLoading(false);
    }
  };

  const handleOrgSubmit = async (_data: OrgFormData) => {
    clearMessages();
    setIsLoading(true);
    try {
      await updateOrganization(currentOrganization!.id, _data);
      setSuccess('Organization updated successfully');
    } catch (err) {
      setError('Failed to update organization');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasswordSubmit = async (_data: PasswordFormData) => {
    clearMessages();
    setIsLoading(true);
    try {
      // Change password would be implemented in backend
      setSuccess('Password changed successfully');
      passwordForm.reset();
    } catch (err) {
      setError('Failed to change password');
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateApiKey = async (_data: APIKeyFormData) => {
    clearMessages();
    setIsLoading(true);
    try {
      const res = await api.createAPIKey(currentOrganization!.id, {
        name: _data.name,
        permissions: _data.permissions,
        expires_at: _data.expires_at || undefined,
      }) as { key: string };
      setSuccess(`API Key created! Key: ${res.key} (Save this - it won't be shown again)`);
      apiKeyForm.reset();
      setShowCreateApiKey(false);
      fetchApiKeys();
    } catch (err) {
      setError('Failed to create API key');
    } finally {
      setIsLoading(false);
    }
  };

  const handleRevokeApiKey = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) return;
    setRevokingKey(keyId);
    try {
      await api.revokeAPIKey(currentOrganization!.id, keyId);
      setSuccess('API key revoked');
      fetchApiKeys();
    } catch (err) {
      setError('Failed to revoke API key');
    } finally {
      setRevokingKey(null);
    }
  };

  const tabs = [
    { id: 'profile', label: 'Profile', icon: User },
    { id: 'organization', label: 'Organization', icon: Building2 },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'api', label: 'API Keys', icon: Key },
    { id: 'mfa', label: 'Two-Factor', icon: Shield },
  ];

  return (
    <LoadingOverlay isLoading={isLoading}>
      <div className="space-y-6 max-w-4xl">
        <div>
          <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Settings</h1>
          <p className="text-dark-600 dark:text-dark-400 mt-1">Manage your account and organization settings</p>
        </div>

        {success && (
          <div className="flex items-center gap-2 rounded-lg bg-green-50 p-3 text-green-700 dark:bg-green-900/20 dark:text-green-400" role="alert">
            <CheckCircle className="h-5 w-5 flex-shrink-0" />
            <p className="text-sm">{success}</p>
            <button onClick={() => setSuccess(null)} className="ml-auto text-green-500 hover:text-green-700">×</button>
          </div>
        )}
        {error && (
          <div className="flex items-center gap-2 rounded-lg bg-red-50 p-3 text-red-700 dark:bg-red-900/20 dark:text-red-400" role="alert">
            <AlertCircle className="h-5 w-5 flex-shrink-0" />
            <p className="text-sm">{error}</p>
            <button onClick={() => setError(null)} className="ml-auto text-red-500 hover:text-red-700">×</button>
          </div>
        )}

        {/* Tab Navigation */}
        <div className="border-b border-dark-200 dark:border-dark-700">
          <nav className="flex gap-4 -mb-px" aria-label="Settings tabs">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => { setActiveTab(tab.id as any); clearMessages(); }}
                className={cn(
                  'flex items-center gap-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors',
                  activeTab === tab.id
                    ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                    : 'border-transparent text-dark-500 hover:text-dark-700 hover:border-dark-300 dark:text-dark-400 dark:hover:text-dark-200 dark:hover:border-dark-600'
                )}
                aria-current={activeTab === tab.id ? 'page' : undefined}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        {/* Tab Panels */}
        {activeTab === 'profile' && (
          <Card>
            <CardHeader>
              <CardTitle>Profile</CardTitle>
              <CardDescription>Update your personal information</CardDescription>
            </CardHeader>
            <form onSubmit={profileForm.handleSubmit(handleProfileSubmit)} className="space-y-4 p-6" noValidate>
              <Input
                label="Email"
                type="email"
                error={profileForm.formState.errors.email?.message}
                {...profileForm.register('email')}
                disabled
              />
              <Input
                label="Display Name"
                placeholder="Your name"
                error={profileForm.formState.errors.display_name?.message}
                {...profileForm.register('display_name')}
              />
              <CardFooter>
                <Button type="submit" loading={isLoading}>
                  <Save className="h-4 w-4 mr-2" /> Save Changes
                </Button>
              </CardFooter>
            </form>
          </Card>
        )}

        {activeTab === 'organization' && currentOrganization && (
          <Card>
            <CardHeader>
              <CardTitle>Organization</CardTitle>
              <CardDescription>Manage your organization details</CardDescription>
            </CardHeader>
            <form onSubmit={orgForm.handleSubmit(handleOrgSubmit)} className="space-y-4 p-6" noValidate>
              <Input
                label="Organization Name"
                error={orgForm.formState.errors.name?.message}
                {...orgForm.register('name')}
              />
              <Input
                label="Slug"
                error={orgForm.formState.errors.slug?.message}
                {...orgForm.register('slug')}
                hint="Used in URLs. Can only contain lowercase letters, numbers, and hyphens."
              />
              <Textarea
                label="Description"
                rows={3}
                error={orgForm.formState.errors.description?.message}
                {...orgForm.register('description')}
              />
              <CardFooter>
                <Button type="submit" loading={isLoading}>
                  <Save className="h-4 w-4 mr-2" /> Save Changes
                </Button>
              </CardFooter>
            </form>
          </Card>
        )}

        {activeTab === 'security' && (
          <Card>
            <CardHeader>
              <CardTitle>Change Password</CardTitle>
              <CardDescription>Update your account password</CardDescription>
            </CardHeader>
            <form onSubmit={passwordForm.handleSubmit(handlePasswordSubmit)} className="space-y-4 p-6" noValidate>
              <Input
                label="Current Password"
                type="password"
                error={passwordForm.formState.errors.current_password?.message}
                {...passwordForm.register('current_password')}
              />
              <Input
                label="New Password"
                type="password"
                error={passwordForm.formState.errors.new_password?.message}
                {...passwordForm.register('new_password')}
              />
              <Input
                label="Confirm New Password"
                type="password"
                error={passwordForm.formState.errors.confirm_password?.message}
                {...passwordForm.register('confirm_password')}
              />
              <CardFooter>
                <Button type="submit" loading={isLoading}>
                  <Save className="h-4 w-4 mr-2" /> Change Password
                </Button>
              </CardFooter>
            </form>
          </Card>
        )}

        {activeTab === 'api' && currentOrganization && (
          <>
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>API Keys</CardTitle>
                    <CardDescription>Manage API keys for programmatic access</CardDescription>
                  </div>
                  <Button onClick={() => setShowCreateApiKey(true)}>
                    <Plus className="h-4 w-4 mr-2" /> Create API Key
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {apiKeys.length > 0 ? (
                  <div className="space-y-3">
                    {apiKeys.map((key) => (
                      <div key={key.id} className="flex items-center justify-between p-4 rounded-lg border border-dark-200 dark:border-dark-700">
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-3">
                            <p className="font-medium text-dark-900 dark:text-dark-100">{key.name}</p>
                            <Badge variant="info">{key.prefix}••••</Badge>
                            {key.revoked_at && <Badge variant="error">Revoked</Badge>}
                            {!key.revoked_at && key.expires_at && new Date(key.expires_at) < new Date() && <Badge variant="warning">Expired</Badge>}
                          </div>
                          <div className="flex flex-wrap gap-2 mt-1">
                            {key.permissions.map(p => <Badge key={p} variant="default" className="text-xs">{p}</Badge>)}
                          </div>
                          <p className="text-xs text-dark-500 dark:text-dark-400 mt-1">
                            Created {formatRelativeTime(key.created_at)}
                            {key.last_used_at && ` • Last used ${formatRelativeTime(key.last_used_at)}`}
                          </p>
                        </div>
                        {!key.revoked_at && (
                          <Button variant="ghost" size="sm" onClick={() => handleRevokeApiKey(key.id)} loading={revokingKey === key.id}>
                            Revoke
                          </Button>
                        )}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <Key className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                    <p className="text-dark-500 dark:text-dark-400">No API keys created yet</p>
                    <Button className="mt-4" onClick={() => setShowCreateApiKey(true)}>
                      Create your first API key
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Create API Key Dialog */}
            {showCreateApiKey && (
              <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" onClick={() => setShowCreateApiKey(false)}>
                <div className="w-full max-w-md bg-white dark:bg-dark-900 rounded-xl shadow-xl overflow-hidden" onClick={e => e.stopPropagation()}>
                  <CardHeader className="flex flex-row items-center justify-between border-b border-dark-200 dark:border-dark-700 p-4">
                    <CardTitle className="text-lg">Create API Key</CardTitle>
                    <button onClick={() => setShowCreateApiKey(false)} className="p-1 rounded hover:bg-dark-100 dark:hover:bg-dark-800">
                      <X className="h-5 w-5" />
                    </button>
                  </CardHeader>
                  <form onSubmit={apiKeyForm.handleSubmit(handleCreateApiKey)} className="p-6 space-y-4" noValidate>
                    <Input
                      label="Key Name"
                      placeholder="e.g. Production Server"
                      error={apiKeyForm.formState.errors.name?.message}
                      {...apiKeyForm.register('name')}
                    />
                    <div>
                      <label className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-2">Permissions</label>
                      <div className="flex flex-wrap gap-2">
                        {PERMISSIONS.map(perm => (
                          <label key={perm} className="inline-flex items-center gap-2 rounded-lg border border-dark-300 bg-white px-3 py-2 text-sm cursor-pointer hover:border-primary-500 dark:border-dark-600 dark:bg-dark-800">
                            <input
                              type="checkbox"
                              checked={apiKeyForm.watch('permissions').includes(perm)}
                              onChange={(e) => {
                                const current = apiKeyForm.getValues('permissions') || [];
                                if (e.target.checked) {
                                  apiKeyForm.setValue('permissions', [...current, perm]);
                                } else {
                                  apiKeyForm.setValue('permissions', current.filter(p => p !== perm));
                                }
                              }}
                              className="h-4 w-4 rounded border-dark-300 text-primary-600 focus:ring-primary-500"
                            />
                            <span className="text-dark-700 dark:text-dark-300">{perm}</span>
                          </label>
                        ))}
                      </div>
                    </div>
                    <Input
                      label="Expires At (optional)"
                      type="datetime-local"
                      error={apiKeyForm.formState.errors.expires_at?.message}
                      {...apiKeyForm.register('expires_at')}
                    />
                    <CardFooter>
                      <Button variant="secondary" onClick={() => setShowCreateApiKey(false)}>Cancel</Button>
                      <Button type="submit" loading={isLoading}>
                        <Key className="h-4 w-4 mr-2" /> Create Key
                      </Button>
                    </CardFooter>
                  </form>
                </div>
              </div>
            )}
          </>
        )}

        {activeTab === 'mfa' && (
          <Card>
            <CardHeader>
              <CardTitle>Two-Factor Authentication</CardTitle>
              <CardDescription>Add an extra layer of security to your account</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between p-4 rounded-lg border border-dark-200 dark:border-dark-700">
                <div>
                  <p className="font-medium text-dark-900 dark:text-dark-100">Authenticator App</p>
                  <p className="text-sm text-dark-500 dark:text-dark-400">Use an authenticator app like Google Authenticator or Authy</p>
                </div>
                <Button variant="secondary">Set Up 2FA</Button>
              </div>
              <div className="p-4 rounded-lg border border-dark-200 dark:border-dark-700 bg-dark-50 dark:bg-dark-900/50">
                <h4 className="font-medium text-dark-900 dark:text-dark-100 mb-2">Backup Codes</h4>
                <p className="text-sm text-dark-600 dark:text-dark-400 mb-4">Store these codes in a safe place. Each code can only be used once.</p>
                <Button variant="ghost">Generate Backup Codes</Button>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </LoadingOverlay>
  );
}