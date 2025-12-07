# Production Line Manager - Web UI

React TypeScript web application for managing production lines with real-time status monitoring and analytics.

## Features

- **Dashboard**: Real-time production line monitoring with auto-refresh
- **CRUD Operations**: Create, edit, and delete production lines
- **Status Management**: Change line status with instant feedback
- **Analytics**:
  - Status history chart (Recharts area chart)
  - Status timeline with source tracking
  - Uptime metrics and MTTR calculations
  - Status distribution visualization
- **Real-time Updates**: TanStack Query auto-refetches data every 5 seconds
- **Toast Notifications**: Elegant user feedback for all operations
- **Responsive Design**: Mobile-first with Tailwind CSS

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

## License

Proprietary - All rights reserved
