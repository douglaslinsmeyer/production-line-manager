import { Link } from 'react-router-dom';
import { PencilIcon, TrashIcon } from '@heroicons/react/24/outline';
import type { SignalProfile } from '../../api/types';
import Card from '../common/Card';

interface SignalProfileCardProps {
  profile: SignalProfile;
  onDelete?: (id: string) => void;
}

export default function SignalProfileCard({ profile, onDelete }: SignalProfileCardProps) {
  return (
    <Card className="hover:shadow-lg transition-shadow">
      <div className="p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <Link
              to={`/signal-profiles/${profile.id}`}
              className="text-lg font-semibold text-gray-900 hover:text-blue-600"
            >
              {profile.name}
            </Link>
            {profile.description && (
              <p className="text-sm text-gray-500 mt-1">{profile.description}</p>
            )}
          </div>

          {/* Version Badge */}
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
            v{profile.version}
          </span>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-4 mb-4">
          <div>
            <p className="text-sm text-gray-500">States</p>
            <p className="text-2xl font-semibold text-gray-900">{profile.states.length}</p>
          </div>
          <div>
            <p className="text-sm text-gray-500">Default State</p>
            <p className="text-sm font-medium text-gray-900">{profile.defaultState}</p>
          </div>
        </div>

        {/* Button Cycles */}
        <div className="text-xs text-gray-600 mb-4">
          <div>Short: {profile.buttonBehavior.shortPressCycle.join(' → ')}</div>
          <div>Long: {profile.buttonBehavior.longPressCycle.join(' → ')}</div>
        </div>

        {/* Actions */}
        <div className="flex gap-2 pt-4 border-t border-gray-200">
          <Link
            to={`/signal-profiles/${profile.id}/edit`}
            className="flex items-center gap-1 px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded-md transition-colors"
          >
            <PencilIcon className="w-4 h-4" />
            Edit
          </Link>
          {onDelete && (
            <button
              onClick={() => onDelete(profile.id)}
              className="flex items-center gap-1 px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-md transition-colors"
            >
              <TrashIcon className="w-4 h-4" />
              Delete
            </button>
          )}
        </div>
      </div>
    </Card>
  );
}
