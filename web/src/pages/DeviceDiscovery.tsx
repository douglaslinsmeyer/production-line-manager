import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { listDevices, identifyDevice, assignDevice, unassignDevice, forgetDevice } from '../api/devices';
import { linesApi } from '../api/lines';
import type { DeviceWithAssignment } from '../types/device';
import type { ProductionLine } from '../api/types';

export default function DeviceDiscovery() {
  const queryClient = useQueryClient();
  const [flashingDevices, setFlashingDevices] = useState<Set<string>>(new Set());

  // Fetch devices
  const { data: devices, isLoading, error } = useQuery({
    queryKey: ['devices'],
    queryFn: listDevices,
    refetchInterval: 5000, // Poll every 5 seconds for real-time updates
  });

  // Fetch production lines for assignment dropdown
  const { data: lines } = useQuery({
    queryKey: ['lines'],
    queryFn: linesApi.getLines,
  });

  // Flash identify mutation
  const flashMutation = useMutation({
    mutationFn: (macAddress: string) => identifyDevice(macAddress, { duration: 10 }),
    onSuccess: (_, macAddress) => {
      // Add to flashing set
      setFlashingDevices(prev => new Set(prev).add(macAddress));

      // Remove after 10 seconds
      setTimeout(() => {
        setFlashingDevices(prev => {
          const next = new Set(prev);
          next.delete(macAddress);
          return next;
        });
      }, 10000);
    },
  });

  // Assign device mutation
  const assignMutation = useMutation({
    mutationFn: ({ macAddress, lineId }: { macAddress: string; lineId: string }) =>
      assignDevice(macAddress, { line_id: lineId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });

  // Unassign device mutation
  const unassignMutation = useMutation({
    mutationFn: (macAddress: string) => unassignDevice(macAddress),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });

  // Delete/forget device mutation
  const forgetMutation = useMutation({
    mutationFn: (macAddress: string) => forgetDevice(macAddress),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });

  const handleFlash = (device: DeviceWithAssignment) => {
    flashMutation.mutate(device.mac_address);
  };

  const handleAssign = (device: DeviceWithAssignment, lineId: string) => {
    assignMutation.mutate({ macAddress: device.mac_address, lineId });
  };

  const handleUnassign = (device: DeviceWithAssignment) => {
    if (confirm(`Unassign device ${device.mac_address}?`)) {
      unassignMutation.mutate(device.mac_address);
    }
  };

  const handleForget = (device: DeviceWithAssignment) => {
    if (confirm(`Are you sure you want to forget device ${device.mac_address}? This will permanently remove it from the system.`)) {
      forgetMutation.mutate(device.mac_address);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'online':
        return 'text-green-600';
      case 'offline':
        return 'text-gray-400';
      default:
        return 'text-yellow-600';
    }
  };

  const getStatusIcon = (status: string) => {
    return status === 'online' ? '●' : '○';
  };

  const formatTimeSince = (timestamp: string) => {
    const seconds = Math.floor((Date.now() - new Date(timestamp).getTime()) / 1000);
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    return `${Math.floor(seconds / 3600)}h ago`;
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="text-gray-600">Loading devices...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="text-red-600">Error loading devices</div>
          <div className="text-sm text-gray-500 mt-2">{(error as Error).message}</div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Discovered Devices</h1>
        <p className="text-gray-600 mt-2">
          Manage ESP32-S3 devices and assign them to production lines
        </p>
      </div>

      {devices && devices.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-8 text-center">
          <div className="text-gray-500 text-lg">No devices discovered yet</div>
          <div className="text-sm text-gray-400 mt-2">
            Power on ESP32-S3 devices - they will auto-discover via MQTT
          </div>
        </div>
      ) : (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Device
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  IP Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Assigned To
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Capabilities
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {devices?.map((device) => (
                <tr key={device.mac_address} className="hover:bg-gray-50">
                  {/* Device Info */}
                  <td className="px-6 py-4">
                    <div className="text-sm font-medium text-gray-900">
                      {device.mac_address}
                    </div>
                    <div className="text-sm text-gray-500">
                      {device.device_type}
                    </div>
                    {device.firmware_version && (
                      <div className="text-xs text-gray-400">
                        v{device.firmware_version}
                      </div>
                    )}
                  </td>

                  {/* IP Address */}
                  <td className="px-6 py-4 text-sm text-gray-900">
                    {device.ip_address || '-'}
                  </td>

                  {/* Status */}
                  <td className="px-6 py-4">
                    <div className={`text-sm font-medium ${getStatusColor(device.status)}`}>
                      <span className="mr-1">{getStatusIcon(device.status)}</span>
                      {device.status.charAt(0).toUpperCase() + device.status.slice(1)}
                    </div>
                    <div className="text-xs text-gray-500">
                      {formatTimeSince(device.last_seen)}
                    </div>
                  </td>

                  {/* Assignment */}
                  <td className="px-6 py-4">
                    {device.assigned_line_id ? (
                      <div>
                        <div className="text-sm font-medium text-gray-900">
                          {device.assigned_line_code}
                        </div>
                        <div className="text-xs text-gray-500">
                          {device.assigned_line_name}
                        </div>
                        <button
                          onClick={() => handleUnassign(device)}
                          className="text-xs text-red-600 hover:text-red-800 mt-1"
                          disabled={unassignMutation.isPending}
                        >
                          Unassign
                        </button>
                      </div>
                    ) : (
                      <select
                        className="text-sm border-gray-300 rounded-md"
                        onChange={(e) => e.target.value && handleAssign(device, e.target.value)}
                        disabled={assignMutation.isPending}
                        defaultValue=""
                      >
                        <option value="">Select line...</option>
                        {lines?.map((line: ProductionLine) => (
                          <option key={line.id} value={line.id}>
                            {line.code} - {line.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </td>

                  {/* Capabilities */}
                  <td className="px-6 py-4">
                    {device.capabilities && (
                      <div className="text-xs text-gray-600">
                        <div>{device.capabilities.digital_inputs}DI / {device.capabilities.digital_outputs}DO</div>
                        <div className="text-gray-400">
                          {device.capabilities.ethernet && 'ETH'}
                          {device.capabilities.wifi && ' WiFi'}
                        </div>
                      </div>
                    )}
                  </td>

                  {/* Actions */}
                  <td className="px-6 py-4 text-right text-sm font-medium">
                    <div className="flex justify-end gap-2">
                      <button
                        onClick={() => handleFlash(device)}
                        disabled={flashMutation.isPending || flashingDevices.has(device.mac_address)}
                        className={`px-4 py-2 rounded-md text-white transition-colors ${
                          flashingDevices.has(device.mac_address)
                            ? 'bg-green-500 animate-pulse'
                            : 'bg-blue-600 hover:bg-blue-700'
                        } disabled:opacity-50`}
                      >
                        {flashingDevices.has(device.mac_address) ? 'Flashing...' : 'Flash'}
                      </button>
                      <button
                        onClick={() => handleForget(device)}
                        disabled={forgetMutation.isPending}
                        className="px-4 py-2 rounded-md bg-red-600 text-white hover:bg-red-700 transition-colors disabled:opacity-50"
                        title="Forget this device"
                      >
                        Forget
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Summary Info */}
      {devices && devices.length > 0 && (
        <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">Total Devices</div>
            <div className="text-2xl font-bold text-gray-900">{devices.length}</div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">Online</div>
            <div className="text-2xl font-bold text-green-600">
              {devices.filter(d => d.status === 'online').length}
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-4">
            <div className="text-sm text-gray-500">Assigned</div>
            <div className="text-2xl font-bold text-blue-600">
              {devices.filter(d => d.assigned_line_id).length}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
