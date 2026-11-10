import { useEffect, useState } from 'react';
import { X, Loader2, Plus, Trash2, Copy } from 'lucide-react';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { Card, CardHeader, CardTitle, CardFooter } from '@/components/ui/Card';
import api from '@/api/client';
import type { Plugin, PluginConfiguration } from '@/types/api';

const pluginConfigSchema = z.object({
  configs: z.array(z.object({
    id: z.string().optional(),
    room_id: z.string().uuid().optional().nullable(),
    bot_id: z.string().uuid().optional().nullable(),
    settings: z.string().default('{}'),
    is_enabled: z.boolean().default(true),
    version: z.number().default(1),
  })).min(1, 'At least one configuration is required'),
});

type PluginConfigFormData = z.infer<typeof pluginConfigSchema>;

interface PluginConfigDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  plugin: Plugin | null;
  organizationId: string | undefined;
  existingConfigs: PluginConfiguration[];
}

export function PluginConfigDialog({ isOpen, onClose, onSuccess, plugin, organizationId, existingConfigs }: PluginConfigDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [rooms, setRooms] = useState<Array<{ id: string; name: string }>>([]);
  const [bots, setBots] = useState<Array<{ id: string; name: string }>>([]);

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
    setValue,
  } = useForm<PluginConfigFormData>({
    resolver: zodResolver(pluginConfigSchema),
    defaultValues: {
      configs: [],
    },
  });

  const { fields, append, remove } = useFieldArray({ control, name: 'configs' });

  useEffect(() => {
    if (isOpen) {
      fetchRoomsAndBots();
      if (existingConfigs.length > 0) {
        reset({
          configs: existingConfigs.map(c => ({
            id: c.id,
            room_id: c.room_id,
            bot_id: c.bot_id,
            settings: c.settings,
            is_enabled: c.is_enabled,
            version: c.version,
          })),
        });
      } else {
        reset({ configs: [{ room_id: null, bot_id: null, settings: '{}', is_enabled: true, version: 1 }] });
      }
      setError(null);
    }
  }, [isOpen, existingConfigs, reset]);

  const fetchRoomsAndBots = async () => {
    if (!organizationId) return;
    try {
      const [roomsRes, botsRes] = await Promise.all([
        api.getRooms(organizationId, { limit: 100 }),
        api.getBots(organizationId, { limit: 100 }),
      ]);
      setRooms(roomsRes.items.map((r: { id: string; name: string }) => ({ id: r.id, name: r.name })) as { id: string; name: string }[]);
      setBots(botsRes.items.map((b: { id: string; name: string }) => ({ id: b.id, name: b.name })) as { id: string; name: string }[]);
    } catch (err) {
      console.error('Failed to fetch rooms/bots:', err);
    }
  };

  const addConfig = () => {
    append({ room_id: null, bot_id: null, settings: '{}', is_enabled: true, version: 1 });
  };

  const removeConfig = (index: number) => {
    remove(index);
  };

  const onSubmit = async (data: PluginConfigFormData) => {
    if (!organizationId || !plugin) return;
    setIsSubmitting(true);
    setError(null);

    try {
      for (const config of data.configs) {
        if (config.id) {
          // Update existing
          await api.updatePluginConfig(organizationId, config.id, {
            settings: JSON.parse(config.settings || '{}'),
            is_enabled: config.is_enabled,
            version: config.version,
          });
        } else {
          // Create new
          await api.createPluginConfig(organizationId, {
            plugin_id: plugin.id,
            room_id: config.room_id || undefined,
            bot_id: config.bot_id || undefined,
            settings: JSON.parse(config.settings || '{}'),
            is_enabled: config.is_enabled,
          });
        }
      }
      onSuccess();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } };
      setError(axiosError.response?.data?.message || 'Failed to save plugin configuration');
    } finally {
      setIsSubmitting(false);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  if (!isOpen || !plugin) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 overflow-y-auto" onClick={onClose}>
      <div className="w-full max-w-4xl bg-white dark:bg-dark-900 rounded-xl shadow-xl overflow-hidden min-h-[90vh] flex flex-col" onClick={e => e.stopPropagation()}>
        <CardHeader className="flex flex-row items-center justify-between border-b border-dark-200 dark:border-dark-700 p-4">
          <div>
            <CardTitle className="text-lg">Configure: {plugin.name}</CardTitle>
            <p className="text-sm text-dark-500 dark:text-dark-400">Plugin #{plugin.plugin_number} • {plugin.compatibility}</p>
          </div>
          <button onClick={onClose} className="p-1 rounded hover:bg-dark-100 dark:hover:bg-dark-800" aria-label="Close">
            <X className="h-5 w-5" />
          </button>
        </CardHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="flex-1 overflow-y-auto p-4 space-y-4" noValidate>
          {error && (
            <div className="flex items-center gap-2 rounded-lg bg-red-50 p-3 text-red-700 dark:bg-red-900/20 dark:text-red-400" role="alert">
              <p className="text-sm">{error}</p>
            </div>
          )}

          <div className="space-y-4">
            {fields.map((field, index) => (
              <Card key={field.id} className="p-4">
                <div className="flex items-center justify-between mb-4">
                  <h4 className="font-medium text-dark-900 dark:text-dark-100">Configuration #{index + 1}</h4>
                  {fields.length > 1 && (
                    <Button variant="ghost" size="sm" type="button" onClick={() => removeConfig(index)}>
                      <Trash2 className="h-4 w-4 text-red-600" />
                    </Button>
                  )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Select
                    label="Room (optional)"
                    value={watch(`configs.${index}.room_id`) || ''}
                    onChange={(e) => setValue(`configs.${index}.room_id`, e.target.value || null)}
                    options={[
                      { value: '', label: 'No room (global)' },
                      ...rooms.map(r => ({ value: r.id, label: r.name })),
                    ]}
                    placeholder="Select a room"
                    disabled={isSubmitting}
                  />

                  <Select
                    label="Bot (optional)"
                    value={watch(`configs.${index}.bot_id`) || ''}
                    onChange={(e) => setValue(`configs.${index}.bot_id`, e.target.value || null)}
                    options={[
                      { value: '', label: 'No bot (global)' },
                      ...bots.map(b => ({ value: b.id, label: b.name })),
                    ]}
                    placeholder="Select a bot"
                    disabled={isSubmitting}
                  />
                </div>

                <div className="mt-4">
                  <div className="flex items-center justify-between mb-2">
                    <label className="block text-sm font-medium text-dark-700 dark:text-dark-300">Settings (JSON)</label>
                    <Button variant="ghost" size="sm" type="button" onClick={() => copyToClipboard(watch(`configs.${index}.settings`) || '{}')}>
                      <Copy className="h-4 w-4 mr-1" /> Copy
                    </Button>
                  </div>
                  <textarea
                    {...register(`configs.${index}.settings`)}
                    rows={6}
                    className="w-full font-mono text-sm rounded-lg border border-dark-300 bg-white px-3 py-2 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                    placeholder='{"key": "value"}'
                    disabled={isSubmitting}
                  />
                  {errors.configs?.[index]?.settings && (
                    <p className="mt-1 text-sm text-red-600 dark:text-red-400">{errors.configs[index].settings.message}</p>
                  )}
                </div>

                <div className="flex items-center gap-4 mt-4">
                  <div className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      id={`is_enabled_${index}`}
                      {...register(`configs.${index}.is_enabled`)}
                      className="h-4 w-4 rounded border-dark-300 text-primary-600 focus:ring-primary-500"
                      disabled={isSubmitting}
                    />
                    <label htmlFor={`is_enabled_${index}`} className="text-sm text-dark-700 dark:text-dark-300 cursor-pointer">
                      Enabled
                    </label>
                  </div>
                  <input
                    type="hidden"
                    {...register(`configs.${index}.version`, { valueAsNumber: true })}
                  />
                </div>
              </Card>
            ))}
          </div>

          <div className="flex justify-end gap-2 pt-4 border-t border-dark-200 dark:border-dark-700">
            <Button type="button" variant="secondary" onClick={addConfig} disabled={isSubmitting}>
              <Plus className="h-4 w-4 mr-1" /> Add Configuration
            </Button>
          </div>
        </form>

        <CardFooter className="flex justify-end gap-2 border-t border-dark-200 dark:border-dark-700 p-4">
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="plugin-config-form" loading={isSubmitting}>
            {isSubmitting ? <Loader2 className="h-4 w-4" /> : 'Save Configurations'}
          </Button>
        </CardFooter>
      </div>
    </div>
  );
}