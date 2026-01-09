#!/bin/bash
# Test different XLS export formats

set -e

echo "🧪 Testing Different XLS Format Options"
echo "========================================"
echo ""

# Create test XLSX
cat > test-input.xlsx.json << 'EOF'
{
  "SheetNames": ["Test"],
  "Sheets": {
    "Test": {
      "A1": {"t":"s", "v":"Name"},
      "B1": {"t":"s", "v":"Value"},
      "A2": {"t":"s", "v":"Test1"},
      "B2": {"t":"n", "v":123},
      "A3": {"t":"s", "v":"Test2"},
      "B3": {"t":"n", "v":456},
      "!ref": "A1:B3"
    }
  }
}
EOF

# Generate test XLSX using Node.js
node << 'NODESCRIPT'
const XLSX = require('xlsx');
const fs = require('fs');

const data = require('./test-input.xlsx.json');
const wb = {
  SheetNames: data.SheetNames,
  Sheets: data.Sheets
};

const xlsxBuffer = XLSX.write(wb, { type: 'buffer', bookType: 'xlsx' });
fs.writeFileSync('test-input.xlsx', xlsxBuffer);
console.log('✅ Created test-input.xlsx');
NODESCRIPT

echo ""
echo "📝 Testing Format Options:"
echo ""

# Test 1: Current BIFF8
echo "1️⃣ Testing BIFF8 (current)..."
cat test-input.xlsx | node ./scripts/xlsx-to-xls.js > test-biff8.xls 2>/dev/null || echo "Failed"
ls -lh test-biff8.xls 2>/dev/null && echo "✅ Generated test-biff8.xls"

# Test 2: XLML
echo "2️⃣ Testing XLML (XML-based)..."
cat test-input.xlsx | node ./scripts/xlsx-to-xls-xlml.js > test-xlml.xls 2>/dev/null || echo "Failed"
ls -lh test-xlml.xls 2>/dev/null && echo "✅ Generated test-xlml.xls"

# Test 3: Alternative BIFF8 options
echo "3️⃣ Testing BIFF8 (no SST)..."
cat test-input.xlsx | node -e "
const XLSX = require('xlsx');
const fs = require('fs');
const chunks = [];
process.stdin.on('data', c => chunks.push(c));
process.stdin.on('end', () => {
  const buf = Buffer.concat(chunks);
  const wb = XLSX.read(buf, {type:'buffer'});
  const out = XLSX.write(wb, {type:'buffer', bookType:'biff8', bookSST:false});
  process.stdout.write(out);
});
" > test-biff8-nosst.xls 2>/dev/null || echo "Failed"
ls -lh test-biff8-nosst.xls 2>/dev/null && echo "✅ Generated test-biff8-nosst.xls"

echo ""
echo "📊 File Analysis:"
echo ""

for file in test-biff8.xls test-xlml.xls test-biff8-nosst.xls; do
  if [ -f "$file" ]; then
    echo "File: $file"
    echo "  Size: $(ls -lh $file | awk '{print $5}')"
    echo "  Type: $(file -b $file | head -c 50)"
    echo "  Magic: $(xxd -l 8 -p $file)"
    echo ""
  fi
done

echo "🎯 Recommendation:"
echo "  Upload each file to your system and see which one works"
echo "  Then update the converter script accordingly"
echo ""
echo "🧹 Cleanup: rm -f test-*.xls test-input.xlsx*"
