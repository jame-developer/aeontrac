# AeonTrac - Time Tracking System

AeonTrac is a sophisticated command-line time tracking application written in Go, designed to help users track their working hours, manage overtime, and handle various types of time entries including regular work and compensatory time.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/aeontrac.git

# Change to the project directory
cd aeontrac

# Build the application
go build -o aeontrac cmd/cli/main.go

# (Optional) Build the REST API server
go build -o aeontrac-api cmd/api/main.go

# Run the application
./aeontrac

# Run the REST API server
./aeontrac-api
```

## Features

### Core Time Tracking
- Start/stop time tracking for work units
- Add time entries retroactively
- Support for work and compensatory time types
- Automatic duration calculation
- Comments support for time entries

### Smart Time Management
- Automatic overtime calculation
- Public holiday integration via OpenHolidaysAPI
- Weekend detection
- Vacation day tracking
- ISO week number tracking

### Reporting
- Daily work summaries
- Quarterly reports showing:
  - Weekly total hours
  - Weekly overtime hours
- Public holiday forecasting for upcoming days
- Duration formatting in HH:MM:SS
- Standalone quarterly report tool (`cmd/quartly.go`) for additional reporting options

## REST API

AeonTrac provides a REST API for programmatic access to time tracking features. This allows integration with other tools and automation of time tracking workflows.

### Running the API Server

To build and run the REST API server:

```bash
go build -o aeontrac-api cmd/api/main.go
./aeontrac-api
```

By default, the server listens on port 8080. You can interact with the API using `curl`, Postman, or any HTTP client.

### API Endpoints

#### 1. `/start`

- **Method:** `POST`
- **Description:** Start a new time tracking session.
- **Request Body:**
  ```json
  {
    "type": "WORK",         // "WORK" or "COMPENSATORY"
    "comment": "Optional comment for this session"
  }
  ```
- **Example Request:**
  ```bash
  curl -X POST http://localhost:8080/start \
    -H "Content-Type: application/json" \
    -d '{"type":"WORK","comment":"Project development"}'
  ```
- **Example Response:**
  ```json
  {
    "status": "started",
    "unit_id": "550e8400-e29b-41d4-a716-446655440000",
    "start": "2025-06-20T09:00:00Z"
  }
  ```

#### 2. `/stop`

- **Method:** `POST`
- **Description:** Stop the currently running time tracking session.
- **Request Body:** _(optional)_ You may include a comment.
  ```json
  {
    "comment": "Finished for today"
  }
  ```
- **Example Request:**
  ```bash
  curl -X POST http://localhost:8080/stop \
    -H "Content-Type: application/json" \
    -d '{"comment":"Finished for today"}'
  ```
- **Example Response:**
  ```json
  {
    "status": "stopped",
    "unit_id": "550e8400-e29b-41d4-a716-446655440000",
    "stop": "2025-06-20T17:00:00Z",
    "duration": "08:00:00"
  }
  ```

#### 3. `/report`

- **Method:** `GET`
- **Description:** Retrieve a summary report of tracked time.
- **Request Body:** _None_
- **Example Request:**
  ```bash
  curl http://localhost:8080/report
  ```
- **Example Response:**
  ```json
  {
    "total_hours": "40:00:00",
    "overtime_hours": "05:00:00",
    "days": [
      {
        "date": "2025-06-20",
        "total_hours": "08:00:00",
        "overtime_hours": "01:00:00"
      }
      // ... more days
    ]
  }
  ```

## Data Model

### AeonVault
The main data structure that stores all tracking information:

```json
{
  "aeon_days": {
    "2025-06-20": {
      // AeonDay structure
    }
  },
  "current_running_unit": {
    "DayKey": "2025-06-20",
    "UnitID": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

Fields:
- `aeon_days`: Map of daily tracking entries, keyed by date in YYYY-MM-DD format
- `current_running_unit`: Information about the currently active tracking session (optional)
  - `DayKey`: The date of the running unit
  - `UnitID`: Unique identifier for the running unit

### AeonDay
Represents a single day of tracking:

```json
{
  "iso_week_number": 25,
  "iso_week_day": 5,
  "public_holiday": false,
  "public_holiday_name": "",
  "vacation_day": false,
  "week_end": false,
  "total_hours": "08:00:00",
  "overtime_hours": "00:00:00",
  "units": {
    "550e8400-e29b-41d4-a716-446655440000": {
      // AeonUnit structure
    }
  }
}
```

Fields:
- `iso_week_number`: Week number according to ISO 8601 (1-53)
- `iso_week_day`: Day of week according to ISO 8601 (1-7)
- `public_holiday`: Boolean indicating if the day is a public holiday
- `public_holiday_name`: Name of the holiday if applicable
- `vacation_day`: Boolean indicating if the day is a vacation day
- `week_end`: Boolean indicating if the day is a weekend
- `total_hours`: Total working hours for the day (HH:MM:SS format)
- `overtime_hours`: Overtime hours for the day (HH:MM:SS format)
- `units`: Map of time tracking units, keyed by UUID

### AeonUnit
Individual time tracking entry:

```json
{
  "start": "2025-06-20T09:00:00Z",
  "stop": "2025-06-20T17:00:00Z",
  "duration": "08:00:00",
  "type": "WORK",
  "comment": "Project development"
}
```

Fields:
- `start`: Start time of the work unit (RFC3339 format)
- `stop`: End time of the work unit (RFC3339 format)
- `duration`: Duration of the work unit (HH:MM:SS format)
- `type`: Type of the unit ("WORK" or "COMPENSATORY")
- `comment`: Optional comment for the work unit

## Configuration

The application supports configuration for:
- Working hours
  - Default working day duration
  - Overtime calculation rules
- Public holidays
  - Country-specific holidays via OpenHolidaysAPI
  - Automatic holiday name and date detection

## Commands

- `start [time] [comment]` - Start tracking a new work unit
- `stop [time] [comment]` - Stop the current work unit
- `add [startTime] [stopTime]` - Add a work unit retroactively
- `qrep` - Generate quarterly report

Common flags:
- `-c, --comment` - Add a comment to the time entry

## Storage

The application follows XDG Base Directory Specification:
- Configuration: `$XDG_CONFIG_HOME/aeontrac` or `~/.config/aeontrac`
- Data: `$XDG_DATA_HOME/aeontrac` or `~/.local/share/aeontrac`

Data is stored in JSON format with automatic backup support.

## Additional Tools

### Standalone Quarterly Report Tool
Located in `cmd/quartly.go`, this tool provides an alternative way to generate quarterly reports:
- Processes the last 3 months of time tracking data
- Aggregates hours by ISO week number
- Shows both total hours and overtime hours per week
- Can be run independently of the main application

## Validation

- Comprehensive data validation using go-playground/validator
- Time entry overlap prevention
- Future date prevention
- Required field validation

## Error Handling

The application includes robust error handling for:
- Invalid time entries
- Overlapping work units
- Configuration errors
- API communication issues
- File operations

## Dependencies

- github.com/spf13/cobra - Command line interface
- github.com/go-playground/validator - Data validation
- github.com/google/uuid - Unique identifier generation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.