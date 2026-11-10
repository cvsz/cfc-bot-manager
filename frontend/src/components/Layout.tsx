import { ReactNode, useState } from 'react';
import { Outlet, Link, useLocation, NavLink } from 'react-router-dom';
import { cn } from '@/utils/helpers';
import { LayoutDashboard, Bot, Cpu, FileText, Settings, ChevronLeft, ChevronRight, LogOut, Menu, Bell } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useOrganization } from '@/context/OrganizationContext';
import { OrganizationSwitcher } from './OrganizationSwitcher';

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Rooms', href: '/rooms', icon: Bot },
  { name: 'Bots', href: '/bots', icon: Cpu },
  { name: 'Plugins', href: '/plugins', icon: FileText },
  { name: 'Audit Logs', href: '/audit-logs', icon: FileText },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export function Layout({ children }: { children?: ReactNode }) {
  const { user, logout } = useAuth();
  const { currentOrganization } = useOrganization();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="min-h-screen bg-dark-50 dark:bg-dark-950">
      {/* Mobile sidebar overlay */}
      {mobileMenuOpen && (
        <div 
          className="fixed inset-0 z-40 bg-black/50 lg:hidden" 
          onClick={() => setMobileMenuOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* Sidebar */}
      <aside
        className={cn(
          'fixed lg:static inset-y-0 left-0 z-50 bg-white dark:bg-dark-900 border-r border-dark-200 dark:border-dark-700 transition-all duration-200',
          sidebarOpen ? 'w-64' : 'w-16',
          mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        )}
        aria-label="Main navigation"
      >
        <div className="flex h-16 items-center justify-between px-4 border-b border-dark-200 dark:border-dark-700">
          <Link to="/dashboard" className="flex items-center gap-2" aria-label="CFC Bot Manager Home">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-600">
              <span className="text-white font-bold text-sm">CF</span>
            </div>
            {sidebarOpen && (
              <span className="font-semibold text-dark-900 dark:text-dark-100 truncate">CFC Bot Manager</span>
            )}
          </Link>
          
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className={cn(
              'lg:hidden p-2 rounded-lg text-dark-500 hover:bg-dark-100 dark:hover:bg-dark-800',
              !sidebarOpen && 'rotate-180'
            )}
            aria-label={sidebarOpen ? 'Collapse sidebar' : 'Expand sidebar'}
            aria-expanded={sidebarOpen}
          >
            {sidebarOpen ? <ChevronLeft className="h-5 w-5" /> : <ChevronRight className="h-5 w-5" />}
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-3 py-4 space-y-1">
          <nav aria-label="Main navigation">
            <ul className="space-y-1" role="list">
              {navigation.map((item) => {
                const isActive = location.pathname === item.href || 
                  (item.href !== '/dashboard' && location.pathname.startsWith(item.href));
                
                return (
                  <li key={item.name}>
                    <NavLink
                      to={item.href}
                      className={({ isActive }) => cn(
                        'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                        sidebarOpen
                          ? isActive
                            ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
                            : 'text-dark-600 hover:bg-dark-100 hover:text-dark-900 dark:text-dark-400 dark:hover:bg-dark-800 dark:hover:text-dark-100'
                          : 'justify-center text-dark-500 hover:bg-dark-100 hover:text-dark-700 dark:text-dark-500 dark:hover:bg-dark-800 dark:hover:text-dark-300'
                      )}
                      title={sidebarOpen ? undefined : item.name}
                      aria-current={isActive ? 'page' : undefined}
                    >
                      <item.icon className="h-5 w-5 flex-shrink-0" aria-hidden="true" />
                      {sidebarOpen && <span className="truncate">{item.name}</span>}
                    </NavLink>
                  </li>
                );
              })}
            </ul>
          </nav>
        </div>

        <div className="border-t border-dark-200 dark:border-dark-700 p-3">
          {sidebarOpen ? (
            <>
              <div className="mb-3">
                <OrganizationSwitcher />
              </div>
              <div className="flex items-center gap-3 px-2 py-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
                  <span className="text-sm font-medium text-primary-700 dark:text-primary-400">
                    {user?.display_name ? user.display_name[0].toUpperCase() : user?.email?.[0].toUpperCase() || 'U'}
                  </span>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-dark-900 dark:text-dark-100 truncate">
                    {user?.display_name || user?.email || 'User'}
                  </p>
                  <p className="text-xs text-dark-500 dark:text-dark-400 truncate">
                    {currentOrganization?.name || 'No organization'}
                  </p>
                </div>
                <button
                  onClick={handleLogout}
                  className="p-2 rounded-lg text-dark-500 hover:bg-dark-100 hover:text-dark-700 dark:hover:bg-dark-800"
                  aria-label="Log out"
                >
                  <LogOut className="h-5 w-5" />
                </button>
              </div>
            </>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <button
                onClick={handleLogout}
                className="p-2 rounded-lg text-dark-500 hover:bg-dark-100 hover:text-dark-700 dark:hover:bg-dark-800"
                aria-label="Log out"
                title="Log out"
              >
                <LogOut className="h-5 w-5" />
              </button>
            </div>
          )}
        </div>
      </aside>

      {/* Main content */}
      <div className={cn('lg:pl-64 transition-all duration-200', sidebarOpen ? 'lg:pl-64' : 'lg:pl-16')}>
        {/* Top bar */}
        <header className="sticky top-0 z-30 h-16 bg-white/80 dark:bg-dark-900/80 backdrop-blur-sm border-b border-dark-200 dark:border-dark-700">
          <div className="flex h-full items-center justify-between px-4 lg:px-6">
            <div className="flex items-center gap-4">
              <button
                onClick={() => setMobileMenuOpen(true)}
                className="lg:hidden p-2 rounded-lg text-dark-500 hover:bg-dark-100 dark:hover:bg-dark-800"
                aria-label="Open menu"
              >
                <Menu className="h-6 w-6" />
              </button>
              
              <h1 className="text-lg font-semibold text-dark-900 dark:text-dark-100">
                {navigation.find(n => location.pathname === n.href || 
                  (n.href !== '/dashboard' && location.pathname.startsWith(n.href)))?.name || 'Dashboard'}
              </h1>
            </div>

            <div className="flex items-center gap-4">
              <OrganizationSwitcher />
              
              <div className="hidden sm:flex items-center gap-3">
                <button className="p-2 rounded-lg text-dark-500 hover:bg-dark-100 dark:hover:bg-dark-800 relative" aria-label="Notifications">
                  <Bell className="h-5 w-5" />
                  <span className="absolute top-1 right-1 h-2 w-2 rounded-full bg-red-500" />
                </button>
              </div>
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-4 lg:p-6" role="main">
          {children || <Outlet />}
        </main>
      </div>
    </div>
  );
}