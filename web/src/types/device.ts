export type DeviceStatus = 'online' | 'offline' | 'unknown';

export interface DeviceCapabilities {
  digital_inputs: number;
  digital_outputs: number;
  ethernet: boolean;
  wifi: boolean;
}

export interface DeviceMetadata {
  uptime_seconds?: number;
  free_heap?: number;
  rssi?: number | null;
}

export interface DiscoveredDevice {
  id: string;
  mac_address: string;
  device_type: string;
  firmware_version?: string;
  ip_address?: string;
  capabilities?: DeviceCapabilities;
  first_seen: string;
  last_seen: string;
  status: DeviceStatus;
  metadata?: DeviceMetadata;
  created_at: string;
  updated_at: string;
}

export interface DeviceWithAssignment extends DiscoveredDevice {
  assigned_line_id?: string;
  assigned_line_code?: string;
  assigned_line_name?: string;
  assigned_at?: string;
}

export interface DeviceLineAssignment {
  id: string;
  device_mac: string;
  line_id: string;
  assigned_at: string;
  assigned_by?: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface AssignDeviceRequest {
  line_id: string;
  assigned_by?: string;
}

export interface IdentifyDeviceRequest {
  duration?: number;
}

export interface DeviceCommand {
  command: string;
  channel?: number;
  state?: boolean;
  duration?: number;
  params?: Record<string, unknown>;
}

export interface LineDeviceResponse {
  assigned: boolean;
  device?: DiscoveredDevice;
  assignment?: DeviceLineAssignment;
}
