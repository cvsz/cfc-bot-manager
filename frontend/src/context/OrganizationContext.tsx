import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import api from '@/api/client';
import type { Organization } from '@/types/api';

interface OrganizationContextType {
  currentOrganization: Organization | null;
  organizations: Organization[];
  isLoading: boolean;
  setCurrentOrganization: (org: Organization | null) => void;
  fetchOrganizations: () => Promise<void>;
  createOrganization: (data: { name: string; slug: string; description?: string }) => Promise<Organization>;
  updateOrganization: (id: string, data: Partial<{ name: string; slug: string; description: string }>) => Promise<Organization>;
  deleteOrganization: (id: string) => Promise<void>;
}

const OrganizationContext = createContext<OrganizationContextType | undefined>(undefined);

export function OrganizationProvider({ children }: { children: ReactNode }) {
  const [currentOrganization, setCurrentOrganization] = useState<Organization | null>(null);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [isLoading] = useState(false);

  const fetchOrganizations = useCallback(async () => {
    try {
      const response = await api.getOrganizations({ limit: 100 });
      setOrganizations(response.items as Organization[]);
      
      // If no current org selected, select the first one
      if (!currentOrganization && (response.items as Organization[]).length > 0) {
        setCurrentOrganization((response.items as Organization[])[0]);
      }
    } catch (error) {
      console.error('Failed to fetch organizations:', error);
    }
  }, [currentOrganization]);

  useEffect(() => {
    fetchOrganizations();
  }, [fetchOrganizations]);

  const createOrganization = async (data: { name: string; slug: string; description?: string }) => {
    const org = await api.createOrganization(data) as Organization;
    setOrganizations(prev => [...prev, org]);
    return org;
  };

  const updateOrganization = async (id: string, data: Partial<{ name: string; slug: string; description: string }>) => {
    const org = await api.updateOrganization(id, data) as Organization;
    setOrganizations(prev => prev.map(o => o.id === id ? org : o));
    if (currentOrganization?.id === id) {
      setCurrentOrganization(org);
    }
    return org;
  };

  const deleteOrganization = async (id: string) => {
    await api.deleteOrganization(id);
    setOrganizations(prev => prev.filter(o => o.id !== id));
    if (currentOrganization?.id === id) {
      setCurrentOrganization(organizations.find(o => o.id !== id) || null);
    }
  };

  return (
    <OrganizationContext.Provider value={{
      currentOrganization,
      organizations,
      isLoading,
      setCurrentOrganization,
      fetchOrganizations,
      createOrganization,
      updateOrganization,
      deleteOrganization,
    }}>
      {children}
    </OrganizationContext.Provider>
  );
}

export function useOrganization(): OrganizationContextType {
  const context = useContext(OrganizationContext);
  if (!context) {
    throw new Error('useOrganization must be used within an OrganizationProvider');
  }
  return context;
}