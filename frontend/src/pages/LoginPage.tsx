import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Loader2, AlertCircle, Eye, EyeOff, Building2 } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from '@/components/ui/Card';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(8, 'Password must be at least 8 characters'),
  totpCode: z.string().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
      totpCode: '',
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setError(null);
    setIsLoading(true);
    
    try {
      await login(data.email, data.password, data.totpCode || undefined);
      navigate('/dashboard');
    } catch (err: unknown) {
      const axiosError = err as { response?: { data?: { message?: string } } };
      if (axiosError.response?.data?.message === 'MFA_REQUIRED') {
        setValue('totpCode', '');
        setError('Two-factor authentication required. Please enter your TOTP code.');
      } else {
        setError(axiosError.response?.data?.message || 'Login failed. Please check your credentials.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-dark-50 dark:bg-dark-950 px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <Link to="/" className="inline-flex items-center justify-center gap-2 mb-6" aria-label="CFC Bot Manager Home">
            <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-primary-600">
              <span className="text-white font-bold text-xl">CF</span>
            </div>
          </Link>
          <h1 className="text-3xl font-bold text-dark-900 dark:text-dark-100">Welcome back</h1>
          <p className="mt-2 text-dark-600 dark:text-dark-400">
            Sign in to your CFC Bot Manager account
          </p>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Sign in</CardTitle>
            <CardDescription>Enter your credentials to access your account</CardDescription>
          </CardHeader>
          
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
            {error && (
              <div className="flex items-center gap-2 rounded-lg bg-red-50 p-3 text-red-700 dark:bg-red-900/20 dark:text-red-400" role="alert">
                <AlertCircle className="h-5 w-5 flex-shrink-0" />
                <p className="text-sm">{error}</p>
              </div>
            )}

            <Input
              label="Email"
              type="email"
              placeholder="you@example.com"
              error={errors.email?.message}
              {...register('email')}
              autoComplete="email"
              disabled={isLoading}
            />

            <div className="relative">
              <Input
                label="Password"
                type={showPassword ? 'text' : 'password'}
                placeholder="••••••••"
                error={errors.password?.message}
                {...register('password')}
                autoComplete="current-password"
                disabled={isLoading}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-[38px] text-dark-400 hover:text-dark-600 dark:hover:text-dark-300"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
              </button>
            </div>

            {errors.totpCode && (
              <Input
                label="Two-Factor Code"
                type="text"
                placeholder="123456"
                error={errors.totpCode.message}
                {...register('totpCode')}
                autoComplete="one-time-code"
                disabled={isLoading}
                hint="Enter the 6-digit code from your authenticator app"
              />
            )}

            <Button type="submit" className="w-full" loading={isLoading}>
              {isLoading ? <Loader2 className="h-4 w-4" /> : 'Sign in'}
            </Button>
          </form>

          <CardFooter className="justify-center">
            <p className="text-sm text-dark-600 dark:text-dark-400">
              Don't have an account?{' '}
              <a href="#" className="font-medium text-primary-600 hover:text-primary-500 dark:text-primary-400">
                Contact administrator
              </a>
            </p>
          </CardFooter>
        </Card>

        <div className="mt-6 text-center text-sm text-dark-500 dark:text-dark-400">
          <p>CFC Bot Manager v0.1.0</p>
          <p className="mt-1">
            <Building2 className="h-4 w-4 inline-block align-middle mr-1" />
            Camfrog Bot Management Platform
          </p>
        </div>
      </div>
    </div>
  );
}