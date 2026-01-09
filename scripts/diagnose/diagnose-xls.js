#!/usr/bin/env node
/**
 * XLS File Diagnostic Tool
 * Analyzes XLS files to identify format issues
 * 
 * Usage: node diagnose-xls.js <file.xls>
 */

const fs = require('fs');
const XLSX = require('xlsx');

function analyzeFile(filePath) {
  console.log('📊 Analyzing:', filePath);
  console.log('='.repeat(60));
  
  // Read file
  const buffer = fs.readFileSync(filePath);
  
  // Check magic bytes
  console.log('\n🔍 File Header Analysis:');
  const header = buffer.slice(0, 16);
  console.log('  First 16 bytes (hex):', header.toString('hex'));
  console.log('  First 16 bytes (ascii):', header.toString('ascii', 0, 8));
  
  // Detect format
  const magicBIFF8 = 'D0CF11E0A1B11AE1';
  const headerHex = header.toString('hex').toUpperCase();
  
  if (headerHex.startsWith(magicBIFF8.toUpperCase())) {
    console.log('  ✅ Format: BIFF8 (Binary Excel 97-2003)');
  } else if (buffer.toString('utf8', 0, 5) === '<?xml') {
    console.log('  ✅ Format: SpreadsheetML 2003 (XML)');
  } else if (headerHex.startsWith('504B0304')) {
    console.log('  ✅ Format: XLSX (ZIP-based)');
  } else {
    console.log('  ⚠️  Format: Unknown/Unrecognized');
  }
  
  // File size
  console.log('\n📏 File Information:');
  console.log('  File size:', buffer.length, 'bytes');
  
  // Try to parse with SheetJS
  console.log('\n🔧 SheetJS Parse Attempt:');
  try {
    const workbook = XLSX.read(buffer, { 
      type: 'buffer',
      cellStyles: true,
      cellDates: true 
    });
    
    console.log('  ✅ Parsed successfully');
    console.log('  Sheet count:', workbook.SheetNames.length);
    console.log('  Sheet names:', workbook.SheetNames.join(', '));
    
    // Analyze first sheet
    if (workbook.SheetNames.length > 0) {
      const firstSheet = workbook.Sheets[workbook.SheetNames[0]];
      const range = XLSX.utils.decode_range(firstSheet['!ref'] || 'A1');
      
      console.log('\n📋 First Sheet Details:');
      console.log('  Name:', workbook.SheetNames[0]);
      console.log('  Range:', firstSheet['!ref']);
      console.log('  Rows:', range.e.r - range.s.r + 1);
      console.log('  Cols:', range.e.c - range.s.c + 1);
      
      // Sample first few cells
      console.log('\n📝 Sample Data (first 5 cells):');
      let cellCount = 0;
      for (let R = range.s.r; R <= Math.min(range.s.r + 2, range.e.r); R++) {
        for (let C = range.s.c; C <= Math.min(range.s.c + 2, range.e.c); C++) {
          const cellAddress = XLSX.utils.encode_cell({ r: R, c: C });
          const cell = firstSheet[cellAddress];
          if (cell) {
            console.log(`    ${cellAddress}: ${cell.v} (type: ${cell.t})`);
            cellCount++;
            if (cellCount >= 5) break;
          }
        }
        if (cellCount >= 5) break;
      }
    }
    
    // Check for special features
    console.log('\n🎨 Features Detected:');
    const props = workbook.Props || {};
    console.log('  Workbook Props:', Object.keys(props).length > 0 ? 'Yes' : 'No');
    console.log('  SST (Shared Strings):', workbook.SST ? 'Yes' : 'No');
    
  } catch (error) {
    console.log('  ❌ Parse failed:', error.message);
  }
  
  // Check for common issues
  console.log('\n⚠️  Potential Issues:');
  
  if (buffer.length < 512) {
    console.log('  • File too small (< 512 bytes) - may be corrupted');
  }
  
  if (buffer.length > 10 * 1024 * 1024) {
    console.log('  • File very large (> 10MB) - may cause upload issues');
  }
  
  // Check for null bytes
  const nullCount = buffer.filter(b => b === 0).length;
  const nullPercent = (nullCount / buffer.length * 100).toFixed(2);
  console.log(`  • Null bytes: ${nullCount} (${nullPercent}%)`);
  
  console.log('\n' + '='.repeat(60));
}

// Main
const args = process.argv.slice(2);
if (args.length === 0) {
  console.log('Usage: node diagnose-xls.js <file1.xls> [file2.xls]');
  console.log('');
  console.log('Examples:');
  console.log('  node diagnose-xls.js exported.xls');
  console.log('  node diagnose-xls.js before.xls after.xls');
  process.exit(1);
}

args.forEach((file, index) => {
  if (index > 0) console.log('\n\n');
  analyzeFile(file);
});

// Compare if 2 files provided
if (args.length === 2) {
  console.log('\n\n🔬 COMPARISON');
  console.log('='.repeat(60));
  
  const buf1 = fs.readFileSync(args[0]);
  const buf2 = fs.readFileSync(args[1]);
  
  console.log('\nFile Sizes:');
  console.log('  Before:', buf1.length, 'bytes');
  console.log('  After: ', buf2.length, 'bytes');
  console.log('  Diff:  ', buf2.length - buf1.length, 'bytes');
  
  // Header comparison
  const header1 = buf1.slice(0, 16).toString('hex');
  const header2 = buf2.slice(0, 16).toString('hex');
  
  console.log('\nHeader Match:', header1 === header2 ? '✅ Same' : '❌ Different');
  
  if (header1 !== header2) {
    console.log('  Before:', header1);
    console.log('  After: ', header2);
  }
  
  // Binary comparison
  let diffCount = 0;
  const minLen = Math.min(buf1.length, buf2.length);
  
  for (let i = 0; i < minLen; i++) {
    if (buf1[i] !== buf2[i]) {
      diffCount++;
    }
  }
  
  const diffPercent = (diffCount / minLen * 100).toFixed(2);
  console.log('\nBinary Difference:');
  console.log('  Diff bytes:', diffCount);
  console.log('  Diff %:', diffPercent + '%');
}
