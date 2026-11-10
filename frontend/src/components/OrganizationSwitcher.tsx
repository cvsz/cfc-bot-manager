import { useState, useRef, useEffect } from 'react';
import { ChevronDown, Building2, Plus } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { cn } from '@/utils/helpers';
import { Button } from './ui/Button';

export function OrganizationSwitcher() {
  const { currentOrganization, organizations, setCurrentOrganization, createOrganization } = useOrganization();
  const [isOpen, setIsOpen] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [newOrgName, setNewOrgName] = useState('');
  const [newOrgSlug, setNewOrgSlug] = useState('');
  const [newOrgDescription, setNewOrgDescription] = useState('');
  const [createError, setCreateError] = useState('');
  const dropdownRef = useRef<HTMLDivElement>(null);
  const buttonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node) &&
          buttonRef.current && !buttonRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setIsCreating(false);
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleCreateOrganization = async (e: React.FormEvent) => {
    e.preventDefault();
    setCreateError('');
    
    if (!newOrgName.trim() || !newOrgSlug.trim()) {
      setCreateError('Name and slug are required');
      return;
    }

    try {
      const org = await createOrganization({
        name: newOrgName.trim(),
        slug: newOrgSlug.trim().toLowerCase(),
        description: newOrgDescription.trim() || undefined,
      });
      setCurrentOrganization(org);
      setIsCreating(false);
      setIsOpen(false);
      setNewOrgName('');
      setNewOrgSlug('');
      setNewOrgDescription('');
    } catch (error: unknown) {
      const axiosError = error as { response?: { data?: { message?: string } } };
      setCreateError(axiosError.response?.data?.message || 'Failed to create organization');
    }
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        ref={buttonRef}
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium',
          'bg-dark-100 text-dark-700 hover:bg-dark-200',
          'dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700',
          'border border-dark-200 dark:border-dark-700',
          'transition-colors'
        )}
        aria-haspopup="true"
        aria-expanded={isOpen}
        aria-label="Switch organization"
      >
        <Building2 className="h-4 w-4" />
        <span className="truncate max-w-[150px]">
          {currentOrganization?.name || 'Select organization'}
        </span>
        <ChevronDown className={cn('h-4 w-4 flex-shrink-0 transition-transform', isOpen && 'rotate-180')} />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-72 rounded-lg border border-dark-200 bg-white shadow-lg dark:border-dark-700 dark:bg-dark-900 z-50">
          {isCreating ? (
            <form onSubmit={handleCreateOrganization} className="p-4 space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-dark-900 dark:text-dark-100">Create Organization</h3>
                <button
                  type="button"
                  onClick={() => setIsCreating(false)}
                  className="p-1 rounded text-dark-500 hover:bg-dark-100 dark:hover:bg-dark-800"
                  aria-label="Cancel"
                >
                  <ChevronDown className="h-5 w-5 rotate-90" />
                </button>
              </div>
              <div>
                <label htmlFor="new-org-name" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">
                  Name
                </label>
                <input
                  id="new-org-name"
                  type="text"
                  value={newOrgName}
                  onChange={(e) => setNewOrgName(e.target.value)}
                  className="w-full rounded-lg border border-dark-300 bg-white px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                  placeholder="My Organization"
                  required
                  autoFocus
                />
              </div>
              <div>
                <label htmlFor="new-org-slug" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">
                  Slug
                </label>
                <input
                  id="new-org-slug"
                  type="text"
                  value={newOrgSlug}
                  onChange={(e) => setNewOrgSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                  className="w-full rounded-lg border border-dark-300 bg-white px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                  placeholder="my-organization"
                  required
                  pattern="[a-z0-9-]+"
                />
              </div>
              <div>
                <label htmlFor="new-org-description" className="block text-sm font-medium text-dark-700 dark:text-dark-300 mb-1">
                  Description (optional)
                </label>
                <textarea
                  id="new-org-description"
                  value={newOrgDescription}
                  onChange={(e) => setNewOrgDescription(e.target.value)}
                  rows={2}
                  className="w-full rounded-lg border border-dark-300 bg-white px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                  placeholder="Optional description"
                />
              </div>
              {createError && (
                <p className="text-sm text-red-600 dark:text-red-400" role="alert">{createError}</p>
              )}
              <div className="flex gap-2 pt-2">
                <Button type="submit" className="flex-1">Create</Button>
                <Button type="button" variant="secondary" onClick={() => setIsCreating(false)}>Cancel</Button>
              </div>
            </form>
          ) : (
            <>
              <div className="p-3 border-b border-dark-200 dark:border-dark-700">
                <Button variant="secondary" className="w-full justify-start gap-2" onClick={() => setIsCreating(true)}>
                  <Plus className="h-4 w-4" />
                  New Organization
                </Button>
              </div>
              <ul className="py-1" role="listbox" aria-label="Organizations">
                {organizations.map((org) => (
                  <li key={org.id}>
                    <button
                      onClick={() => {
                        setCurrentOrganization(org);
                        setIsOpen(false);
                      }}
                      className={cn(
                        'w-full flex items-center gap-3 px-3 py-2 text-sm text-left transition-colors',
                        currentOrganization?.id === org.id
                          ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-400'
                          : 'text-dark-700 hover:bg-dark-100 dark:text-dark-200 dark:hover:bg-dark-800'
                      )}
                      role="option"
                      aria-selected={currentOrganization?.id === org.id}
                    >
                      <Building2 className="h-4 w-4 flex-shrink-0" />
                      <span className="truncate">{org.name}</span>
                      {currentOrganization?.id === org.id && (
                        <span className="ml-auto text-primary-600 dark:text-primary-400">✓</span>
                      )}
                    </button>
                  </li>
                ))}
                {organizations.length === 0 && (
                  <li className="px-3 py-4 text-center text-sm text-dark-500 dark:text-dark-400">
                    No organizations yet
                  </li>
                )}
              </ul>
            </>
          )}
        </div>
      )}
    </div>
  );
}