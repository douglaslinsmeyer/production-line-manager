import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useSignalProfile, useUpdateSignalProfile } from '../hooks/useSignalProfiles';
import SignalProfileForm from '../components/signal-profiles/SignalProfileForm';
import Loading from '../components/common/Loading';
import type { UpdateSignalProfileRequest } from '../api/types';

export default function EditSignalProfile() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: profile, isLoading } = useSignalProfile(id!);
  const updateProfile = useUpdateSignalProfile(id!);

  const handleSubmit = async (data: UpdateSignalProfileRequest) => {
    try {
      await updateProfile.mutateAsync(data);
      navigate(`/signal-profiles/${id}`);
    } catch (error) {
      // Error toast already shown by mutation
      console.error('Failed to update profile:', error);
    }
  };

  if (isLoading) {
    return <Loading message="Loading profile..." />;
  }

  if (!profile) {
    return (
      <div className="text-center py-12">
        <p className="text-red-600">Profile not found</p>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <Link
          to={`/signal-profiles/${id}`}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeftIcon className="w-4 h-4" />
          Back to Profile
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Edit Signal Profile</h1>
          <p className="text-gray-600 mt-1">
            Editing will increment the version number and update all assigned devices
          </p>
          <p className="text-sm text-gray-500 mt-1">
            Current version: <span className="font-semibold">v{profile.version}</span> → Next
            version: <span className="font-semibold">v{profile.version + 1}</span>
          </p>
        </div>
      </div>

      {/* Form */}
      <SignalProfileForm
        initialData={profile}
        onSubmit={handleSubmit}
        isSubmitting={updateProfile.isPending}
        submitLabel="Update Profile"
      />
    </div>
  );
}
