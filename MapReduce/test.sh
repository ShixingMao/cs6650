# #!/bin/bash

# # Configuration
# BUCKET="your-mapreduce-bucket"
# SPLITTER_IP="54.149.55.122"
# MAPPER1_IP="35.88.150.191"
# MAPPER2_IP="44.252.58.75"
# MAPPER3_IP="52.11.217.181"
# REDUCER_IP="35.86.192.206"
# echo "Creating S3 bucket: $BUCKET"
# aws s3 mb s3://$BUCKET --region us-west-2 2>/dev/null || echo "Bucket may already exist"

# # Create and upload test file
# echo "Creating test file..."
# cat > input.txt << 'EOF'
# The quick brown fox jumps over the lazy dog. The dog was lazy.
# The brown fox was quick. A quick brown fox is faster than a lazy dog.
# The fox and dog became friends. The quick fox jumps high.
# The lazy dog sleeps. The brown fox runs fast. Quick brown foxes are smart.
# Dogs and foxes can be friends. The quick brown fox is very quick indeed.
# EOF

# aws s3 cp input.txt s3://$BUCKET/input.txt

# echo "1. Splitting file..."
# RESPONSE=$(curl -s "http://$SPLITTER_IP:8080/split?bucket=$BUCKET&file=input.txt")
# echo "Splitter response: $RESPONSE"

# # Parse chunks without jq (using grep and sed)
# CHUNKS=$(echo $RESPONSE | grep -o '"chunks":\[[^]]*\]' | sed 's/"chunks":\[//g' | sed 's/\]//g' | sed 's/"//g' | sed 's/,/ /g')

# echo "2. Running mappers..."
# RESULTS=""
# i=1
# for chunk in $CHUNKS; do
#     MAPPER_IP=$(eval echo \$MAPPER${i}_IP)
#     echo "   Mapper $i ($MAPPER_IP) processing chunk: $chunk"
#     RESULT=$(curl -s "http://$MAPPER_IP:8080/map?bucket=$BUCKET&chunk=$chunk")
#     echo "   Mapper response: $RESULT"
    
#     # Extract result file without jq
#     RESULT_FILE=$(echo $RESULT | grep -o '"result":"[^"]*"' | sed 's/"result":"//g' | sed 's/"//g')
#     RESULTS="$RESULTS \"$RESULT_FILE\""
#     i=$((i+1))
# done

# echo "3. Running reducer..."
# RESULTS_JSON=$(echo $RESULTS | sed 's/ /,/g')
# echo "   Sending to reducer: {\"bucket\":\"$BUCKET\",\"results\":[$RESULTS_JSON]}"

# FINAL_RESPONSE=$(curl -X POST "http://$REDUCER_IP:8080/reduce" \
#   -H "Content-Type: application/json" \
#   -d "{\"bucket\":\"$BUCKET\",\"results\":[$RESULTS_JSON]}")

# echo "Final response: $FINAL_RESPONSE"

# # Extract and display results
# echo ""
# echo "=== MapReduce Complete ==="
# echo "S3 Bucket: $BUCKET"
# echo "Check results with:"
# echo "  aws s3 ls s3://$BUCKET/final/ --recursive"
# echo "  aws s3 cp s3://$BUCKET/final/ . --recursive"

# # Cleanup option
# echo ""
# echo "To cleanup resources later:"
# echo "  aws s3 rm s3://$BUCKET --recursive"
# echo "  aws s3 rb s3://$BUCKET"

#!/bin/bash

# Configuration
BUCKET="your-mapreduce-bucket"
SPLITTER_IP="54.149.55.122"
MAPPER1_IP="35.88.150.191"
MAPPER2_IP="44.252.58.75"
MAPPER3_IP="52.11.217.181"
REDUCER_IP="35.86.192.206"

