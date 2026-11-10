import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Search, Bot, Edit, Trash2 } from 'lucide-react';
import { useOrganization } from '@/context/OrganizationContext';
import { Card, CardContent } from '@/components/ui/Card';
import { StatusBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { LoadingOverlay } from '@/components/ui/LoadingSpinner';
import { formatRelativeTime, truncate } from '@/utils/helpers';
import api from '@/api/client';
import type { Room } from '@/types/api';
import { RoomFormDialog } from '@/components/RoomFormDialog';

export function RoomsPage() {
  const { currentOrganization } = useOrganization();
  const [rooms, setRooms] = useState<Room[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [_error, setError] = useState<string | null>(null);
  
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [editingRoom, setEditingRoom] = useState<Room | null>(null);
  const [pagination, setPagination] = useState({ limit: 20, offset: 0, total: 0 });

  useEffect(() => {
    if (!currentOrganization) return;
    fetchRooms();
  }, [currentOrganization, pagination.offset, searchQuery, statusFilter]);

  const fetchRooms = async () => {
    if (!currentOrganization) return;
    setIsLoading(true);
    setError(null);
    try {
      const res = await api.getRooms(currentOrganization.id, {
        limit: pagination.limit,
        offset: pagination.offset,
        filter: {
          ...(searchQuery && { search: searchQuery }),
          ...(statusFilter !== 'all' && { status: statusFilter }),
        },
      });
      setRooms(res.items as Room[]);
      setPagination(prev => ({ ...prev, total: res.total_count }));
    } catch (err) {
      console.error('Failed to fetch rooms:', err);
      setError('Failed to load rooms');
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeleteRoom = async (roomId: string) => {
    if (!confirm('Are you sure you want to delete this room? This action cannot be undone.')) return;
    try {
      await api.deleteRoom(currentOrganization!.id, roomId);
      fetchRooms();
    } catch (err) {
      console.error('Failed to delete room:', err);
      alert('Failed to delete room');
    }
  };

  const handleRoomSaved = () => {
    setShowCreateDialog(false);
    setEditingRoom(null);
    fetchRooms();
  };

  return (
    <LoadingOverlay isLoading={isLoading && rooms.length === 0}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-100">Rooms</h1>
            <p className="text-dark-600 dark:text-dark-400 mt-1">Manage your Camfrog rooms</p>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="h-4 w-4 mr-2" /> Add Room
          </Button>
        </div>

        {/* Filters */}
        <Card className="overflow-hidden">
          <CardContent className="p-4">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1 max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-dark-400" />
                <input
                  type="text"
                  placeholder="Search rooms..."
                  value={searchQuery}
                  onChange={(e) => {
                    setSearchQuery(e.target.value);
                    setPagination(prev => ({ ...prev, offset: 0 }));
                  }}
                  className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 bg-white text-sm focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-100"
                />
              </div>
              <Select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPagination(prev => ({ ...prev, offset: 0 }));
                }}
                options={[
                  { value: 'all', label: 'All Statuses' },
                  { value: 'connected', label: 'Connected' },
                  { value: 'disconnected', label: 'Disconnected' },
                  { value: 'reconnecting', label: 'Reconnecting' },
                  { value: 'error', label: 'Error' },
                ]}
                className="w-full sm:w-48"
              />
            </div>
          </CardContent>
        </Card>

        {/* Rooms Table */}
        <Card>
          <CardContent className="p-0">
            {rooms.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b border-dark-200 dark:border-dark-700 bg-dark-50 dark:bg-dark-900/50">
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Room</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Status</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Language</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Bot</th>
                      <th className="px-4 py-3 text-left text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Last Seen</th>
                      <th className="px-4 py-3 text-right text-xs font-semibold text-dark-600 dark:text-dark-400 uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-dark-200 dark:divide-dark-700">
                    {rooms.map((room) => (
                      <tr key={room.id} className="hover:bg-dark-50 dark:hover:bg-dark-800/50">
                        <td className="px-4 py-3">
                          <div>
                            <p className="font-medium text-dark-900 dark:text-dark-100">{room.name}</p>
                            <p className="text-xs text-dark-500 dark:text-dark-400 font-mono">{truncate(room.id, 8)}</p>
                          </div>
                        </td>
                        <td className="px-4 py-3">
                          <StatusBadge status={room.status} />
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400">
                          {room.language}
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-600 dark:text-dark-400">
                          {room.bot_id ? (
                            <span className="font-mono text-xs">{truncate(room.bot_id, 8)}</span>
                          ) : (
                            <span className="text-dark-400 italic">No bot assigned</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-sm text-dark-500 dark:text-dark-400">
                          {room.last_seen_at ? formatRelativeTime(room.last_seen_at) : 'Never'}
                        </td>
                        <td className="px-4 py-3 text-right">
                          <div className="flex items-center justify-end gap-2">
                            <Link to={`/rooms/${room.id}`}>
                              <Button variant="ghost" size="sm">
                                <Edit className="h-4 w-4" />
                              </Button>
                            </Link>
                            <Button variant="ghost" size="sm" onClick={() => handleDeleteRoom(room.id)}>
                              <Trash2 className="h-4 w-4 text-red-600" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-12">
                <Bot className="h-12 w-12 text-dark-300 dark:text-dark-600 mx-auto mb-3" />
                <p className="text-dark-500 dark:text-dark-400">No rooms found</p>
                <Button variant="secondary" className="mt-4" onClick={() => setShowCreateDialog(true)}>
                  Create your first room
                </Button>
              </div>
            )}

            {/* Pagination */}
            {pagination.total > pagination.limit && (
              <div className="px-4 py-3 border-t border-dark-200 dark:border-dark-700 flex items-center justify-between">
                <p className="text-sm text-dark-600 dark:text-dark-400">
                  Showing {pagination.offset + 1} to {Math.min(pagination.offset + pagination.limit, pagination.total)} of {pagination.total}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={pagination.offset === 0}
                    onClick={() => setPagination(prev => ({ ...prev, offset: Math.max(0, prev.offset - prev.limit) }))}
                  >
                    Previous
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    disabled={pagination.offset + pagination.limit >= pagination.total}
                    onClick={() => setPagination(prev => ({ ...prev, offset: prev.offset + prev.limit }))}
                  >
                    Next
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Create/Edit Room Dialog */}
      <RoomFormDialog
        isOpen={showCreateDialog || !!editingRoom}
        onClose={() => {
          setShowCreateDialog(false);
          setEditingRoom(null);
        }}
        onSuccess={handleRoomSaved}
        initialData={editingRoom}
        organizationId={currentOrganization?.id}
      />
    </LoadingOverlay>
  );
}