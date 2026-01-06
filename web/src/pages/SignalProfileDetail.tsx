import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon, PencilIcon, TrashIcon, ClockIcon } from '@heroicons/react/24/outline';
import {
  useSignalProfile,
  useProfileVersions,
  useProfileDeviceStatus,
  useDeleteSignalProfile,
} from '../hooks/useSignalProfiles';
import StateOutputPreview from '../components/signal-profiles/StateOutputPreview';
import Loading from '../components/common/Loading';
import Button from '../components/common/Button';
import Card from '../components/common/Card';
import { formatDistanceToNow } from 'date-fns';

export default function SignalProfileDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: profile, isLoading } = useSignalProfile(id!);
  const { data: versions } = useProfileVersions(id!);
  const { data: deviceStatus } = useProfileDeviceStatus(id!);
  const deleteProfile = useDeleteSignalProfile();

  const handleDelete = async () => {
    if (confirm('Are you sure you want to delete this profile?')) {
      try {
        await deleteProfile.mutateAsync(id!);
        navigate('/signal-profiles');
      } catch (error) {
        // Error toast shown by mutation
      }
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

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'up-to-date':
        return 'bg-green-100 text-green-800';
      case 'update-pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'failed':
        return 'bg-red-100 text-red-800';
      case 'offline':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="max-w-5xl mx-auto space-y-6">
      {/* Header */}
      <div>
        <Link
          to="/signal-profiles"
          className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
        >
          <ArrowLeftIcon className="w-4 h-4" />
          Back to Signal Profiles
        </Link>

        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-2xl font-bold text-gray-900">{profile.name}</h1>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                Version {profile.version}
              </span>
            </div>
            {profile.description && (
              <p className="text-gray-600 mt-2">{profile.description}</p>
            )}
          </div>

          <div className="flex gap-2">
            <Link to={`/signal-profiles/${profile.id}/edit`}>
              <Button variant="secondary">
                <PencilIcon className="w-4 h-4 mr-2" />
                Edit
              </Button>
            </Link>
            <Button variant="danger" onClick={handleDelete} disabled={deleteProfile.isPending}>
              <TrashIcon className="w-4 h-4 mr-2" />
              Delete
            </Button>
          </div>
        </div>
      </div>

      {/* States */}
      <Card title="States">
        <div className="space-y-4">
          {profile.states.map((state, index) => (
            <div
              key={index}
              className={`flex items-center gap-4 p-4 rounded-lg ${
                state.name === profile.defaultState ? 'bg-blue-50 border border-blue-200' : 'bg-gray-50'
              }`}
            >
              <StateOutputPreview outputs={state.outputs} />
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-gray-900">{state.name}</h3>
                  {state.name === profile.defaultState && (
                    <span className="text-xs px-2 py-0.5 bg-blue-600 text-white rounded">
                      Default
                    </span>
                  )}
                </div>
                <p className="text-sm text-gray-600 mt-1">
                  R:{state.outputs.redLight} Y:{state.outputs.yellowLight} G:
                  {state.outputs.greenLight} Buzzer:{state.outputs.buzzer}
                </p>
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* Button Behavior */}
      <Card title="Button Behavior">
        <div className="space-y-4">
          <div>
            <h4 className="text-sm font-medium text-gray-700 mb-2">Short Press Cycle (&lt; 1s)</h4>
            <div className="flex items-center gap-2 flex-wrap">
              {profile.buttonBehavior.shortPressCycle.map((stateName, index) => (
                <span key={index} className="flex items-center gap-2">
                  <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-md text-sm">
                    {stateName}
                  </span>
                  {index < profile.buttonBehavior.shortPressCycle.length - 1 && (
                    <span className="text-gray-400">→</span>
                  )}
                </span>
              ))}
              {profile.buttonBehavior.shortPressCycle.length > 1 && (
                <>
                  <span className="text-gray-400">→</span>
                  <span className="text-sm text-gray-500">(cycles)</span>
                </>
              )}
            </div>
          </div>

          <div>
            <h4 className="text-sm font-medium text-gray-700 mb-2">Long Press Cycle (≥ 1s)</h4>
            <div className="flex items-center gap-2 flex-wrap">
              {profile.buttonBehavior.longPressCycle.map((stateName, index) => (
                <span key={index} className="flex items-center gap-2">
                  <span className="px-3 py-1 bg-purple-100 text-purple-800 rounded-md text-sm">
                    {stateName}
                  </span>
                  {index < profile.buttonBehavior.longPressCycle.length - 1 && (
                    <span className="text-gray-400">→</span>
                  )}
                </span>
              ))}
              {profile.buttonBehavior.longPressCycle.length > 1 && (
                <>
                  <span className="text-gray-400">→</span>
                  <span className="text-sm text-gray-500">(cycles)</span>
                </>
              )}
            </div>
          </div>
        </div>
      </Card>

      {/* Device Sync Status */}
      {deviceStatus && (
        <Card title="Device Synchronization Status">
          <div className="space-y-4">
            {/* Summary */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="bg-green-50 p-3 rounded-lg">
                <p className="text-sm text-gray-600">Up to Date</p>
                <p className="text-2xl font-semibold text-green-700">
                  {deviceStatus.summary.up_to_date}
                </p>
              </div>
              <div className="bg-yellow-50 p-3 rounded-lg">
                <p className="text-sm text-gray-600">Update Pending</p>
                <p className="text-2xl font-semibold text-yellow-700">
                  {deviceStatus.summary.update_pending}
                </p>
              </div>
              <div className="bg-red-50 p-3 rounded-lg">
                <p className="text-sm text-gray-600">Failed</p>
                <p className="text-2xl font-semibold text-red-700">
                  {deviceStatus.summary.failed}
                </p>
              </div>
              <div className="bg-gray-50 p-3 rounded-lg">
                <p className="text-sm text-gray-600">Offline</p>
                <p className="text-2xl font-semibold text-gray-700">
                  {deviceStatus.summary.offline}
                </p>
              </div>
            </div>

            {/* Device List */}
            {deviceStatus.devices.length > 0 && (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Device
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Version
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Status
                      </th>
                      <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                        Last Check
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {deviceStatus.devices.map((device) => (
                      <tr key={device.device_mac}>
                        <td className="px-4 py-3 text-sm text-gray-900">{device.device_mac}</td>
                        <td className="px-4 py-3 text-sm text-gray-900">v{device.version}</td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${getStatusBadgeClass(device.status)}`}>
                            {device.status}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-sm text-gray-500">
                          {formatDistanceToNow(new Date(device.last_check), { addSuffix: true })}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Version History */}
      {versions && versions.length > 0 && (
        <Card title="Version History">
          <div className="space-y-3">
            {versions.map((version) => (
              <div key={version.id} className="flex items-start gap-3 pb-3 border-b last:border-0">
                <div className="flex-shrink-0 w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
                  <ClockIcon className="w-6 h-6 text-blue-600" />
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="font-semibold text-gray-900">Version {version.version}</span>
                    {version.version === profile.version && (
                      <span className="text-xs px-2 py-0.5 bg-green-100 text-green-800 rounded">
                        Current
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-gray-600 mt-1">
                    {formatDistanceToNow(new Date(version.created_at), { addSuffix: true })} by{' '}
                    {version.changed_by}
                  </p>
                  {version.changes && version.changes.length > 0 && (
                    <ul className="mt-2 text-sm text-gray-700 list-disc list-inside">
                      {version.changes.map((change, idx) => (
                        <li key={idx}>{change}</li>
                      ))}
                    </ul>
                  )}
                  {version.change_description && (
                    <p className="mt-1 text-sm text-gray-600 italic">
                      {version.change_description}
                    </p>
                  )}
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}
    </div>
  );
}
