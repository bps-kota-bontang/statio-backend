#!/usr/bin/env node
/**
 * Deep Binary Structure Analysis for BIFF8 XLS files
 */

const fs = require('fs');

function analyzeCFBStructure(filePath) {
  const buf = fs.readFileSync(filePath);
  
  console.log(`\n📁 Analyzing: ${filePath}`);
  console.log('='.repeat(70));
  
  // CFB Header Analysis (first 512 bytes)
  console.log('\n🔍 Compound File Binary (CFB) Header:');
  
  // Magic number
  const magic = buf.slice(0, 8).toString('hex');
  console.log(`  Signature: ${magic} ${magic === 'd0cf11e0a1b11ae1' ? '✅' : '❌'}`);
  
  // CLSID
  const clsid = buf.slice(8, 24).toString('hex');
  console.log(`  CLSID: ${clsid}`);
  
  // Version
  const minorVersion = buf.readUInt16LE(24);
  const majorVersion = buf.readUInt16LE(26);
  console.log(`  Version: ${majorVersion}.${minorVersion}`);
  
  // Byte order
  const byteOrder = buf.readUInt16LE(28);
  console.log(`  Byte Order: 0x${byteOrder.toString(16)} ${byteOrder === 0xFFFE ? '✅ Little-endian' : '❌'}`);
  
  // Sector sizes
  const sectorShift = buf.readUInt16LE(30);
  const sectorSize = Math.pow(2, sectorShift);
  console.log(`  Sector Shift: ${sectorShift} → Sector Size: ${sectorSize} bytes`);
  
  const miniSectorShift = buf.readUInt16LE(32);
  const miniSectorSize = Math.pow(2, miniSectorShift);
  console.log(`  Mini Sector Shift: ${miniSectorShift} → Mini Sector Size: ${miniSectorSize} bytes`);
  
  // FAT info
  const totalSectors = buf.readUInt32LE(44);
  const fatSectors = buf.readUInt32LE(48);
  const firstDirSector = buf.readUInt32LE(48);
  const firstMiniFatSector = buf.readUInt32LE(60);
  const miniFatSectors = buf.readUInt32LE(64);
  const firstDifatSector = buf.readUInt32LE(68);
  const difatSectors = buf.readUInt32LE(72);
  
  console.log(`\n📊 FAT (File Allocation Table) Info:`);
  console.log(`  Total Sectors: ${totalSectors}`);
  console.log(`  FAT Sectors: ${fatSectors}`);
  console.log(`  First Directory Sector: ${firstDirSector}`);
  console.log(`  First Mini FAT Sector: ${firstMiniFatSector}`);
  console.log(`  Mini FAT Sectors: ${miniFatSectors}`);
  console.log(`  First DIFAT Sector: ${firstDifatSector}`);
  console.log(`  DIFAT Sectors: ${difatSectors}`);
  
  // DIFAT array (first 109 entries)
  console.log(`\n🔗 DIFAT Array (first 10 entries):`);
  for (let i = 0; i < 10 && i < 109; i++) {
    const offset = 76 + (i * 4);
    const entry = buf.readUInt32LE(offset);
    if (entry !== 0xFFFFFFFF) {
      console.log(`  [${i}]: ${entry}`);
    }
  }
  
  // Check for specific BIFF records
  console.log(`\n📋 BIFF Record Analysis:`);
  let foundBOF = false;
  let foundWorkbook = false;
  
  // Scan for BOF record (0x0809)
  for (let i = 512; i < Math.min(buf.length, 4096); i++) {
    const recordId = buf.readUInt16LE(i);
    if (recordId === 0x0809) {
      console.log(`  ✅ Found BOF record at offset: 0x${i.toString(16)}`);
      const recordLen = buf.readUInt16LE(i + 2);
      console.log(`     Record length: ${recordLen} bytes`);
      foundBOF = true;
      break;
    }
  }
  
  if (!foundBOF) {
    console.log(`  ⚠️  BOF record not found in first 4KB`);
  }
  
  // Summary
  console.log(`\n✅ Summary:`);
  console.log(`  File Size: ${buf.length} bytes`);
  console.log(`  Sector Size: ${sectorSize} bytes`);
  console.log(`  Expected File Size: ${512 + (totalSectors * sectorSize)} bytes`);
  console.log(`  Structure: ${buf.length >= 512 + (totalSectors * sectorSize) ? '✅ Complete' : '⚠️  May be truncated'}`);
}

// Main
const args = process.argv.slice(2);
if (args.length === 0) {
  console.log('Usage: node cfb-analysis.js <file1.xls> [file2.xls]');
  process.exit(1);
}

args.forEach(file => {
  analyzeCFBStructure(file);
});
