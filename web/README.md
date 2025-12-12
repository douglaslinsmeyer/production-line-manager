# Production Line Manager - Web UI

React TypeScript web application for managing production lines with real-time status monitoring and analytics.

## Features

- **Dashboard**: Real-time production line monitoring with auto-refresh every 5 seconds
- **Production Lines**: Full CRUD operations with status management and label system
- **Schedules Management**:
  - Weekly schedule builder (7-day templates)
  - Holiday management (manual + auto-import from OpenHolidays API)
  - Schedule exceptions (global and line-specific)
  - Effective schedule viewer
- **Analytics Dashboard**:
  - Aggregate metrics (uptime, MTTR, downtime)
  - Per-line analytics table
  - Status distribution charts (Recharts)
  - Daily KPI trends
  - Timeline of all status changes
- **Compliance Tracking**:
  - Overall compliance percentage
  - Per-line compliance with color coding
  - Daily compliance breakdown
  - Unplanned downtime analysis
- **Device Management**:
  - Device discovery list (ESP32-S3 IoT devices)
  - Flash identify (LED + buzzer for physical identification)
  - Device-to-line assignment
  - Online/offline status tracking
- **Real-time Updates**: TanStack Query auto-refetches with caching
- **Toast Notifications**: Elegant user feedback for all operations
- **Responsive Design**: Mobile-first with Tailwind CSS

For detailed feature documentation, see [docs/features.md](docs/features.md)

## Tech Stack

- **React 18** with TypeScript
- **Vite** - Fast build tool and dev server
- **TanStack Query** - Server state management with automatic caching
- **React Router v6** - Client-side routing
- **Axios** - HTTP client
- **Tailwind CSS** - Utility-first styling
- **Headless UI** - Accessible UI components
- **Heroicons** - Icon library
- **Recharts** - Charting library for analytics
- **React Hook Form** - Form state management
- **Zod** - Schema validation
- **react-hot-toast** - Toast notifications

## Quick Start

```bash
# Install dependencies
npm install

# Start dev server
npm run dev

# Build for production
npm run build
```

Access at http://localhost:5173

## Environment Variables

Create `.env`:

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_MQTT_WS_URL=ws://localhost:9001
```

## Documentation

- **[Feature Documentation](docs/features.md)**: Complete UI features guide
- **[API Documentation](../api/README.md)**: Backend API reference
- **[Architecture Overview](../docs/architecture.md)**: System architecture
- **[Getting Started](../docs/getting-started.md)**: Development setup

## License

Proprietary - All rights reserved
