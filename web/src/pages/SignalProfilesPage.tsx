import { useState } from 'react';
import { Link } from 'react-router-dom';
import { PlusIcon, MagnifyingGlassIcon } from '@heroicons/react/24/outline';
import { useSignalProfiles, useDeleteSignalProfile } from '../hooks/useSignalProfiles';
import SignalProfileCard from '../components/signal-profiles/SignalProfileCard';
import Loading from '../components/common/Loading';
import Button from '../components/common/Button';
import Modal from '../components/common/Modal';

export default function SignalProfilesPage() {
  const { data: profiles, isLoading, error } = useSignalProfiles();
  const deleteProfile = useDeleteSignalProfile();
  const [searchQuery, setSearchQuery] = useState('');
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  const filteredProfiles = profiles?.filter((profile) =>
    profile.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleDelete = async () => {
    if (deleteConfirmId) {
      await deleteProfile.mutateAsync(deleteConfirmId);
      setDeleteConfirmId(null);
    }
  };

  if (isLoading) {
    return <Loading message="Loading signal profiles..." />;
  }

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600">Failed to load signal profiles</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Signal Profiles</h1>
          <p className="text-gray-600 mt-1">
            Manage configurable signal profiles for assembly line devices
          </p>
        </div>
        <Link to="/signal-profiles/new">
          <Button variant="primary">
            <PlusIcon className="w-5 h-5 mr-2" />
            Create Profile
          </Button>
        </Link>
      </div>

      {/* Search */}
      <div className="relative">
        <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
        <input
          type="text"
          placeholder="Search profiles..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
        />
      </div>

      {/* Profile Grid */}
      {filteredProfiles && filteredProfiles.length > 0 ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredProfiles.map((profile) => (
            <SignalProfileCard
              key={profile.id}
              profile={profile}
              onDelete={setDeleteConfirmId}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-12">
          <p className="text-gray-500">
            {searchQuery ? 'No profiles found matching your search' : 'No signal profiles yet'}
          </p>
          {!searchQuery && (
            <Link to="/signal-profiles/new">
              <Button variant="primary" className="mt-4">
                <PlusIcon className="w-5 h-5 mr-2" />
                Create Your First Profile
              </Button>
            </Link>
          )}
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteConfirmId !== null}
        onClose={() => setDeleteConfirmId(null)}
        title="Delete Signal Profile"
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete this profile? This action cannot be undone.
          </p>
          <p className="text-sm text-gray-500">
            Note: Profiles assigned to production lines cannot be deleted.
          </p>
          <div className="flex gap-3 justify-end">
            <Button variant="secondary" onClick={() => setDeleteConfirmId(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              onClick={handleDelete}
              disabled={deleteProfile.isPending}
            >
              {deleteProfile.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
