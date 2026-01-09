# XLSX to XLS Converter - Node.js Service

This directory contains a Node.js service for converting XLSX files to XLS format using SheetJS.

## Overview

The Go backend generates XLSX files (with full feature support including merge cells), then communicates with this Node.js script via stdin/stdout pipes to convert XLSX to legacy XLS format (BIFF8).

## Architecture

```
┌──────────────┐         ┌─────────────────┐         ┌──────────────┐
│   Go Backend │         │  IPC (pipes)    │         │   Node.js    │
│              │         │                 │         │  + SheetJS   │
│  ExportXLSX()├────────►│  stdin/stdout   ├────────►│              │
│   []byte     │  XLSX   │  binary data    │  XLS    │  convert()   │
│              │◄────────┤                 │◄────────┤              │
└──────────────┘         └─────────────────┘         └──────────────┘
```

## Installation

```bash
cd statio-backend/scripts
npm install
```

This will install:
- `xlsx` (SheetJS) - For XLSX/XLS conversion

## Usage

### Standalone
```bash
# Convert file
cat input.xlsx | node xlsx-to-xls.js > output.xls

# Test
echo "test" | node xlsx-to-xls.js
```

### From Go
The Go service automatically calls this script via `XLSXConverterService`:

```go
converter, _ := NewXLSXConverterService()
xlsData, err := converter.ConvertXLSXToXLS(xlsxBytes)
```

## Features

- ✅ Pure binary communication (no JSON overhead)
- ✅ Preserves all Excel features (merge cells, styles, formulas)
- ✅ Fast conversion using SheetJS
- ✅ No HTTP server required
- ✅ Error handling via stderr
- ✅ Production-ready

## File Structure

```
scripts/
├── package.json          # Node.js dependencies
├── xlsx-to-xls.js        # Converter script
└── README.md            # This file
```

## Environment Variables

No environment variables required. The script reads from stdin and writes to stdout.

## Error Handling

- Errors are logged to stderr (won't interfere with stdout binary data)
- Exit code 0 = success
- Exit code 1 = error

## Performance

- Typical conversion: < 100ms for small files (< 1MB)
- Memory efficient: streams data through pipes
- No disk I/O required

## Requirements

- Node.js >= 14.0.0
- npm (for installation)
- SheetJS (xlsx package)

## Troubleshooting

### "node: command not found"
Install Node.js from https://nodejs.org/

### "Cannot find module 'xlsx'"
Run `npm install` in the scripts directory

### Conversion fails
Check stderr output for detailed error messages

## License

MIT License - Same as main project
