#!/usr/bin/env node
/**
 * XLSX to XLS Converter Service
 * Uses SheetJS (xlsx package) to convert XLSX binary to XLS binary
 *
 * Communication: stdin/stdout binary pipes
 * Input: XLSX file as binary data via stdin
 * Output: XLS file as binary data via stdout
 * Errors: Logged to stderr
 */

const XLSX = require("xlsx");
const CFB = require("cfb");

/**
 * Read binary data from stdin
 */
async function readStdin() {
  return new Promise((resolve, reject) => {
    const chunks = [];

    process.stdin.on("data", (chunk) => {
      chunks.push(chunk);
    });

    process.stdin.on("end", () => {
      resolve(Buffer.concat(chunks));
    });

    process.stdin.on("error", (err) => {
      reject(err);
    });
  });
}

/**
 * Convert XLSX buffer to XLS buffer
 */
function convertXLSXtoXLS(xlsxBuffer) {
  try {
    // Parse XLSX from buffer
    const workbook = XLSX.read(xlsxBuffer, {
      type: "buffer",
      cellDates: false, // Keep as numbers
      cellNF: false,
      cellStyles: true, // Preserve styles
      sheetStubs: false, // Don't create stubs for empty cells
    });

    // Clean up empty cells to reduce file size
    workbook.SheetNames.forEach((sheetName) => {
      const sheet = workbook.Sheets[sheetName];
      if (!sheet || !sheet["!ref"]) return;

      // Remove undefined/empty cells
      Object.keys(sheet).forEach((key) => {
        if (key[0] === "!") return; // Skip special keys
        const cell = sheet[key];
        if (!cell || cell.v === undefined || cell.v === null || cell.v === "") {
          delete sheet[key];
        }
      });
    });

    // Write as XLS (BIFF8 format - Excel 97-2003 binary)
    // Must match template.xls format (CFB binary structure)
    const xlsBuffer = XLSX.write(workbook, {
      type: "buffer",
      bookType: "biff8", // Binary Excel 97-2003 format
      cellDates: false,
      bookSST: true, // Use shared string table for Excel compatibility
    });

    // CRITICAL FIX: Rebuild CFB container using cfb library
    // SheetJS generates incompatible CFB structure, so we extract BIFF data
    // and rewrap it using CFB library which creates Excel-compatible structure
    try {
      // Parse existing CFB to extract BIFF workbook stream
      const cfbData = CFB.read(xlsBuffer, { type: "buffer" });

      // Find the Workbook stream (contains BIFF8 data)
      const workbookEntry =
        CFB.find(cfbData, "Workbook") || CFB.find(cfbData, "Book");

      if (workbookEntry) {
        // Create new CFB container with proper structure
        const newCFB = CFB.utils.cfb_new();

        // Add workbook stream with BIFF data
        CFB.utils.cfb_add(newCFB, "Workbook", workbookEntry.content);

        // Copy other streams if they exist
        cfbData.FileIndex.forEach((entry) => {
          if (
            entry.name &&
            entry.name !== "Workbook" &&
            entry.name !== "Book" &&
            entry.type === 2 &&
            entry.content
          ) {
            CFB.utils.cfb_add(newCFB, entry.name, entry.content);
          }
        });

        // Write CFB with Excel-compatible structure
        return CFB.write(newCFB, { type: "buffer" });
      }
    } catch (rebuildError) {
      // If CFB rebuild fails, return original buffer
      console.error("CFB rebuild warning:", rebuildError.message);
    }

    return xlsBuffer;
  } catch (error) {
    throw new Error(`Conversion failed: ${error.message}`);
  }
}

/**
 * Main execution
 */
async function main() {
  try {
    // Read XLSX data from stdin
    const xlsxBuffer = await readStdin();

    if (xlsxBuffer.length === 0) {
      throw new Error("No input data received");
    }

    // Convert XLSX to XLS
    const xlsBuffer = convertXLSXtoXLS(xlsxBuffer);

    // Write XLS to stdout
    process.stdout.write(xlsBuffer);

    // Exit successfully
    process.exit(0);
  } catch (error) {
    // Log error to stderr (won't interfere with stdout binary data)
    console.error(`[xlsx-to-xls] Error: ${error.message}`);
    process.exit(1);
  }
}

// Execute
main();
