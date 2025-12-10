import axios from 'axios';
import type {
  DeviceWithAssignment,
  DiscoveredDevice,
  AssignDeviceRequest,
  IdentifyDeviceRequest,
  DeviceCommand,
  LineDeviceResponse,
} from '../types/device';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// List all discovered devices
export async function listDevices(): Promise<DeviceWithAssignment[]> {
  const response = await axios.get(`${API_URL}/devices`);
  return response.data;
}

// Get a specific device by MAC address
export async function getDevice(macAddress: string): Promise<DiscoveredDevice> {
  const response = await axios.get(`${API_URL}/devices/${macAddress}`);
  return response.data;
}

// Assign a device to a production line
export async function assignDevice(
  macAddress: string,
  request: AssignDeviceRequest
): Promise<void> {
  await axios.post(`${API_URL}/devices/${macAddress}/assign`, request);
}

// Unassign a device from its production line
export async function unassignDevice(macAddress: string): Promise<void> {
  await axios.delete(`${API_URL}/devices/${macAddress}/assign`);
}

// Flash device for physical identification (LED + buzzer)
export async function identifyDevice(
  macAddress: string,
  request?: IdentifyDeviceRequest
): Promise<void> {
  await axios.post(`${API_URL}/devices/${macAddress}/identify`, request || { duration: 10 });
}

// Send custom command to device
export async function sendDeviceCommand(
  macAddress: string,
  command: DeviceCommand
): Promise<void> {
  await axios.post(`${API_URL}/devices/${macAddress}/command`, command);
}

// Get device assigned to a specific line
export async function getLineDevice(lineId: string): Promise<LineDeviceResponse> {
  const response = await axios.get(`${API_URL}/lines/${lineId}/device`);
  return response.data;
}
