import { useEffect, useState } from 'react';
import { X, Loader2 } from 'lucide-react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { CardHeader, CardTitle, CardFooter } from '@/components/ui/Card';
import api from '@/api/client';
import type { Bot, BotConfig } from '@/types/api';

const botSchema = z.object({
  name: z.string().min(1, 'Name is required').max(128, 'Name too long'),
  username: z.string().min(1, 'Username is required').max(64, 'Username too long'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  room_id: z.string().uuid().optional(),
  config: z.object({
    auto_reconnect: z.boolean().default(true),
    log_level: z.enum(['debug', 'info', 'warn', 'error']).default('info'),
    proxy_url: z.string().url().optional().or(z.literal('')).nullable().default(''),
  }).default({
    auto_reconnect: true,
    log_level: 'info',
    proxy_url: '',
  }),
});

type BotFormData = z.infer<typeof botSchema>;

interface BotFormDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
  initialData: Bot | null;
  organizationId: string | undefined;
}

export function BotFormDialog({ isOpen, onClose, onSuccess, initialData, organizationId }: BotFormDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [rooms, setRooms] = useState<Array<{ id: string; name: string }>>([]);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
    watch,
    setValue,
  } = useForm<BotFormData>({
    resolver: zodResolver(botSchema),
    defaultValues: {
      name: '',
      username: '',
      password: '',
      room_id: undefined,
      config: {
        auto_reconnect: true,
        log_level: 'info',
        proxy_url: '',
      },
    },
  });

  useEffect(() => {
    if (isOpen) {
      fetchRooms();
      if (initialData) {
        reset({
          name: initialData.name,
          username: initialData.username,
          password: '', // Don't pre-fill for security
          room_id: initialData.room_id || undefined,
          config: initialData.config,
        });
      } else {
        reset({
          name: '',
          username: '',
          password: '',
          room_id: undefined,
          config: {
            auto_reconnect: true,
            log_level: 'info',
            proxy_url: '',
          },
        });
      }
      setError(null);
    }
  }, [isOpen, initialData, reset]);

  const fetchRooms = async () => {
    if (!organizationId) return;
    try {
      const res = await api.getRooms(organizationId, { limit: 100 });
      setRooms(res.items.map((r: { id: string; name: string }) => ({ id: r.id, name: r.name })));
    } catch (err) {
      console.error('Failed to fetch rooms:', err);
    }
  };

  const onSubmit = async (data: BotFormData) => {
    if (!organizationId) return;
    setIsSubmitting(true);
    setError(null);

    try {
      const payload = {
        name: data.name,
        username: data.username,
        password: data.password,
        room_id: data.room_id || undefined,
        config: {
          auto_reconnect: data.config.auto_reconnect,
          log_level: data.config.log_level,
          proxy_url: data.config.proxy_url || null,
        },
      };

      if (initialData) {
        await api.updateBot(organizationId, initialData.id, payload);
      } else {
        await api.createBot(organizationId, payload);
      }

      onSuccess();
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } };
      setError(axiosError.response?.data?.message || 'Failed to save bot');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50" onClick={onClose}>
      <div className="w-full max-w-lg bg-white dark:bg-dark-900 rounded-xl shadow-xl overflow-hidden" onClick={e => e.stopPropagation()}>
        <CardHeader className="flex flex-row items-center justify-between border-b border-dark-200 dark:border-dark-700 p-4">
          <CardTitle className="text-lg">{initialData ? 'Edit Bot' : 'Create Bot'}</CardTitle>
          <button onClick={onClose} className="p-1 rounded hover:bg-dark-100 dark:hover:bg-dark-800" aria-label="Close">
            <X className="h-5 w-5" />
          </button>
        </CardHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="p-4 space-y-4" noValidate>
          {error && (
            <div className="flex items-center gap-2 rounded-lg bg-red-50 p-3 text-red-700 dark:bg-red-900/20 dark:text-red-400" role="alert">
              <p className="text-sm">{error}</p>
            </div>
          )}

          <Input
            label="Bot Name"
            placeholder="e.g. Main Bot"
            error={errors.name?.message}
            {...register('name')}
            disabled={isSubmitting}
          />

          <Input
            label="Camfrog Username"
            placeholder="bot_username"
            error={errors.username?.message}
            {...register('username')}
            disabled={isSubmitting}
          />

          <Input
            label="Camfrog Password"
            type="password"
            placeholder={initialData ? 'Leave blank to keep current' : 'Enter password'}
            error={errors.password?.message}
            {...register('password')}
            disabled={isSubmitting}
            hint={initialData ? 'Leave blank to keep current password' : 'Minimum 8 characters'}
          />

          <Select
            label="Assign Room (optional)"
            value={watch('room_id') || ''}
            onChange={(e) => setValue('room_id', e.target.value || undefined)}
            options={[
              { value: '', label: 'No room assigned' },
              ...rooms.map((r: { id: string; name: string }) => ({ value: r.id, label: r.name })),
            ]}
            placeholder="Select a room"
            disabled={isSubmitting}
          />

          <div className="space-y-3 pt-2 border-t border-dark-200 dark:border-dark-700">
            <h4 className="text-sm font-medium text-dark-900 dark:text-dark-100">Configuration</h4>
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                id="auto_reconnect"
                {...register('config.auto_reconnect')}
                className="h-4 w-4 rounded border-dark-300 text-primary-600 focus:ring-primary-500"
                disabled={isSubmitting}
              />
              <label htmlFor="auto_reconnect" className="text-sm text-dark-700 dark:text-dark-300 cursor-pointer">
                Auto Reconnect
              </label>
            </div>
            <Select
              label="Log Level"
              value={watch('config.log_level')}
              onChange={(e) => setValue('config.log_level', e.target.value as BotConfig['log_level'])}
              options={[
                { value: 'debug', label: 'Debug' },
                { value: 'info', label: 'Info' },
                { value: 'warn', label: 'Warn' },
                { value: 'error', label: 'Error' },
              ]}
              disabled={isSubmitting}
            />
            <Input
              label="Proxy URL (optional)"
              type="url"
              placeholder="http://proxy.example.com:8080"
              error={errors.config?.proxy_url?.message}
              {...register('config.proxy_url')}
              disabled={isSubmitting}
            />
          </div>
        </form>

        <CardFooter className="flex justify-end gap-2 border-t border-dark-200 dark:border-dark-700 p-4">
          <Button variant="secondary" onClick={onClose} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button type="submit" form="bot-form" loading={isSubmitting}>
            {isSubmitting ? <Loader2 className="h-4 w-4" /> : (initialData ? 'Save Changes' : 'Create Bot')}
          </Button>
        </CardFooter>
      </div>
    </div>
  );
}