# Check if custom input file is provided as argument
if [ $# -eq 1 ]; then
    INPUT_FILE="$1"
    if [ ! -f "$INPUT_FILE" ]; then
        echo "Error: File '$INPUT_FILE' not found!"
        exit 1
    fi
    echo "Using custom input file: $INPUT_FILE"
else
    # Use default test file if no argument provided
    echo "No input file specified. Creating default test file..."
    echo "Usage: ./test.sh [your_input_file.txt]"
    echo ""
    
    INPUT_FILE="input.txt"
    cat > $INPUT_FILE << 'EOF'
The quick brown fox jumps over the lazy dog. The dog was lazy.
The brown fox was quick. A quick brown fox is faster than a lazy dog.
The fox and dog became friends. The quick fox jumps high.
The lazy dog sleeps. The brown fox runs fast. Quick brown foxes are smart.
Dogs and foxes can be friends. The quick brown fox is very quick indeed.
EOF
fi

echo "Creating S3 bucket: $BUCKET"
aws s3 mb s3://$BUCKET --region us-west-2 2>/dev/null || echo "Bucket may already exist"

# Upload the input file
echo "Uploading file to S3..."
aws s3 cp "$INPUT_FILE" s3://$BUCKET/input.txt

# Display file statistics
WORD_COUNT=$(wc -w < "$INPUT_FILE")
LINE_COUNT=$(wc -l < "$INPUT_FILE")
CHAR_COUNT=$(wc -c < "$INPUT_FILE")
echo "File statistics: $WORD_COUNT words, $LINE_COUNT lines, $CHAR_COUNT characters"

echo "1. Splitting file..."
SPLIT_START=$(date +%s%3N)
RESPONSE=$(curl -s "http://$SPLITTER_IP:8080/split?bucket=$BUCKET&file=input.txt")
SPLIT_END=$(date +%s%3N)
SPLIT_TIME=$((SPLIT_END - SPLIT_START))
echo "Splitter response: $RESPONSE"
echo "   Split time: ${SPLIT_TIME}ms"

# Parse chunks without jq (using grep and sed)
CHUNKS=$(echo $RESPONSE | grep -o '"chunks":\[[^]]*\]' | sed 's/"chunks":\[//g' | sed 's/\]//g' | sed 's/"//g' | sed 's/,/ /g')

echo ""
echo "=== PARALLEL PROCESSING (3 Mappers) ==="
PARALLEL_START=$(date +%s%3N)

# Run all mappers in parallel
i=1
PIDS=""
RESULTS=""
for chunk in $CHUNKS; do
    MAPPER_IP=$(eval echo \$MAPPER${i}_IP)
    echo "   Starting Mapper $i ($MAPPER_IP) for chunk: $chunk"
    
    # Run mapper in background and save PID
    (
        RESULT=$(curl -s "http://$MAPPER_IP:8080/map?bucket=$BUCKET&chunk=$chunk")
        echo "$RESULT" > /tmp/mapper_${i}_result.txt
    ) &
    PIDS="$PIDS $!"
    i=$((i+1))
done

# Wait for all mappers to complete
echo "   Waiting for all mappers to complete..."
for pid in $PIDS; do
    wait $pid
done

PARALLEL_END=$(date +%s%3N)
PARALLEL_MAP_TIME=$((PARALLEL_END - PARALLEL_START))

# Collect results
for j in 1 2 3; do
    RESULT=$(cat /tmp/mapper_${j}_result.txt)
    RESULT_FILE=$(echo $RESULT | grep -o '"result":"[^"]*"' | sed 's/"result":"//g' | sed 's/"//g')
    RESULTS="$RESULTS \"$RESULT_FILE\""
    rm /tmp/mapper_${j}_result.txt
done

echo "   Parallel mapping completed in: ${PARALLEL_MAP_TIME}ms"

echo ""
echo "=== SEQUENTIAL PROCESSING (1 Mapper) ==="
SEQUENTIAL_START=$(date +%s%3N)

# Run all chunks through one mapper sequentially
SEQ_RESULTS=""
chunk_num=1
for chunk in $CHUNKS; do
    echo "   Mapper 1 processing chunk $chunk_num: $chunk"
    RESULT=$(curl -s "http://$MAPPER1_IP:8080/map?bucket=$BUCKET&chunk=$chunk")
    RESULT_FILE=$(echo $RESULT | grep -o '"result":"[^"]*"' | sed 's/"result":"//g' | sed 's/"//g')
    SEQ_RESULTS="$SEQ_RESULTS \"seq_$RESULT_FILE\""
    chunk_num=$((chunk_num + 1))
done

SEQUENTIAL_END=$(date +%s%3N)
SEQUENTIAL_MAP_TIME=$((SEQUENTIAL_END - SEQUENTIAL_START))

echo "   Sequential mapping completed in: ${SEQUENTIAL_MAP_TIME}ms"

echo ""
echo "3. Running reducer..."
REDUCER_START=$(date +%s%3N)
RESULTS_JSON=$(echo $RESULTS | sed 's/ /,/g')
echo "   Sending to reducer: {\"bucket\":\"$BUCKET\",\"results\":[$RESULTS_JSON]}"

FINAL_RESPONSE=$(curl -X POST "http://$REDUCER_IP:8080/reduce" \
  -H "Content-Type: application/json" \
  -d "{\"bucket\":\"$BUCKET\",\"results\":[$RESULTS_JSON]}")

REDUCER_END=$(date +%s%3N)
REDUCER_TIME=$((REDUCER_END - REDUCER_START))

echo "Final response: $FINAL_RESPONSE"
echo "   Reducer time: ${REDUCER_TIME}ms"

# Calculate total times and speedup
TOTAL_PARALLEL=$((SPLIT_TIME + PARALLEL_MAP_TIME + REDUCER_TIME))
TOTAL_SEQUENTIAL=$((SPLIT_TIME + SEQUENTIAL_MAP_TIME + REDUCER_TIME))

# Calculate speedup with 2 decimal places
SPEEDUP_INT=$((SEQUENTIAL_MAP_TIME * 100 / PARALLEL_MAP_TIME))
SPEEDUP_WHOLE=$((SPEEDUP_INT / 100))
SPEEDUP_DECIMAL=$((SPEEDUP_INT % 100))
SPEEDUP="${SPEEDUP_WHOLE}.${SPEEDUP_DECIMAL}"

# Calculate efficiency (speedup / 3 * 100)
EFFICIENCY=$((SPEEDUP_INT * 100 / 3 / 100))

echo ""
echo "╔══════════════════════════════════════════════════════╗"
echo "║           PERFORMANCE COMPARISON RESULTS             ║"
echo "╠══════════════════════════════════════════════════════╣"
echo "║ PHASE               │ Time (ms)                      ║"
echo "╠─────────────────────┼────────────────────────────────╣"
echo "║ Splitting           │ $(printf "%-30s" "${SPLIT_TIME}ms") ║"
echo "║ Mapping (Parallel)  │ $(printf "%-30s" "${PARALLEL_MAP_TIME}ms (3 mappers)") ║"
echo "║ Mapping (Sequential)│ $(printf "%-30s" "${SEQUENTIAL_MAP_TIME}ms (1 mapper)") ║"
echo "║ Reducing            │ $(printf "%-30s" "${REDUCER_TIME}ms") ║"
echo "╠═════════════════════╪════════════════════════════════╣"
echo "║ TOTAL (Parallel)    │ $(printf "%-30s" "${TOTAL_PARALLEL}ms") ║"
echo "║ TOTAL (Sequential)  │ $(printf "%-30s" "${TOTAL_SEQUENTIAL}ms") ║"
echo "╠═════════════════════╪════════════════════════════════╣"
echo "║ Map Phase Speedup   │ $(printf "%-30s" "${SPEEDUP}x") ║"
echo "║ Parallel Efficiency │ $(printf "%-30s" "${EFFICIENCY}%") ║"
echo "╚═════════════════════╧════════════════════════════════╝"

echo ""
echo "📊 Visual Comparison (Map Phase):"
SEQ_BARS=$((SEQUENTIAL_MAP_TIME / 100))
PAR_BARS=$((PARALLEL_MAP_TIME / 100))

echo "Sequential: $(printf '█%.0s' $(seq 1 ${SEQ_BARS:-1})) ${SEQUENTIAL_MAP_TIME}ms"
echo "Parallel:   $(printf '█%.0s' $(seq 1 ${PAR_BARS:-1})) ${PARALLEL_MAP_TIME}ms"

echo ""
echo "=== MapReduce Complete ==="
echo "S3 Bucket: $BUCKET"
echo "Check results with:"
echo "  aws s3 ls s3://$BUCKET/final/ --recursive"
echo "  aws s3 cp s3://$BUCKET/final/ . --recursive"

# Cleanup option
echo ""
echo "To cleanup resources later:"
echo "  aws s3 rm s3://$BUCKET --recursive"
echo "  aws s3 rb s3://$BUCKET"