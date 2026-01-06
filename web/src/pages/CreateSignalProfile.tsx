import { useNavigate } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useCreateSignalProfile } from '../hooks/useSignalProfiles';
import SignalProfileForm from '../components/signal-profiles/SignalProfileForm';
import type { CreateSignalProfileRequest } from '../api/types';

export default function CreateSignalProfile() {
  const navigate = useNavigate();
  const createProfile = useCreateSignalProfile();

  const handleSubmit = async (data: CreateSignalProfileRequest) => {
    try {
      const newProfile = await createProfile.mutateAsync(data);
      navigate(`/signal-profiles/${newProfile.id}`);
    } catch (error) {
      // Error toast already shown by mutation
      console.error('Failed to create profile:', error);
    }
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <button
          onClick={() => navigate('/signal-profiles')}
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeftIcon className="w-4 h-4" />
          Back to Signal Profiles
        </button>
        <h1 className="text-2xl font-bold text-gray-900">Create Signal Profile</h1>
        <p className="text-gray-600 mt-1">
          Define states, outputs, and button behavior for assembly line devices
        </p>
      </div>

      {/* Form */}
      <SignalProfileForm
        onSubmit={handleSubmit}
        isSubmitting={createProfile.isPending}
        submitLabel="Create Profile"
      />
    </div>
  );
}
